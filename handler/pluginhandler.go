package handler

import (
	"database/sql"
	"fmt"
	"github.com/dop251/goja"
	"github.com/fasthttp/router"
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
	"os"
	"path/filepath"
	"strings"
	"sync"
)

type Script struct {
	name string
	data string
}

type PluginHandler struct {
	DB            *sql.DB
	PluginDir     string
	Logger        *logrus.Logger
	Config        *config.Config
	Client        *mautrix.Client
	netListener   net.Listener
	loadedPlugins []Script
	currentPlugin *Script
	vm            *goja.Runtime
	mutex         sync.Mutex
	router        *router.Router
	routeCount    int16
}

func (pm *PluginHandler) Index(ctx *fasthttp.RequestCtx) {
	ctx.WriteString("By mc8051.de")
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

	pm.vm = goja.New()
	err := pm.setupVM()
	if err != nil {
		panic(err)
	}

	err = filepath.Walk(pm.PluginDir, func(path string, info os.FileInfo, err error) error {
		if path == pm.PluginDir {
			return nil
		}
		pm.Logger.Debugf("plugin found %v", info.Name())
		if pm.IsPluginEnabled(info.Name()) {
			scriptData, err := ioutil.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read data %v", path)
			}
			pm.loadedPlugins = append(pm.loadedPlugins, Script{
				name: info.Name(),
				data: string(scriptData),
			})
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
		pm.currentPlugin = &plugin
		_, err := pm.vm.RunString(plugin.data)
		if err != nil {
			pm.Logger.Errorf("failed to load plugin %v", plugin.name)
			pm.mutex.Unlock()
			continue
		}

		initFn, ok := goja.AssertFunction(pm.vm.Get("init"))
		if !ok {
			pm.Logger.Errorf("init of plugin %v not found", plugin.name)
			pm.mutex.Unlock()
			continue
		}

		_, err = initFn(pm.vm.ToValue(pm.Config))
		if err != nil {
			pm.Logger.Errorf("failed to init plugin %v", plugin.name)
			pm.mutex.Unlock()
			continue
		}
		pm.mutex.Unlock()
		pm.Logger.Infof("loaded plugin %v", plugin.name)
	}

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

func (pm *PluginHandler) setupVM() error {
	pm.vm.SetFieldNameMapper(goja.UncapFieldNameMapper())
	functions := map[string]interface{}{
		"LogInfo":  pm.scriptLoggerInfo,
		"LogError": pm.scriptLoggerError,
		"DBPS":     pm.scriptDBPreparedStatement,
		"DBQuery":  pm.scriptDBQuery,
		"AddRoute": pm.scriptAddRoute,
		"SHA256":   pm.scriptSHA256,
	}

	for name, function := range functions {
		err := pm.vm.Set(name, function)
		if err != nil {
			return err
		}
	}

	return nil
}

func (pm *PluginHandler) End() {
	pm.netListener.Close()
}

func (pm *PluginHandler) response(rID id.RoomID, msg string) {
	content := format.RenderMarkdown(msg, true, false)
	pm.Client.SendMessageEvent(rID, event.EventMessage, &content)
}

func (pm *PluginHandler) Handle(evt *event.Event, msg string) {
	err := pm.Client.MarkRead(evt.RoomID, evt.ID)
	if err != nil {
		return
	}

	msg = strings.TrimSpace(msg)

	packet := types.Message{
		Message:   msg,
		Response:  "",
		RoomID:    evt.RoomID.String(),
		EventType: evt.Type.String(),
		Canceled:  false,
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
		pm.currentPlugin = &plugin
		_, err := pm.vm.RunString(plugin.data)
		if err != nil {
			pluginLogger.Error("failed to load plugin")
			pm.mutex.Unlock()
			continue
		}

		onMessageFunc, ok := goja.AssertFunction(pm.vm.Get("onMessage"))
		if !ok {
			pluginLogger.Error("onMessage not found")
			pm.mutex.Unlock()
			continue
		}

		res, err := onMessageFunc(pm.vm.ToValue(pm.Config), pm.vm.ToValue(packet))
		if err != nil {
			pluginLogger.Error("failed to send message to plugin")
			pm.mutex.Unlock()
			continue
		}

		// Skipping if not implemented or just returned
		if !res.StrictEquals(goja.Undefined()) {
			err := pm.vm.ExportTo(res, &packet)
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
