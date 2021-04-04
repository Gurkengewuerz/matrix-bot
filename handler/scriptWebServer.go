package handler

import (
	"encoding/json"
	"github.com/dop251/goja"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"matrix-bot/types"
	"strings"
)

func (pm *PluginHandler) scriptAddRoute(path goja.Value, method goja.Value, callback goja.Value) {
	currentPlugin := pm.currentPlugin
	currentPluginName := currentPlugin.name
	currentVM := currentPlugin.vm
	pm.Logger.WithField("script", currentPluginName).Debug("AddRoute()")

	var cbGoFunc goja.Callable
	if f, ok := goja.AssertFunction(callback); ok {
		cbGoFunc = f
	} else {
		pm.Logger.WithField("script", currentPluginName).Errorln("callback is not a function")
		panic(currentVM.NewTypeError("Not a function"))
	}

	routerCallback := func(ctx *fasthttp.RequestCtx) {
		routerLog := pm.Logger.WithFields(logrus.Fields{
			"script": currentPluginName,
			"path":   path.String(),
		})
		pm.mutex.Lock()
		pm.currentPlugin = currentPlugin

		jsonMap := make(map[string]interface{})
		buf := ctx.PostBody()
		if len(buf) > 0 {
			err := json.Unmarshal(ctx.PostBody(), &jsonMap)
			if err != nil {
				panic(currentVM.ToValue(err))
			}
		}

		userData := make(map[string]interface{})
		ctx.VisitUserValues(func(key []byte, i interface{}) {
			userData[string(key)] = i
		})

		headerData := make(map[string]string)
		ctx.Request.Header.VisitAll(func(key []byte, i []byte) {
			headerData[string(key)] = string(i)
		})

		res, err := cbGoFunc(currentVM.ToValue(pm.Config), currentVM.ToValue(types.HTTPCall{
			MatchedPath: path.String(),
			Path:        string(ctx.Path()),
			StatusCode:  200,
			Body:        jsonMap,
			ContentType: "application/json",
			Response:    `{"ok": true}`,
			Params:      userData,
			Headers:     headerData,
		}))
		if err != nil {
			routerLog.Errorf("failed to run callback function: %v", err)
			pm.mutex.Unlock()
			return
		}

		if res.StrictEquals(goja.Undefined()) {
			routerLog.Warn("does not return anything")
			ctx.SetStatusCode(200)
			ctx.SetContentType("application/json")
			ctx.SetBodyString(`{"ok": true}`)
		} else {
			parsedResponse := types.HTTPCall{}
			err := currentVM.ExportTo(res, &parsedResponse)
			if err != nil {
				routerLog.Error("failed to parse response from script")
				pm.mutex.Unlock()
				return
			}
			ctx.SetStatusCode(parsedResponse.StatusCode)
			ctx.SetContentType(parsedResponse.ContentType)
			ctx.SetBodyString(parsedResponse.Response)
		}

		pm.mutex.Unlock()
	}

	switch strings.ToUpper(method.String()) {
	case "GET":
		pm.router.GET(path.String(), routerCallback)
		break
	case "POST":
		pm.router.POST(path.String(), routerCallback)
		break
	case "PUT":
		pm.router.PUT(path.String(), routerCallback)
		break
	case "DELETE":
		pm.router.DELETE(path.String(), routerCallback)
		break
	default:
		panic(currentVM.ToValue("unknown http method"))
	}
	pm.routeCount++
	pm.Logger.WithField("script", currentPluginName).Debugf("added route %v", path.String())
}
