package client

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
)

// HandleRequest send message to tg
func HandleRequest(httpMethod string, url string, reqBody []byte) ([]byte, error) {
	req, err := http.NewRequest(httpMethod, url, bytes.NewReader(reqBody))
	if err != nil {
		log.Printf("new request [%s] %s error %s", httpMethod, url, err)
		return nil, err
	}

	req.Header.Add("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Printf("client Do failed: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Printf("readAll failed: %v", err)
		return nil, err
	}

	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		log.Printf("receive body:%s\n", body)
		return body, nil
	}
	return body, fmt.Errorf("server return error %+v", resp)

}
