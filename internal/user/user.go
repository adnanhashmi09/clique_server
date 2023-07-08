package user

import (
	"context"
	"time"

	"github.com/gocql/gocql"
)

type User struct {
	ID                gocql.UUID   `json:"id"`
	Username          string       `json:"username"`
	Email             string       `json:"email"`
	Name              string       `json:"name"`
	Password          string       `json:"password"`
	Rooms             []gocql.UUID `json:"rooms"`
	DirectMsgChannels []gocql.UUID `json:"direct_msg_channel_id"`
}

type CreateUserReq struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Password string `json:"password"`
}

type CreateUserRes struct {
	ID       gocql.UUID `json:"id"`
	Username string     `json:"username"`
	Email    string     `json:"email"`
	Name     string     `json:"name"`
}

type LoginUserReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginUserRes struct {
	accessToken string
	ID          gocql.UUID `json:"id"`
	Username    string     `json:"username"`
	Email       string     `json:"email"`
	Name        string     `json:"name"`
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

type Channel struct {
	ID              gocql.UUID `json:"id"`
	ChannelName     string     `json:"channel_name"`
	Room            gocql.UUID `json:"room_id"`
	Members         []string   `json:"members"`
	IsDirectChannel bool       `json:"is_direct"`
	CreatedAt       time.Time  `json:"created_at"`
}

type UserAllInfo struct {
	DirectMsgChannels []*Channel `json:"direct_msg_channels"`
	Rooms             []*Room    `json:"rooms"`
	User              *User      `json:"user"`
}

type REPOSITORY interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	FetchAllInformation(ctx context.Context, user_id gocql.UUID, username, email string) ([]*Channel, []*Room, *User, error)
}

type SERVICE interface {
	CreateUser(c context.Context, req *CreateUserReq) (*CreateUserRes, error)
	Login(c context.Context, req *LoginUserReq) (*LoginUserRes, error)
	AllInfo(c context.Context, user_id gocql.UUID) (*UserAllInfo, error)
}
