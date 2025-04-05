package services

import (
	"encoding/json"
	"fmt"
	// Local Packages
	config "emailer/config"
	models "emailer/models"

	// External Packages
	"go.uber.org/zap"
	"gopkg.in/gomail.v2"
)

type MailProcessor struct {
	Logger *zap.Logger
	Creds  config.Credentials
}

func NewProcessor(logger *zap.Logger, creds config.Credentials) *MailProcessor {
	return &MailProcessor{Logger: logger, Creds: creds}
}

func (p *MailProcessor) ProcessRecord(record models.Record) error {
	var mail models.Mail
	err := json.Unmarshal(record.Value, &mail)
	if err != nil {
		p.Logger.Error("failed to unmarshal mail", zap.Error(err))
	}

	m := gomail.NewMessage()
	m.SetHeader("From", p.Creds.MailID)
	m.SetHeader("To", mail.To)
	m.SetHeader("Subject", mail.Subject)

	if mail.IsHTML {
		m.SetBody("text/html", mail.Body)
	} else {
		m.SetBody("text/plain", mail.Body)
	}

	d := gomail.NewDialer("smtp.gmail.com", 587, p.Creds.MailID, p.Creds.Password)
	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("could not send email: %v", err)
	}
	return nil
}
