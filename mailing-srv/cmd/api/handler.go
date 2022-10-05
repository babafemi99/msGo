package main

import (
	"log"
	"net/http"
)

type MailMessage struct {
	From    string `json:"from"`
	To      string `json:"to"`
	Subject string `json:"subject"`
	Message string `json:"message"`
}

func (c *Config) SendMail(writer http.ResponseWriter, request *http.Request) {
	var requestPayload MailMessage
	err := c.ReadJson(writer, request, &requestPayload)
	if err != nil {
		log.Println(1)
		c.ErrorJson(writer, err)
		return
	}
	msg := Message{
		From:    requestPayload.From,
		To:      requestPayload.To,
		Subject: requestPayload.Subject,
		Data:    requestPayload.Message,
	}

	err = c.Mail.SendMessage(msg)
	if err != nil {
		log.Println(2)
		c.ErrorJson(writer, err)
		return
	}
	payload := jsonRes{
		Error:   false,
		Message: "sent to" + msg.To,
	}
	c.WriteJson(writer, http.StatusAccepted, &payload)
}
