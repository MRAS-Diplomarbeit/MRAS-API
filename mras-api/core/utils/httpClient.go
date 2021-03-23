package utils

import (
	"bytes"
	"encoding/json"
	. "github.com/mras-diplomarbeit/mras-api/core/logger"
	"net/http"
)

func GetRequest(url string) (*http.Response, error) {
	Log.WithField("module", "requestDispatch").Debug("sending Request: " + url)
	return http.Get(url)
}

func PostRequest(url string, contentType string, body interface{}) (*http.Response, error) {
	Log.WithField("module", "requestDispatch").Debug("sending Request: " + url)

	reqBody, err := json.Marshal(body)
	if err != nil {
		Log.WithField("module", "requestDispatch").WithError(err)
		return nil, err
	}

	return http.Post(url, contentType, bytes.NewBuffer(reqBody))
}

func PutRequest(url string, contentType string, body interface{}) (*http.Response, error) {
	Log.WithField("module", "requestDispatch").Debug("sending Request: " + url)

	reqBody, err := json.Marshal(body)
	if err != nil {
		Log.WithField("module", "requestDispatch").WithError(err)
		return nil, err
	}

	client := http.Client{}

	request, err := http.NewRequest("PUT", url, bytes.NewBuffer(reqBody))
	if err != nil {
		Log.WithField("module", "requestDispatch").WithError(err)
		return nil, err
	}
	request.Header.Set("Content-Type", contentType)

	return client.Do(request)
}

func DeleteRequest(url string, contentType string, body interface{}) (*http.Response, error) {
	Log.WithField("module", "requestDispatch").Debug("sending Request: " + url)

	reqBody, err := json.Marshal(body)
	if err != nil {
		Log.WithField("module", "requestDispatch").WithError(err)
		return nil, err
	}

	client := http.Client{}

	request, err := http.NewRequest("DELETE", url, bytes.NewBuffer(reqBody))
	if err != nil {
		Log.WithField("module", "requestDispatch").WithError(err)
		return nil, err
	}
	request.Header.Set("Content-Type", contentType)

	return client.Do(request)
}
