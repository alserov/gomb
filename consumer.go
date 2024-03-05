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

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"X-Request-ID": []string{p.ID}})
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

func (c *Consumer) Consume() chan Message {
	chMsgs := make(chan Message, 32)

	go func() {
		for {
			t, bytes, err := c.conn.ReadMessage()
			if err != nil {
				return
			}
			if t == websocket.CloseMessage {
				return
			}

			var msg Message
			err = json.Unmarshal(bytes, &msg)
			if err != nil {
				continue
			}

			chMsgs <- msg
		}
	}()

	return chMsgs
}

func (c *Consumer) Close() error {
	c.conn.WriteMessage(websocket.CloseMessage, nil)
	return c.conn.Close()
}
