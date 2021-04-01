package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"github.com/dop251/goja"
	"io"
)

func (pm *PluginHandler) scriptSHA256(val goja.Value) goja.Value {
	h := sha256.New()
	_, err := io.WriteString(h, val.String())
	if err != nil {
		panic(pm.vm.ToValue(err))
	}
	return pm.vm.ToValue(hex.EncodeToString(h.Sum(nil)))
}
