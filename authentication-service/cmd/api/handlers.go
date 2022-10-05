package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
)

func (c *Config) Authenticate(w http.ResponseWriter, r *http.Request) {
	log.Println("inside auth srv main")
	var RequestPayload struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	err := c.ReadJson(w, r, &RequestPayload)
	if err != nil {
		log.Println("ERR 1")
		c.ErrorJson(w, err, http.StatusBadRequest)
		return
	}
	log.Println("trying to get by email", RequestPayload)
	user, err := c.Model.User.GetByEmail(RequestPayload.Email)
	if err != nil {
		log.Println("ERR 2")
		c.ErrorJson(w, errors.New("invalid Credentials"), http.StatusBadRequest)
		return
	}
	matches, err := user.PasswordMatches(RequestPayload.Password)
	if err != nil || !matches {
		log.Println("ERR 3")
		c.ErrorJson(w, errors.New("invalid Credentials"), http.StatusBadRequest)
		return
	}

	// log user
	err = c.logRequest("authentication", fmt.Sprintf("logged in user: %v", user.Email))
	if err != nil {
		c.ErrorJson(w, err)
		return
	}

	payload := jsonRes{
		Error:   false,
		Message: fmt.Sprintf("Logged in user: %v", RequestPayload.Email),
		Data:    *user,
	}
	log.Printf("payload %#v", payload)
	c.WriteJson(w, http.StatusOK, payload)
}

func (c *Config) logRequest(name, data string) error {
	var entry struct {
		Name string `json:"name"`
		Data string `json:"data"`
	}
	entry.Name = name
	entry.Data = data
	jsonData, err := json.MarshalIndent(entry, "", "\t")
	if err != nil {
		return err
	}
	logSrvUrl := "http://logger-service/log"
	request, err := http.NewRequest("POST", logSrvUrl, bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json")
	client := http.Client{}

	_, err = client.Do(request)
	if err != nil {
		return err
	}
	return nil
}
