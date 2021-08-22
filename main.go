package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strconv"
)

const (
	requestProtocol              = "http"
	tcpServerHost                = "localhost"
	tcpServerPort                = 8081
	urlGetConnectedClients       = "/clients/connected"
	urlSendMessageToClientsByIds = "/clients/message"
)

type ResourceError struct {
	Url      string
	HttpCode int
	Message  string
	Err      error
	Body     interface{}
}

type ClientParams struct {
	Id       string `json:"id"`
	HttpPort int    `json:"http_port"`
	Name     string `json:"name"`
}

type MessageToClients struct {
	Ids  []string `json:"ids"`
	Text string   `json:"text"`
}

func (r ResourceError) Error() string {
	return fmt.Sprintf("Resource error: URL: %s, http code: %d, err:%v, body:%v.", r.Url, r.HttpCode, r.Err, r.Body)
}

func main() {
	for i := 1; i <= 10; i++ {
		port := tcpServerPort
		cmnd := exec.Command("main.exe", strconv.Itoa(port+i))
		cmnd.Start()
	}

	// Getting all connected clients' params

	var clientParams []ClientParams

	httpStatus, responseBody, err := SendJSONRequest("GET", requestProtocol+"://"+tcpServerHost+":"+strconv.Itoa(tcpServerPort)+urlGetConnectedClients, nil, nil, &clientParams)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error SendJSONRequest . httpCode: %d. responseBody: %v. err: %v.", httpStatus, responseBody, err))
	}

	var message = MessageToClients{
		Ids:  []string{clientParams[1].Id},
		Text: "hello client" + clientParams[1].Name + "! This is a message from client " + clientParams[0].Name,
	}

	data, err := json.Marshal(message)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error json.Marshal message. message:%v", message))
	}

	// Sending message from first client to second
	httpStatus, responseBody, err = SendJSONRequest("POST", requestProtocol+"://"+tcpServerHost+":"+strconv.Itoa(clientParams[0].HttpPort)+urlSendMessageToClientsByIds, data, nil, nil)
	if err != nil {
		log.Fatal(fmt.Sprintf("Error SendJSONRequest. httpCode: %d. responseBody: %v. err: %v.", httpStatus, responseBody, err))
	}
}

func SendJSONRequest(method, url string, data []byte, headers map[string]string, responseStruct interface{}) (httpStatus int, responseBody []byte, err error) {
	if headers == nil {
		headers = make(map[string]string)
	}
	headers["Content-Type"] = "application/json"

	httpStatus, responseBody, err = send(method, url, data, headers)
	if err != nil {
		return
	}

	if responseStruct != nil && len(responseBody) != 0 {
		err = json.Unmarshal(responseBody, responseStruct)
	}

	return
}

func send(method, url string, data []byte, headers map[string]string) (httpStatus int, responseBody []byte, err error) {
	client := http.Client{}
	request, err := http.NewRequest(method, url, bytes.NewBuffer(data))
	if err != nil {
		return httpStatus, nil, &ResourceError{
			Url: url,
			Err: err,
		}
	}

	for i, v := range headers {
		request.Header.Add(i, v)
	}

	response, err := client.Do(request)
	if err != nil {
		return httpStatus, nil, &ResourceError{
			Url: url,
			Err: err,
		}
	}
	defer response.Body.Close()

	responseBody, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return httpStatus, nil, &ResourceError{
			Url:      url,
			Err:      err,
			HttpCode: response.StatusCode,
		}
	}

	httpStatus = response.StatusCode
	return
}
