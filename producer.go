package gomb

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

const (
	DEFAULT_RETRY_AMOUNT = 3
)

type Producer struct {
	Topic string
	Addr  string
	ID    string

	mu           *sync.Mutex
	retryTimeout time.Duration
	retries      int
	conn         *websocket.Conn
}

type Params struct {
	Addr  string
	Topic string
	ID    string

	RetryAmount  int
	RetryTimeout time.Duration
}

func NewProducer(p Params) (*Producer, error) {
	u := url.URL{Scheme: "ws", Host: p.Addr, Path: fmt.Sprintf("/produce/%s", p.Topic)}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"X-Producer-ID": []string{p.ID}})
	if err != nil {
		return nil, fmt.Errorf("failed to server with server: %w", err)
	}

	retries := p.RetryAmount
	if retries <= 0 {
		retries = DEFAULT_RETRY_AMOUNT
	}

	return &Producer{
		Topic:        p.Topic,
		ID:           p.ID,
		Addr:         p.Addr,
		conn:         conn,
		retries:      retries,
		retryTimeout: p.RetryTimeout,
		mu:           &sync.Mutex{},
	}, nil
}

func (p *Producer) Reconnect() error {
	u := url.URL{Scheme: "ws", Host: p.Addr, Path: fmt.Sprintf("/%s", p.Topic)}

	retries := p.retries
	var e error

	for retries > 0 {
		conn, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{"X-Producer-ID": []string{p.ID}})
		if err == nil {
			p.conn = conn
			return nil
		}

		sync.OnceFunc(func() {
			e = err
		})

		retries--
	}

	return fmt.Errorf("failed to reconnect to server: %w", e)
}

func (p *Producer) Produce(m *Message) error {
	m.ProducerID = p.ID
	m.ID = strconv.Itoa(time.Now().Nanosecond())

	b, err := json.Marshal(m)
	if err != nil {
		return fmt.Errorf("failed to encode message: %w", err)
	}

	p.mu.Lock()
	defer p.mu.Unlock()

	if err = p.conn.WriteMessage(websocket.BinaryMessage, b); err != nil {
		return fmt.Errorf("failed to send message: %w", err)
	}

	return nil
}

func (p *Producer) Close() error {
	if err := p.conn.Close(); err != nil {
		return fmt.Errorf("failed to close connection: %w", err)
	}
	return nil
}
