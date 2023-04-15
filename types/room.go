package types

import "time"

type Room struct {
	ID       int       `json:"id"`
	RoomName string    `json:"room_name"`
	Channels []Channel `json:"channels"`
	Members  []User    `json:"members"`
}

type Channel struct {
	ID              int       `json:"id"`
	ChannelName     string    `json:"channel_name"`
	Members         []User    `json:"members"`
	Messages        []Message `json:"messages"`
	IsDirectChannel bool      `json:"is_direct"`
}

type Message struct {
	ID        int       `json:"id"`
	SenderId  int       `json:"sender_id"`
	Content   string    `json:"content"`
	Timestamp time.Time `json:"timestamp"`
}
