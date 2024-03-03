package gomb

import "time"

type Message struct {
	Key        string    `json:"key"`
	Val        []byte    `json:"val"`
	SentAt     time.Time `json:"sentAt"`
	ID         string    `json:"ID"`
	ProducerID string    `json:"producerID"`
}
