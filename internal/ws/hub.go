package ws

import (
	"context"
	"time"

	// . "github.com/adnanhashmi09/clique_server/internal/user"
	"github.com/gocql/gocql"
)

type Hub struct {
	// map of room id to room struct
	Rooms map[gocql.UUID]*Room
}

type Room struct {
	ID        gocql.UUID                  `json:"id"`
	RoomName  string                      `json:"room_name"`
	Channels  []gocql.UUID                `json:"channels"`
	Members   map[gocql.UUID][]gocql.UUID `json:"members"` // map of channel id to Members
	CreatedAt time.Time                   `json:"created_at"`
	Admin     gocql.UUID                  `json:"admin"`
}

type Channel struct {
	ID              gocql.UUID   `json:"id"`
	ChannelName     string       `json:"channel_name"`
	Room            gocql.UUID   `json:"room_id"`
	Members         []gocql.UUID `json:"members"`
	IsDirectChannel bool         `json:"is_direct"`
	Messages        []int        `json:"messages"`
	CreatedAt       time.Time    `json:"created_at"`

	Clients map[gocql.UUID]*Client
}

type CreateRoomReq struct {
	RoomName string     `json:"room_name"`
	Admin    gocql.UUID `json:"admin"`
}

type JoinRoomReq struct {
	ID       gocql.UUID `json:"room_id"`
	UserID   gocql.UUID `json:"user_id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
}

type REPOSITORY interface {
	CreateRoom(ctx context.Context, room *Room, default_channel *Channel) (*Room, error)
	JoinRoom(ctx context.Context, room_id gocql.UUID, user_id gocql.UUID, username string, email string) (*Room, error)
}

type SERVICE interface {
	CreateRoom(c context.Context, req *CreateRoomReq) (*Room, error)
	JoinRoom(c context.Context, req *JoinRoomReq) (*Room, error)
}

func NewHub() *Hub {
	return &Hub{
		Rooms: make(map[gocql.UUID]*Room),
	}
}
