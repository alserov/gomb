package server

import "github.com/gorilla/websocket"

type Buffer chan []byte

type Consumer struct {
	id   string
	conn *websocket.Conn
}
