package handler

import (
	"database/sql"
	"fmt"
	"github.com/dop251/goja"
	"github.com/fasthttp/router"
	"github.com/robfig/cron/v3"
	_ "github.com/russross/blackfriday/v2"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"io/ioutil"
	"matrix-github-bot/config"
	"matrix-github-bot/types"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/format"
	"maunium.net/go/mautrix/id"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type Cron struct {
	cronTime int64
	lastCron int64
	callable *goja.Callable
}

type Script struct {
	name  string
	data  string
	vm    *goja.Runtime
	crons []*Cron
}

type PluginHandler struct {
	DB            *sql.DB
	PluginDir     string
	Logger        *logrus.Logger
	Config        *config.Config
	Client        *mautrix.Client
	netListener   net.Listener
	loadedPlugins []*Script
	currentPlugin *Script
	router        *router.Router
	routeCount    int16
	httpClient    *http.Client
	cronContext   *cron.Cron
	mutex         *sync.Mutex
}

func (pm *PluginHandler) Index(ctx *fasthttp.RequestCtx) {
	_, _ = ctx.WriteString("By mc8051.de")
}

func (pm *PluginHandler) IsPluginEnabled(e string) bool {
	for _, a := range pm.Config.Plugins {
		if a == e {
			return true
		}
	}
	return false
}

func (pm *PluginHandler) Init() error {
	pm.router = router.New()
	pm.router.GET("/", pm.Index)

	pm.httpClient = &http.Client{}
	pm.mutex = &sync.Mutex{}
	pm.cronContext = cron.New()

	err := filepath.Walk(pm.PluginDir, func(path string, info os.FileInfo, err error) error {
		if path == pm.PluginDir {
			return nil
		}
		pm.Logger.Debugf("plugin found %v", info.Name())
		if pm.IsPluginEnabled(info.Name()) {
			scriptData, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read data %v", path)
			}

			vm := goja.New()

			script := Script{
				name: info.Name(),
				data: string(scriptData),
				vm:   vm,
				crons: make([]*Cron, 0),
			}

			err = pm.setupVM(&script)
			if err != nil {
				panic(err)
			}

			_, err = vm.RunString(script.data)
			if err != nil {
				pm.Logger.Errorf("failed to load plugin %v:\n%v", script.name, err)
				return nil
			}

			pm.loadedPlugins = append(pm.loadedPlugins, &script)
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	if len(pm.loadedPlugins) == 0 {
		pm.Logger.Warn("no plugins found to load")
	}

	for _, plugin := range pm.loadedPlugins {
		pm.mutex.Lock()
		pm.currentPlugin = plugin

		initFn, ok := goja.AssertFunction(plugin.vm.Get("init"))
		if !ok {
			pm.Logger.Errorf("init of plugin %v not found", plugin.name)
			pm.mutex.Unlock()
			continue
		}

		_, err = initFn(plugin.vm.ToValue(pm.Config))
		if err != nil {
			pm.Logger.Errorf("failed to init plugin %v:\n%v", plugin.name, err)
			pm.mutex.Unlock()
			continue
		}
		pm.mutex.Unlock()
		pm.Logger.Infof("loaded plugin %v", plugin.name)
	}

	_, _ = pm.cronContext.AddFunc("* * * * *", func() {
		now := time.Now().Unix()
		for _, plugin := range pm.loadedPlugins {
			pm.mutex.Lock()
			pm.currentPlugin = plugin
			for _, pluginCron := range plugin.crons {
				if now - pluginCron.lastCron >= pluginCron.cronTime {
					pluginCron.lastCron = now
					_, _ = (*pluginCron.callable)(plugin.vm.ToValue(pm.Config))
				}
			}
			pm.mutex.Unlock()
		}
	})
	pm.cronContext.Start()

	pm.Logger.Infof("Loaded %v routes", pm.routeCount)

	pm.netListener, err = net.Listen("tcp", fmt.Sprintf("%s:%v", pm.Config.WebServer.ListenOn, pm.Config.WebServer.Port))
	if err != nil {
		return err
	}
	pm.Logger.Infof("Listen on %v", pm.netListener.Addr().String())

	go func() {
		_ = fasthttp.Serve(pm.netListener, pm.router.Handler)
	}()

	pm.Logger.Infof("%v plugins loaded", len(pm.loadedPlugins))

	return nil
}

func (pm *PluginHandler) setupVM(s *Script) error {
	s.vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	functions := map[string]interface{}{
		"LogInfo":         pm.scriptLoggerInfo,
		"LogError":        pm.scriptLoggerError,
		"DBPS":            pm.scriptDBPreparedStatement,
		"DBQuery":         pm.scriptDBQuery,
		"AddRoute":        pm.scriptAddRoute,
		"SHA256":          pm.scriptSHA256,
		"HumanizeSeconds": pm.scriptHumanizeSeconds,
		"SendMessage":     pm.scriptSendMessage,
		"AddCron":         pm.scriptAddCron,
		"DoFetch":         pm.scriptFetch,
	}

	for name, function := range functions {
		err := s.vm.Set(name, function)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *PluginHandler) End() {
	_ = pm.netListener.Close()
}

func (pm *PluginHandler) response(rID id.RoomID, msg string) {
	content := format.RenderMarkdown(msg, true, true)
	_, _ = pm.Client.SendMessageEvent(rID, event.EventMessage, &content)
	//pm.Client.SendText(rID, msg)
}

func (pm *PluginHandler) Handle(evt *event.Event, msg string) {
	err := pm.Client.MarkRead(evt.RoomID, evt.ID)
	if err != nil {
		return
	}

	msg = strings.TrimSpace(msg)

	localpart, _, _ := evt.Sender.Parse()

	packet := types.Message{
		Message:         msg,
		Response:        "",
		RoomID:          evt.RoomID.String(),
		EventType:       evt.Type.String(),
		Canceled:        false,
		Sender:          evt.Sender.String(),
		SenderLocalPart: localpart,
	}

	for _, plugin := range pm.loadedPlugins {
		pluginLogger := pm.Logger.WithFields(logrus.Fields{
			"script": plugin.name,
			"event":  evt.ID.String(),
			"func":   "onMessage",
		})
		if packet.Canceled {
			break
		}
		packet.Response = ""
		pm.mutex.Lock()
		pm.currentPlugin = plugin

		onMessageFunc, ok := goja.AssertFunction(plugin.vm.Get("onMessage"))
		if !ok {
			pluginLogger.Error("onMessage not found")
			pm.mutex.Unlock()
			continue
		}

		res, err := onMessageFunc(plugin.vm.ToValue(pm.Config), plugin.vm.ToValue(packet))
		if err != nil {
			pluginLogger.Error("failed to send message to plugin")
			pm.mutex.Unlock()
			continue
		}

		// Skipping if not implemented or just returned
		if !res.StrictEquals(goja.Undefined()) {
			err := plugin.vm.ExportTo(res, &packet)
			if err != nil {
				pluginLogger.Error("failed to parse message response from script")
				pm.mutex.Unlock()
				continue
			}
			if packet.Response != "" {
				pm.response(evt.RoomID, packet.Response)
			}
			if packet.Canceled {
				pluginLogger.Debug("canceled event")
			}
		}

		pm.mutex.Unlock()
	}
}
