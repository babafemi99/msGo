package main

import (
	"bsv/cmd/events"
	"bsv/cmd/logs"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net/http"
	"net/rpc"
	"time"
)

type RequestPayload struct {
	Action string      `json:"action"`
	Auth   AuthPayload `json:"auth,omitempty"`
	Log    LogPayload  `json:"log,omitempty"`
	Mail   MailPayload `json:"mail,omitempty"`
}

type AuthPayload struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type MailPayload struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}
type LogPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func Broker(w http.ResponseWriter, r *http.Request) {
	fmt.Println("in here")
	payload := jsonRes{
		Error:   false,
		Message: "hit broker",
	}

	marshal, _ := json.MarshalIndent(payload, "", "\t")
	w.Header().Set("Content-Type", "Application/json")
	w.WriteHeader(http.StatusAccepted)
	w.Write(marshal)
}

func (c *Config) HandleSubmission(w http.ResponseWriter, r *http.Request) {
	var requestPayload RequestPayload
	err := c.ReadJson(w, r, &requestPayload)
	if err != nil {
		c.ErrorJson(w, err, http.StatusBadRequest)
		return
	}
	switch requestPayload.Action {
	case "auth":
		c.authenticate(w, r, requestPayload.Auth)
	case "log":
		c.logViaRPC(w, requestPayload.Log)
	case "mail":
		c.sendMail(w, requestPayload.Mail)
	default:
		c.ErrorJson(w, err, http.StatusBadRequest)
	}

}

func (c *Config) authenticate(w http.ResponseWriter, r *http.Request, a AuthPayload) {
	// create json to send to auth microservice
	jsonData, je := json.MarshalIndent(a, "", "\t")
	if je != nil {
		c.ErrorJson(w, errors.New("error indenting"))
		return
	}
	// call the auth microservice
	request, err := http.NewRequest("POST", "http://authentication-service/authenticate", bytes.NewBuffer(jsonData))
	if err != nil {
		c.ErrorJson(w, err)
		return
	}
	client := &http.Client{}
	response, err := client.Do(request)
	if err != nil {
		c.ErrorJson(w, err)
		return
	}
	defer response.Body.Close()
	// make sure we get back correct status code
	if response.StatusCode == http.StatusUnauthorized {
		c.ErrorJson(w, err)
		return
	} else if response.StatusCode != http.StatusOK {
		c.ErrorJson(w, errors.New("error calling auth service"+err.Error()))
		return
	}

	// create a variable to read response.body to
	var jsonFromSrv jsonRes
	err = json.NewDecoder(response.Body).Decode(&jsonFromSrv)
	if err != nil {
		c.ErrorJson(w, err)
		return
	}
	if jsonFromSrv.Error {
		c.ErrorJson(w, err, http.StatusUnauthorized)
		return
	}
	var payload jsonRes
	payload.Error = false
	payload.Message = "Authenticated"
	payload.Data = jsonFromSrv.Data

	c.WriteJson(w, http.StatusAccepted, jsonFromSrv.Data)
	fmt.Println("exit auth")
}

func (c *Config) logItem(w http.ResponseWriter, l LogPayload) {
	jsonData, je := json.MarshalIndent(l, "", "\t")
	if je != nil {
		c.ErrorJson(w, errors.New("error indenting"))
		return
	}
	logSrvUrl := "http://logger-service/log"
	request, err := http.NewRequest("POST", logSrvUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		c.ErrorJson(w, err)
		return
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}

	res, err := client.Do(request)
	if err != nil {
		c.ErrorJson(w, err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		c.ErrorJson(w, err)
		return
	}

	var payload jsonRes
	payload.Error = false
	payload.Message = "logged"
	c.WriteJson(w, http.StatusAccepted, payload)
}

func (c *Config) sendMail(w http.ResponseWriter, payload MailPayload) {
	jsonData, err := json.MarshalIndent(payload, "", "\t")
	if err != nil {
		c.ErrorJson(w, err)
		return
	}
	mailSrvUrl := "http://mailing-service/send"
	request, err := http.NewRequest("POST", mailSrvUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		c.ErrorJson(w, err)
		return
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}

	res, err := client.Do(request)
	if err != nil {
		c.ErrorJson(w, err)
		return
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusAccepted {
		c.ErrorJson(w, err)
		return
	}

	var payloadFinal jsonRes
	payloadFinal.Error = false
	payloadFinal.Message = "message sent to msg service"
	c.WriteJson(w, http.StatusAccepted, payloadFinal)

}

func (c *Config) logWithEvent(w http.ResponseWriter, l LogPayload) {
	err := c.pushToQueue(l.Name, l.Data)
	if err != nil {
		c.ErrorJson(w, err)
	}
	var payload jsonRes
	payload.Error = false
	payload.Message = "logged via event handler"
	c.WriteJson(w, http.StatusAccepted, payload)
}

func (c *Config) pushToQueue(name, message string) error {
	emitter, err := events.NewEventEmitter(c.Rabbit)
	if err != nil {
		return err
	}
	payload := LogPayload{
		Name: name,
		Data: message,
	}

	j, _ := json.MarshalIndent(payload, "", "\t")
	err = emitter.Push(string(j), "log.INFO")
	if err != nil {
		return err
	}
	return nil
}

type RPCPayload struct {
	Name string
	Data string
}

func (c *Config) logViaRPC(w http.ResponseWriter, l LogPayload) {
	client, err := rpc.Dial("tcp", "logger-service:50051")
	if err != nil {
		c.ErrorJson(w, err)
		return
	}

	payload := RPCPayload{
		Name: l.Name,
		Data: l.Data,
	}
	var result string
	err = client.Call("RPCServer.LogInfo", payload, &result)
	if err != nil {
		c.ErrorJson(w, err)
		return
	}

	data := jsonRes{
		Error:   false,
		Message: result,
	}
	c.WriteJson(w, http.StatusAccepted, &data)
}

func (c *Config) logViaGRPC(w http.ResponseWriter, r *http.Request) {
	var request RequestPayload
	err := c.ReadJson(w, r, &request)
	if err != nil {
		c.ErrorJson(w, err)
		return
	}

	conn, err := grpc.Dial("logger-service:50001", grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	if err != nil {
		c.ErrorJson(w, err)
		return
	}
	defer conn.Close()

	client := logs.NewLogServiceClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	_, err = client.WriteLog(ctx, &logs.LogRequest{LogEntry: &logs.Log{
		Name: request.Log.Name,
		Data: request.Log.Data,
	}})
	if err != nil {
		c.ErrorJson(w, err)
		return
	}

	data := jsonRes{
		Error:   false,
		Message: "logged",
	}
	c.WriteJson(w, http.StatusAccepted, &data)
}
