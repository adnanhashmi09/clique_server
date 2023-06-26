package ws

import (
	"context"
	"log"
	"time"

	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
)

type Client struct {
	Conn      *websocket.Conn
	Message   chan *Message
	ID        gocql.UUID `json:"id"`
	RoomID    gocql.UUID `json:"room_id"`
	ChannelID gocql.UUID `json:"channel_id"`
	Username  string     `json:"username"`
}

type Message struct {
	RoomID    gocql.UUID `json:"room_id"`
	SenderID  gocql.UUID `json:"sender_id"`
	ChannelID gocql.UUID `json:"channel_id"`
	Content   string     `json:"content"`
	Timestamp time.Time  `json:"timestamp"`
	Type      string     `json:"type"`
}

func (cl *Client) writeMessage() {
	log.Println("write message function")
	defer func() {
		cl.Conn.Close()
	}()

	for {
		message, ok := <-cl.Message
		if !ok {
			return
		}

		cl.Conn.WriteJSON(message)
	}
}

func (cl *Client) readMessage(ctx context.Context, h *Handler) {
	defer func() {
		h.Hub.Unregister <- cl
		cl.Conn.Close()
	}()

	log.Println("read message function")

	for {
		_, m, err := cl.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error in websocket hub readMessage function: %v", err)
			}

			break
		}

		msg := &Message{
			RoomID:    cl.RoomID,
			SenderID:  cl.ID,
			ChannelID: cl.ChannelID,
			Content:   string(m),
			Timestamp: time.Now(),
			Type:      "text_message",
		}

		// write msg to DB
		err = h.WriteMessageToDB(ctx, msg)
		if err != nil {
			log.Println("Error sending message to db. err: ", err)
			msg = &Message{
				RoomID:    cl.RoomID,
				SenderID:  cl.ID,
				ChannelID: cl.ChannelID,
				Content:   string("Cannot send this message since this message wasn't saved."),
				Timestamp: time.Now(),
				Type:      "error",
			}

			h.Hub.NotifyTheSender <- msg
			continue
		}

		h.Hub.Broadcast <- msg
	}
}
