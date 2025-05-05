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

type OrderInKafka struct {
	ID     string `json:"id" schema:"id"`
	Type   string `json:"type" schema:"type"`
	Header string `json:"header" schema:"header"`
}
