package ws

import (
	"time"

	"github.com/gocql/gocql"
	"golang.org/x/net/websocket"
)

type Client struct {
	Conn    *websocket.Conn
	Message chan *Message
	ID      gocql.UUID `json:"id"`
	RoomID  gocql.UUID `json:"room_id"`
}

type Message struct {
	ID        int        `json:"id"`
	SenderId  gocql.UUID `json:"sender_id"`
	ChannelId gocql.UUID `json:"channel_id"`
	Content   string     `json:"content"`
	Timestamp time.Time  `json:"timestamp"`
}
