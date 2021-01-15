package utils

import (
	"bytes"
	"encoding/json"
	"net/http"
)

func DispatchRequest(url string, contentType string, reqBody interface{}) (*http.Response, error) {
	reqBodyJson, err := json.Marshal(reqBody)
	if err != nil {
		return nil, err
	}

	resp, err := http.Post(url,
		contentType, bytes.NewBuffer(reqBodyJson))
	if err != nil {
		return nil, err
	}

	return resp, err
}
