package utils

import (
	"bytes"
	"encoding/json"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"net/http"
)

func DispatchRequest(url string, contentType string, reqBody interface{}) (*http.Response, error) {
	Log.WithField("module", "requestDispatch").Debug("sending Request: " + url)
	Log.WithField("module", "requestDispatch").Debug(reqBody)

	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	Log.WithField("module", "requestDispatch").Debug(string(reqBodyJson))

	resp, err := http.Post(url, contentType, bytes.NewBuffer(reqBodyJson))
	if err != nil {
		return nil, err
	}

	Log.WithField("module", "requestDispatch").Debug(resp)

	return resp, err
}
