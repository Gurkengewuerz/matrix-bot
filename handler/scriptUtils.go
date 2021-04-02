package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/dop251/goja"
	"io"
	"strings"
	"time"
)

func (pm *PluginHandler) scriptSHA256(val goja.Value) goja.Value {
	h := sha256.New()
	_, err := io.WriteString(h, val.String())
	if err != nil {
		panic(pm.vm.ToValue(err))
	}
	return pm.vm.ToValue(hex.EncodeToString(h.Sum(nil)))
}

const (
	day  = time.Minute * 60 * 24
	year = 365 * day
)

func (pm *PluginHandler) scriptHumanizeSeconds(val goja.Value) goja.Value {

	d := time.Duration(val.ToInteger()) * time.Second

	if d < day {
		return pm.vm.ToValue(d.String())
	}

	var b strings.Builder

	if d >= year {
		years := d / year
		_, _ = fmt.Fprintf(&b, "%dy", years)
		d -= years * year
	}

	days := d / day
	d -= days * day
	_, _ = fmt.Fprintf(&b, "%dd%s", days, d)

	return pm.vm.ToValue(b.String())
}