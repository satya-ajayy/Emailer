package slack

import (
	// Go Internal Packages
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	// Local Packages
	config "emailer/config"
	models "emailer/models"
)

type Text struct {
	Type string `json:"type"`
	Text string `json:"text"`
}
type Block struct {
	Type string `json:"type"`
	Text Text   `json:"text"`
}

type Payload struct {
	Blocks []Block `json:"blocks"`
}

type Sender = func(record models.Record, err error) error

func NewSender(k config.Slack, isProdMode bool) Sender {
	return func(record models.Record, err error) error {
		if (isProdMode) || (!isProdMode && k.SendAlertInDev) {
			var order models.Order
			err = json.Unmarshal(record.Value, &order)
			if err != nil {
				return err
			}

			header := Block{
				Type: "header",
				Text: Text{
					Type: "plain_text",
					Text: "Error in Emailer",
				},
			}
			body := Block{
				Type: "section",
				Text: Text{
					Type: "mrkdwn",
					Text: fmt.Sprintf("```OrderID:%s\nCustomerName:%s\nError:%s\n```",
						order.ID, order.Customer.Name, err.Error()),
				},
			}
			payload := Payload{
				Blocks: []Block{header, body},
			}
			jsonPayload, _ := json.Marshal(payload)
			_, err = http.Post(k.WebhookURL, "application/json", bytes.NewReader(jsonPayload))
			return err
		}
		return nil
	}
}
