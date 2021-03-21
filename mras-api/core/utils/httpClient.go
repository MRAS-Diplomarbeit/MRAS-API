package utils

import (
	"bytes"
	"encoding/json"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"net/http"
	"strings"
)

func DispatchRequest(url string, contentType string, method string, reqBody interface{}) (*http.Response, error) {
	Log.WithField("module", "requestDispatch").Debug("sending Request: " + url)
	Log.WithField("module", "requestDispatch").Debug(reqBody)

	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	Log.WithField("module", "requestDispatch").Debug(string(reqBodyJson))

	var resp *http.Response
	client := &http.Client{}

	if strings.ToUpper(method) == "POST" {
		resp, err = http.Post(url, contentType, bytes.NewBuffer(reqBodyJson))
		if err != nil {
			return nil, err
		}
	} else if strings.ToUpper(method) == "POST" {
		resp, err = http.Get(url)
		if err != nil {
			return nil, err
		}
	} else if strings.ToUpper(method) == "DELETE" {
		req, err := http.NewRequest("DELETE", url, bytes.NewBuffer(reqBodyJson))
		req.Header.Add("Content-Type", contentType)
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
	} else if strings.ToUpper(method) == "PUT" {
		req, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBodyJson))
		req.Header.Add("Content-Type", contentType)
		resp, err = client.Do(req)
		if err != nil {
			return nil, err
		}
	}

	Log.WithField("module", "requestDispatch").Debug(resp)

	return resp, err
}
