package processors

import (
	// Go Internal Packages
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"html/template"

	// Local Packages
	config "emailer/config"
	models "emailer/models"

	// External Packages
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type MailProcessor struct {
	logger *zap.Logger
	creds  config.Credentials
}

func NewProcessor(logger *zap.Logger, creds config.Credentials) *MailProcessor {
	return &MailProcessor{logger: logger, creds: creds}
}

func (p *MailProcessor) ProcessRecord(ctx context.Context, record models.Record) error {
	var userLinks models.UserLinks
	err := json.Unmarshal(record.Value, &userLinks)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", p.creds.MailID)
	m.SetHeader("To", userLinks.User.MailID)
	m.SetHeader("Subject", "!!! Checkout Our New Problems !!!")
	body, err := p.GetHTML(userLinks)
	if err != nil {
		return fmt.Errorf("error getting HTML: %v", err)
	}

	m.SetBody("text/html", body)
	d := gomail.NewDialer("smtp.gmail.com", 587, p.creds.MailID, p.creds.Password)
	if err = d.DialAndSend(m); err != nil {
		return fmt.Errorf("could not send email: %v", err)
	}
	return nil
}

func (p *MailProcessor) GetHTML(userLinks models.UserLinks) (string, error) {
	problemsTemplate, err := template.New("problems.html").
		Funcs(template.FuncMap{"call": func(f interface{}) interface{} { return f.(func() string)() }}).
		ParseFiles("templates/problems.html")
	if err != nil {
		return "", fmt.Errorf("template parse error: %v", err)
	}

	var body bytes.Buffer
	err = problemsTemplate.Execute(&body, userLinks)
	if err != nil {
		return "", fmt.Errorf("template execution error: %v", err)
	}

	return body.String(), nil
}
