package handler

import (
	"github.com/dop251/goja"
)

func (pm *PluginHandler) scriptAddCron(spec goja.Value, callback goja.Value) {
	currentPluginName := pm.currentPlugin.name
	currentMutex := pm.currentPlugin.mutex
	currentVM := pm.currentPlugin.vm
	currentCron := pm.currentPlugin.cron

	scriptLogger := pm.Logger.WithField("script", currentPluginName)

	scriptLogger.Debug("AddCron()")

	var cbGoFunc goja.Callable
	if f, ok := goja.AssertFunction(callback); ok {
		cbGoFunc = f
	} else {
		scriptLogger.Errorln("callback is not a function")
		panic(currentVM.NewTypeError("Not a function"))
	}

	_, err := currentCron.AddFunc(spec.String(), func() {
		currentMutex.Lock()
		_, _ = cbGoFunc(currentVM.ToValue(pm.Config))
		currentMutex.Unlock()
	})

	if err != nil {
		scriptLogger.Errorln("failed to add cron", err)
		panic(currentVM.NewTypeError("failed to add cron"))
	}

	scriptLogger.Debugf("added cron %v", spec.String())
}
