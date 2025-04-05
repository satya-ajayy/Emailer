package models

type Record struct {
	Key   []byte
	Value []byte
	Topic string
}

type Mail struct {
	To      string
	Subject string
	Body    string
	IsHTML  bool
}
