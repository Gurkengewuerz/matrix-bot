package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"github.com/sirupsen/logrus"
	"log"
	"matrix-github-bot/config"
	"matrix-github-bot/handler"
	"matrix-github-bot/logger"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
	"os"
	"path"
	"path/filepath"
	"time"
)

var Log *logrus.Logger
var CFG *config.Config

var pluginPath = flag.String("plugin", "", "Plugin path")
var configPath = flag.String("config", "", "Config Path")
var isHelp = flag.Bool("help", false, "Help command")

func main() {
	flag.Parse()
	if *isHelp {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	start := time.Now().UnixNano() / 1000000

	binaryDir, err := filepath.Abs(filepath.Dir(os.Args[0]))
	if err != nil {
		panic(err)
	}

	if *configPath == "" {
		*configPath = path.Join(binaryDir, "./config.yaml")
	}

	CFG, err := config.Load(*configPath)
	if err != nil {
		panic(err)
	}

	Log = logger.InitLogger(CFG.Logger.Debug)

	client, err := mautrix.NewClient(CFG.Homeserver, "", "")
	if err != nil {
		panic(err)
	}
	flows, err := client.GetLoginFlows()
	if err != nil {
		panic(err)
	}
	if !flows.HasFlow(mautrix.AuthTypePassword) {
		panic(fmt.Errorf("homeserver does not support password login"))
	}
	req := &mautrix.ReqLogin{
		Type:                     mautrix.AuthTypePassword,
		Identifier:               mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: CFG.Bot.Username},
		Password:                 CFG.Bot.Password,
		InitialDeviceDisplayName: CFG.Bot.Displayname,
		StoreCredentials:         true,
	}
	if CFG.Bot.DeviceID != "" {
		req.DeviceID = id.DeviceID(CFG.Bot.DeviceID)
	}

	resp, err := client.Login(req)
	if err != nil {
		panic(err)
	}

	Log.Info("logged in")

	if CFG.Bot.DeviceID == "" {
		log.Println(resp.DeviceID)
		os.Exit(1)
	}

	Log.Debug("updating device info")
	err = client.SetDeviceInfo(client.DeviceID, &mautrix.ReqDeviceInfo{DisplayName: CFG.Bot.Displayname})
	if err != nil {
		panic(err)
	}

	Log.Debug("updating display name")
	err = client.SetDisplayName(CFG.Bot.Displayname)
	if err != nil {
		panic(err)
	}

	if len(CFG.DBFile) == 0 {
		CFG.DBFile = path.Join(binaryDir, "./data.db")
	}

	db, err := sql.Open("sqlite3", CFG.DBFile)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if *pluginPath == "" {
		*pluginPath = path.Join(binaryDir, "./plugins/")
	}

	pluginHandler := handler.PluginHandler{
		DB:        db,
		PluginDir: *pluginPath,
		Logger:    Log,
		Config:    CFG,
		Client:    client,
	}
	err = pluginHandler.Init()
	if err != nil {
		panic(err)
	}
	defer pluginHandler.End()

	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		if evt.Timestamp < start {
			// Ignore events from before the program started
			return
		}
		if evt.Content.AsMember().Membership == event.MembershipInvite && client.UserID.String() == *evt.StateKey {
			_, err := client.JoinRoomByID(evt.RoomID)
			roomLogger := Log.WithFields(logrus.Fields{
				"inviter": evt.Sender,
				"room":    evt.RoomID,
			})
			if err != nil {
				roomLogger.Error("Failed to join room")
			} else {
				roomLogger.Info("Joined room")
			}
		}
	})

	syncer.OnEventType(event.EventMessage, func(source mautrix.EventSource, evt *event.Event) {
		if evt.Timestamp < start {
			// Ignore events from before the program started
			return
		}
		message, isMessage := evt.Content.Parsed.(*event.MessageEventContent)
		if isMessage {
			pluginHandler.Handle(evt, message.Body)
		}
	})

	syncer.OnEvent(func(source mautrix.EventSource, evt *event.Event) {
		if evt.Timestamp < start {
			// Ignore events from before the program started
			return
		}
		Log.WithFields(logrus.Fields{
			"sender": evt.Sender,
			"type":   evt.Type.String(),
			"id":     evt.ID,
		}).Debug(evt.Content.AsMessage().Body)
	})

	// Start long polling in the background
	go func() {
		err = client.Sync()
		if err != nil {
			panic(err)
		}
	}()

	for {
		time.Sleep(5 * time.Second)
	}
}
