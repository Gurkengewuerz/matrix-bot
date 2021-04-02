package handler

import (
	"github.com/dop251/goja"
	"maunium.net/go/mautrix/id"
)

func (pm *PluginHandler) scriptSendMessage(room goja.Value, message goja.Value) {
	pm.response(id.RoomID(room.String()), message.String())
}