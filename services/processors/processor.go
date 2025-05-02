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

type OrdersRepository interface {
	GetOrder(ctx context.Context, orderID string) (models.Order, error)
}

type MailProcessor struct {
	logger     *zap.Logger
	ordersRepo OrdersRepository
	creds      config.Credentials
}

func NewProcessor(logger *zap.Logger, creds config.Credentials, repo OrdersRepository) *MailProcessor {
	return &MailProcessor{logger: logger, creds: creds, ordersRepo: repo}
}

func (p *MailProcessor) ProcessRecord(ctx context.Context, record models.Record) error {
	var mail models.MailQP
	err := json.Unmarshal(record.Value, &mail)
	if err != nil {
		return fmt.Errorf("error unmarshalling JSON: %v", err)
	}

	order, err := p.ordersRepo.GetOrder(ctx, mail.Order)
	if err != nil {
		return fmt.Errorf("error getting order: %v", err)
	}

	m := gomail.NewMessage()
	m.SetHeader("From", p.creds.MailID)
	m.SetHeader("To", order.Customer.MailID)
	m.SetHeader("Subject", mail.Header)
	body, err := p.GetHTML(order)
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

func (p *MailProcessor) GetHTML(order models.Order) (string, error) {
	successTemplate, err := template.New("success.html").
		Funcs(template.FuncMap{"call": func(f interface{}) interface{} { return f.(func() string)() }}).
		ParseFiles("templates/success.html")
	if err != nil {
		return "", fmt.Errorf("template parse error:: %v", err)
	}

	var body bytes.Buffer
	err = successTemplate.Execute(&body, order)
	if err != nil {
		return "", fmt.Errorf("template execution error: %v", err)
	}

	return body.String(), nil
}
