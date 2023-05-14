package user

import (
	"context"

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

type REPOSITORY interface {
	CreateUser(ctx context.Context, user *User) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
}

type SERVICE interface {
	CreateUser(c context.Context, req *CreateUserReq) (*CreateUserRes, error)
	Login(c context.Context, req *LoginUserReq) (*LoginUserRes, error)
}
