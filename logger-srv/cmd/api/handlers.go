package main

import (
	"log"
	"logger-srv/cmd/data"
	"net/http"
)

type JsonPayload struct {
	Name string `json:"name"`
	Data string `json:"data"`
}

func (c *Config) WriteToLog(w http.ResponseWriter, r *http.Request) {
	log.Println("inside write to log")
	var reqPayload JsonPayload
	err := c.ReadJson(w, r, &reqPayload)
	if err != nil {
		return
	}
	event := data.LogEntry{
		Name: reqPayload.Name,
		Data: reqPayload.Data,
	}
	log.Println("event is", event)
	log.Println("payload is", reqPayload)
	err = c.Models.LogEntry.Insert(event)
	if err != nil {
		c.ErrorJson(w, err, http.StatusInternalServerError)
		return
	}
	resp := jsonRes{
		Error:   false,
		Message: "logged",
	}
	c.WriteJson(w, http.StatusAccepted, resp)

}
