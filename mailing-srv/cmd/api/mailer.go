package main

import (
	"bytes"
	"github.com/vanng822/go-premailer/premailer"
	mail "github.com/xhit/go-simple-mail/v2"
	"html/template"
	"log"
	"time"
)

type Mail struct {
	Domain      string
	Host        string
	Port        int
	Username    string
	Password    string
	Encryption  string
	FromAddress string
	FromName    string
}

type Message struct {
	From        string
	FromName    string
	To          string
	Subject     string
	Attachments []string
	Data        any
	DataMap     map[string]any
}

func (m *Mail) SendMessage(message Message) error {
	if message.From == "" {
		message.From = m.FromAddress
	}
	if message.FromName == "" {
		message.FromName = m.FromName
	}
	data := map[string]any{
		"message": message.Data,
	}
	message.DataMap = data
	formattedMessage, err := m.BuildHTMLMessage(message)
	if err != nil {
		log.Println("error here", err)
		return err
	}
	plainMsg, err := m.BuildPlainTextMessage(message)
	if err != nil {
		log.Println("error here", err)
		return err
	}

	server := mail.NewSMTPClient()
	server.Host = m.Host
	server.Port = m.Port
	server.Username = m.Username
	server.Password = m.Password
	server.Encryption = m.getEncryption(m.Encryption)
	server.KeepAlive = false
	server.ConnectTimeout = 10 * time.Second
	server.SendTimeout = 10 * time.Second

	client, err := server.Connect()
	if err != nil {
		return err
	}
	msg := mail.NewMSG()
	msg.SetFrom(message.From).AddTo(message.To).SetSubject(message.Subject).SetBody(mail.TextPlain,
		plainMsg).AddAlternative(mail.TextHTML, formattedMessage)

	if len(message.Attachments) > 0 {
		for _, x := range message.Attachments {
			msg.AddAttachment(x)
		}
	}

	err = msg.Send(client)
	if err != nil {
		return err
	}
	return nil
}

func (m *Mail) BuildHTMLMessage(message Message) (string, error) {
	tmplToRender := "./templates/main.gohtml"
	tmpl, err := template.New("email-html").ParseFiles(tmplToRender)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err := tmpl.ExecuteTemplate(&tpl, "body", message.DataMap); err != nil {
		return "", err
	}

	formattedMsg := tpl.String()
	formattedMsg, err = m.InlineCss(formattedMsg)

	return formattedMsg, nil
}

func (m *Mail) InlineCss(msg string) (string, error) {
	opts := premailer.Options{
		RemoveClasses:     false,
		CssToAttributes:   false,
		KeepBangImportant: true,
	}
	prem, err := premailer.NewPremailerFromString(msg, &opts)
	if err != nil {
		return "", err
	}

	html, err := prem.Transform()
	if err != nil {
		return "", err
	}
	return html, nil
}

func (m *Mail) BuildPlainTextMessage(message Message) (string, error) {
	tmplToRender := "./templates/main.plain.gohtml"
	tmpl, err := template.New("email-plain").ParseFiles(tmplToRender)
	if err != nil {
		return "", err
	}
	var tpl bytes.Buffer
	if err := tmpl.ExecuteTemplate(&tpl, "body", message.DataMap); err != nil {
		return "", err
	}

	plainMessage := tpl.String()

	return plainMessage, nil
}

func (m *Mail) getEncryption(s string) mail.Encryption {
	switch s {
	case "tls":
		return mail.EncryptionSTARTTLS
	case "ssl":
		return mail.EncryptionSSLTLS
	case "none", "":
		return mail.EncryptionNone
	default:
		return mail.EncryptionSTARTTLS
	}
}
