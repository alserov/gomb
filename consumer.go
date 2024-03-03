package gomb

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
)

type Consumer struct {
	Addr  string
	Topic string
	ID    string

	conn *websocket.Conn
}

func NewConsumer(p Params) (*Consumer, error) {
	u := url.URL{Scheme: "ws", Host: p.Addr, Path: fmt.Sprintf("/consume/%s", p.Topic)}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"X-Producer-ID": []string{p.ID}})
	if err != nil {
		return nil, fmt.Errorf("failed to server with server: %w", err)
	}

	return &Consumer{
		Addr:  p.Addr,
		Topic: p.Topic,
		ID:    p.ID,
		conn:  conn,
	}, nil
}

func (c *Consumer) Consume() (<-chan Message, <-chan error) {
	chMsgs := make(chan Message, 128)
	chErrors := make(chan error, 8)

	go func() {
		for {
			_, bytes, err := c.conn.ReadMessage()
			if err != nil {
				chErrors <- fmt.Errorf("failed to consume message: %s", err.Error())
				continue
			}

			var msg Message
			err = json.Unmarshal(bytes, &msg)
			if err != nil {
				chErrors <- fmt.Errorf("failed to decode message: %s", err.Error())
				continue
			}

			chMsgs <- msg
		}
	}()

	return chMsgs, chErrors
}
