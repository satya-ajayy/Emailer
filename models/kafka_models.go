package models

type Record struct {
	Key   []byte
	Value []byte
	Topic string
}

type ConsumerConfig struct {
	Brokers        []string
	Name           string
	Topic          string
	RecordsPerPoll int
}

type UserData struct {
	UserName string `json:"user_name" schema:"user_name"`
	MailID   string `json:"mail_id" schema:"mail_id"`
}

type Problem struct {
	ID   string `json:"_id" schema:"_id"`
	Name string `json:"name" schema:"name"`
	Link string `json:"link" schema:"link"`
}

type UserLinks struct {
	User     UserData  `json:"user" schema:"user"`
	Problems []Problem `json:"problems" schema:"problems"`
}
