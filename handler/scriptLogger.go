package handler

import (
	"github.com/dop251/goja"
)

func (pm *PluginHandler) scriptLoggerInfo(value goja.Value) {
	pm.Logger.WithField("script", pm.currentPlugin.name).Info(value.String())
}

func (pm *PluginHandler) scriptLoggerError(value goja.Value) {
	pm.Logger.WithField("script", pm.currentPlugin.name).Error(value.String())
}
