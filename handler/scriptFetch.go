package handler

import (
	"github.com/dop251/goja"
	"io"
	"matrix-github-bot/types"
	"net/http"
	"net/url"
	"strings"
)

func (pm *PluginHandler) scriptFetch(value goja.Value) goja.Value {
	currentVM := pm.currentPlugin.vm
	currentPluginName := pm.currentPlugin.name

	scriptLogger := pm.Logger.WithField("script", currentPluginName)

	parsedRequest := types.HTTPRequest{}
	err := currentVM.ExportTo(value, &parsedRequest)
	if err != nil {
		scriptLogger.Error("failed to parse request")
		panic(err)
		return goja.Undefined()
	}

	scriptLogger.Debugf("%v request to %v", parsedRequest.Method, parsedRequest.Url)

	data := url.Values{}

	for key, value := range parsedRequest.Body {
		data.Set(key, value)
	}

	req, err := http.NewRequest(strings.ToUpper(parsedRequest.Method), parsedRequest.Url, strings.NewReader(data.Encode()))
	if err != nil {
		scriptLogger.Error("failed to create request")
		panic(err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=UTF-8")

	for key, value := range parsedRequest.Headers {
		req.Header.Set(key, value)
	}

	req.Header.Set("User-Agent", "mc8051 matrix bot/1.0")

	res, err := pm.httpClient.Do(req)
	if err != nil {
		scriptLogger.Error("failed to send request")
		panic(err)
	}

	buf := new(strings.Builder)
	_, _ = io.Copy(buf, res.Body)
	defer res.Body.Close()

	headers := make(map[string]string)

	for name, _ := range res.Header {
		headers[name] = res.Header.Get(name)
	}

	return currentVM.ToValue(types.HTTPResponse{
		StatusCode:  res.StatusCode,
		Headers:     headers,
		Body:        buf.String(),
	})
}