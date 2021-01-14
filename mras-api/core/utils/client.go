package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

func DispatchRequest(url string,contentType string){
	reqBody, err := json.Marshal(map[string]string{
		"username": "Goher Go",
		"email":    "go@gmail.com",
	})

	if err != nil {
		print(err)
	}

	resp, err := http.Post(url,
		contentType, bytes.NewBuffer(reqBody))
	if err != nil {
		print(err)
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		print(err)
	}
	fmt.Println(string(body))
}
