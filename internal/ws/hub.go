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

type JoinOrLeaveRoomReq struct {
	ID       gocql.UUID `json:"room_id"`
	UserID   gocql.UUID `json:"user_id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
}

type CreateChannelReq struct {
	RoomID      gocql.UUID `json:"room_id"`
	Admin       gocql.UUID `json:"admin"`
	ChannelName string     `json:"channel_name"`
}

type DeleteChannelReq struct {
	RoomID    gocql.UUID `json:"room_id"`
	Admin     gocql.UUID `json:"admin"`
	ChannelID gocql.UUID `json:"channel_id"`
}

type DeleteRoomReq struct {
	ID     gocql.UUID `json:"room_id"`
	UserID gocql.UUID `json:"user_id"`
}

type REPOSITORY interface {
	CreateRoom(ctx context.Context, room *Room, default_channel *Channel) (*Room, error)
	JoinRoom(ctx context.Context, room_id gocql.UUID, user_id gocql.UUID, username string, email string) (*Room, error)
	LeaveRoom(ctx context.Context, room_id gocql.UUID, user_id gocql.UUID, username string, email string) (*Room, error)
	DeleteRoom(ctx context.Context, room_id gocql.UUID, user_id gocql.UUID) (*Room, error)
	CreateChannel(ctx context.Context, new_channel *Channel, admin gocql.UUID) (*Room, error)
	DeleteChannel(ctx context.Context, chn *Channel, admin gocql.UUID) (*Room, error)
}

type SERVICE interface {
	CreateRoom(c context.Context, req *CreateRoomReq) (*Room, error)
	JoinRoom(c context.Context, req *JoinOrLeaveRoomReq) (*Room, error)
	LeaveRoom(c context.Context, req *JoinOrLeaveRoomReq) (*Room, error)
	DeleteRoom(c context.Context, req *DeleteRoomReq) (*Room, error)
	CreateChannel(c context.Context, req *CreateChannelReq) (*Room, error)
	DeleteChannel(c context.Context, req *DeleteChannelReq) (*Room, error)
}

func NewHub() *Hub {
	return &Hub{
		Rooms: make(map[gocql.UUID]*Room),
	}
}
