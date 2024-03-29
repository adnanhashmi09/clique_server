package ws

import (
	"context"
	"log"
	"time"

	// . "github.com/adnanhashmi09/clique_server/internal/user"
	"github.com/gocql/gocql"
)

type Hub struct {
	// map of room id to room struct
	Rooms           map[gocql.UUID]*Room
	Register        chan *Client
	Unregister      chan *Client
	Broadcast       chan *Message
	NotifyTheSender chan *Message
}

type Room struct {
	ID         gocql.UUID              `json:"id"`
	RoomName   string                  `json:"room_name"`
	Channels   []gocql.UUID            `json:"channels"`
	Members    map[gocql.UUID][]string `json:"members"` // map of channel id to Members
	CreatedAt  time.Time               `json:"created_at"`
	Admin      string                  `json:"admin"`
	ChannelMap map[gocql.UUID]*Channel
}

type RoomDetails struct {
	ID           gocql.UUID              `json:"id"`
	RoomName     string                  `json:"room_name"`
	ChannelsList []gocql.UUID            `json:"channels_list"`
	Channels     []*Channel              `json:"channels"`
	Members      map[gocql.UUID][]string `json:"members"` // map of channel id to Members
	CreatedAt    time.Time               `json:"created_at"`
	Admin        string                  `json:"admin"`
}

type Channel struct {
	ID              gocql.UUID `json:"id"`
	ChannelName     string     `json:"channel_name"`
	Room            gocql.UUID `json:"room_id"`
	Members         []string   `json:"members"`
	IsDirectChannel bool       `json:"is_direct"`
	CreatedAt       time.Time  `json:"created_at"`

	Clients map[gocql.UUID]*Client
}

type CreateRoomReq struct {
	RoomName string `json:"room_name"`
	Admin    string `json:"admin"`
}

type JoinOrLeaveRoomReq struct {
	ID       gocql.UUID `json:"room_id"`
	UserID   gocql.UUID `json:"user_id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
}

type CreateChannelReq struct {
	RoomID      gocql.UUID `json:"room_id"`
	Admin       string     `json:"admin"`
	Username    string     `json:"username"`
	ChannelName string     `json:"channel_name"`
}

type CreateDirectChannelReq struct {
	Sender          string     `json:"sender"`
	Reciever        string     `json:"reciever"`
	ChannelID       gocql.UUID `json:"channel_id"`
	IsDirectChannel bool
}

type DeleteChannelReq struct {
	RoomID    gocql.UUID `json:"room_id"`
	Admin     string     `json:"admin"`
	ChannelID gocql.UUID `json:"channel_id"`
}

type DeleteRoomReq struct {
	ID       gocql.UUID `json:"room_id"`
	UserID   gocql.UUID `json:"user_id"`
	Username string     `json:"username"`
}

type JoinChannelReq struct {
	ChannelID       gocql.UUID `json:"channel_id"`
	RoomID          gocql.UUID `json:"room_id"`
	UserID          gocql.UUID `json:"user_id"`
	Username        string     `json:"username"`
	IsDirectChannel bool       `json:"is_direct_channel"`
}

type FetchRoomsReq struct {
	UserID   gocql.UUID `json:"user_id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
}

type FetchRoomsRes struct {
	DirectChannels []gocql.UUID `json:"direct_channels"`
	Rooms          []gocql.UUID `json:"rooms"`
	Channels       []gocql.UUID `json:"channels"`
}

type REPOSITORY interface {
	CreateRoom(ctx context.Context, room *Room, default_channel *Channel) (*Room, error)
	JoinRoom(ctx context.Context, room_id gocql.UUID, user_id gocql.UUID, username string, email string) (*Room, error)
	LeaveRoom(ctx context.Context, room_id gocql.UUID, user_id gocql.UUID, username string, email string) (*Room, error)
	DeleteRoom(ctx context.Context, room_id gocql.UUID, user_id gocql.UUID) (*Room, error)
	CreateChannel(ctx context.Context, new_channel *Channel, admin string) (*Room, *Channel, error)
	DeleteChannel(ctx context.Context, chn *Channel, admin string) (*Room, gocql.UUID, error)
	FetchAllRooms() (map[gocql.UUID]*Room, error)
	WriteMessageToDB(ctx context.Context, msg *Message) error
	CheckChannelMembership(ctx context.Context, username string, roomID, channelID gocql.UUID) (bool, error)
	CreateDirectChannel(ctx context.Context, new_channel *Channel, admin gocql.UUID, reciever string) (gocql.UUID, *Channel, error)
	CheckIfChannelExists(ctx context.Context, req *CreateDirectChannelReq) (*Channel, error)
	FetchAllMessages(ctx context.Context, chn_id gocql.UUID, user_id gocql.UUID, limit int, pg_state []byte) ([]Message, []byte, error)
	GetAllRoomDetails(ctx context.Context, room_id gocql.UUID) (*RoomDetails, error)
}

type SERVICE interface {
	CreateRoom(c context.Context, req *CreateRoomReq) (*Room, error)
	JoinRoom(c context.Context, req *JoinOrLeaveRoomReq) (*Room, error)
	LeaveRoom(c context.Context, req *JoinOrLeaveRoomReq) (*Room, error)
	DeleteRoom(c context.Context, req *DeleteRoomReq) (*Room, error)
	CreateChannel(c context.Context, req *CreateChannelReq) (*Room, *Channel, error)
	DeleteChannel(c context.Context, req *DeleteChannelReq) (*Room, gocql.UUID, error)
	WriteMessageToDB(c context.Context, msg *Message) error
	CheckChannelMembership(c context.Context, join_channel_req *JoinChannelReq) (bool, error)
	CreateDirectChannel(c context.Context, req *CreateDirectChannelReq) (gocql.UUID, *Channel, error)
	CheckIfChannelExists(c context.Context, req *CreateDirectChannelReq) (*Channel, error)
	FetchAllMessages(c context.Context, chn_id gocql.UUID, user_id gocql.UUID, limit int, pg_state []byte) ([]Message, []byte, error)
	GetAllRoomDetails(c context.Context, room_id gocql.UUID) (*RoomDetails, error)
}

func NewHub(repo REPOSITORY) *Hub {
	// fetch all the rooms and assgin them to a map RoodId -> Room
	rooms, err := repo.FetchAllRooms()
	if err != nil {
		log.Fatalf("Cannot fetch all rooms at initialisation. err: %v", err)
	}

	return &Hub{
		Rooms:           rooms,
		Register:        make(chan *Client),
		Unregister:      make(chan *Client),
		Broadcast:       make(chan *Message, 10), // TODO: Size of the buffer channel??
		NotifyTheSender: make(chan *Message, 10), // TODO: Size of the buffer channel??
	}
}

func (h *Hub) Run() {
	for {
		select {

		case cl := <-h.Register:
			log.Println("Client: ", cl)
			log.Println("hub room", h.Rooms)
			if room, ok := h.Rooms[cl.RoomID]; ok {
				if channel, ok := room.ChannelMap[cl.ChannelID]; ok {
					if _, ok := channel.Clients[cl.ID]; !ok {
						log.Println("Register")
						if channel.Clients == nil {
							channel.Clients = make(map[gocql.UUID]*Client)
						}
						channel.Clients[cl.ID] = cl
					}
				}
			}

		case cl := <-h.Unregister:
			if _, ok := h.Rooms[cl.RoomID]; ok {
				room := h.Rooms[cl.RoomID]
				if _, ok := room.ChannelMap[cl.ChannelID]; ok {
					log.Println("Unregister")
					channel := room.ChannelMap[cl.ChannelID]
					delete(channel.Clients, cl.ID)
					close(cl.Message)
				}
			}

		case msg := <-h.Broadcast:

			log.Println("------------------")
			log.Println("Broadcast")
			log.Println(msg)
			if _, ok := h.Rooms[msg.RoomID]; ok {
				room := h.Rooms[msg.RoomID]
				if _, ok := room.ChannelMap[msg.ChannelID]; ok {
					channel := room.ChannelMap[msg.ChannelID]
					log.Printf("%+v", channel.Clients)
					for _, cl := range channel.Clients {
						if cl.ID == msg.SenderID {
							continue
						}
						log.Println("Broadcast inside")
						log.Println(cl)
						cl.Message <- msg
					}
				}
			}
			log.Println("-------------------")

		case msg := <-h.NotifyTheSender:
			if _, ok := h.Rooms[msg.RoomID]; ok {
				room := h.Rooms[msg.RoomID]
				if _, ok := room.ChannelMap[msg.ChannelID]; ok {
					channel := room.ChannelMap[msg.ChannelID]
					channel.Clients[msg.SenderID].Message <- msg
				}
			}
		}
	}
}
