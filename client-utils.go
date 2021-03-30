package main

import (
	"fmt"
	"maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto"
	"maunium.net/go/mautrix/event"
	"maunium.net/go/mautrix/id"
)

func getUserIDs(cli *mautrix.Client, roomID id.RoomID) []id.UserID {
	members, err := cli.JoinedMembers(roomID)
	if err != nil {
		panic(err)
	}
	userIDs := make([]id.UserID, len(members.Joined))
	i := 0
	for userID := range members.Joined {
		userIDs[i] = userID
		i++
	}
	return userIDs
}

func sendEncrypted(mach *crypto.OlmMachine, cli *mautrix.Client, roomID id.RoomID, text string) {
	content := event.MessageEventContent{
		MsgType: "m.text",
		Body:    text,
	}
	encrypted, err := mach.EncryptMegolmEvent(roomID, event.EventMessage, content)
	// These three errors mean we have to make a new Megolm session
	if err == crypto.SessionExpired || err == crypto.SessionNotShared || err == crypto.NoGroupSession {
		err = mach.ShareGroupSession(roomID, getUserIDs(cli, roomID))
		if err != nil {
			panic(err)
		}
		encrypted, err = mach.EncryptMegolmEvent(roomID, event.EventMessage, content)
	}
	if err != nil {
		panic(err)
	}
	resp, err := cli.SendMessageEvent(roomID, event.EventEncrypted, encrypted)
	if err != nil {
		panic(err)
	}
	fmt.Println("Send response:", resp)
}
