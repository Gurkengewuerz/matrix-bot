package handler

import (
	"github.com/dop251/goja"
)

func (pm *PluginHandler) scriptAddCron(spec goja.Value, callback goja.Value) {
	currentPlugin := pm.currentPlugin
	currentPluginName := currentPlugin.name

	scriptLogger := pm.Logger.WithField("script", currentPluginName)

	scriptLogger.Debug("AddCron()")

	var cbGoFunc goja.Callable
	if f, ok := goja.AssertFunction(callback); ok {
		cbGoFunc = f
	} else {
		scriptLogger.Errorln("callback is not a function")
		panic(currentPlugin.vm.NewTypeError("Not a function"))
	}

	pm.currentPlugin.crons = append(pm.currentPlugin.crons, &Cron{
		cronTime: spec.ToInteger() * 60,
		lastCron: 0,
		callable: &cbGoFunc,
	})

	scriptLogger.Debugf("added %v cron %vm", len(pm.currentPlugin.crons), spec.String())
}
