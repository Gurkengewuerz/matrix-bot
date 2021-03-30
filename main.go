package main

import (
	"database/sql"
	"flag"
	"fmt"
	_ "github.com/mattn/go-sqlite3"
	"log"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
	"os"
	"time"
)

var homeserver = flag.String("homeserver", "", "Matrix homeserver")
var username = flag.String("username", "", "Matrix username localpart")
var password = flag.String("password", "", "Matrix password")

func main() {
	start := time.Now().UnixNano() / 1000000

	flag.Parse()
	if *username == "" || *password == "" || *homeserver == "" {
		_, _ = fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", "./foo.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	fmt.Println("Logging into", *homeserver, "as", *username)
	client, err := mautrix.NewClient(*homeserver, "", "")
	if err != nil {
		panic(err)
	}
	_, err = client.Login(&mautrix.ReqLogin{
		Type:                     "m.login.password",
		Identifier:               mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: *username},
		Password:                 *password,
		InitialDeviceDisplayName: "by mc8051.de",
		StoreCredentials:         true,
	})
	if err != nil {
		panic(err)
	}
	// Log out when the program ends (don't do this in real apps)
	defer func() {
		fmt.Println("Logging out")
		resp, err := client.Logout()
		if err != nil {
			fmt.Println("Logout error:", err)
		}
		fmt.Println("Logout response:", resp)
	}()

	fmt.Println("Login successful")

	sqlStore := crypto.NewSQLCryptoStore(db, "sqlite3", client.UserID.String(), client.DeviceID, []byte{}, &fakeLogger{})

	mach := crypto.NewOlmMachine(client, &fakeLogger{}, sqlStore, &fakeStateStore{})
	// Load data from the crypto store
	err = mach.Load()
	if err != nil {
		panic(err)
	}

	// Hook up the OlmMachine into the Matrix client so it receives e2ee keys and other such things.
	syncer := client.Syncer.(*mautrix.DefaultSyncer)
	syncer.OnSync(func(resp *mautrix.RespSync, since string) bool {
		mach.ProcessSyncResponse(resp, since)
		return true
	})
	syncer.OnEventType(event.StateMember, func(source mautrix.EventSource, evt *event.Event) {
		mach.HandleMemberEvent(evt)
	})
	// Listen to encrypted messages
	syncer.OnEventType(event.EventEncrypted, func(source mautrix.EventSource, evt *event.Event) {
		if evt.Timestamp < start {
			// Ignore events from before the program started
			return
		}
		decrypted, err := mach.DecryptMegolmEvent(evt)
		if err != nil {
			fmt.Println("Failed to decrypt:", err)
		} else {
			fmt.Println("Received encrypted event:", decrypted.Content.Raw)
			message, isMessage := decrypted.Content.Parsed.(*event.MessageEventContent)
			if isMessage && message.Body == "ping" {
				sendEncrypted(mach, client, decrypted.RoomID, "Pong!")
			}
		}
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
