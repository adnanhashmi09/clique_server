package ws

import (
	"context"
	"time"

	"github.com/adnanhashmi09/clique_server/internal/user"
	"github.com/gocql/gocql"
)

type Service struct {
	REPOSITORY
	timeout time.Duration
}

func NewService(repository REPOSITORY) SERVICE {
	return &Service{
		repository,
		time.Duration(2) * time.Second,
	}
}

func (s *Service) CreateRoom(c context.Context, req *CreateRoomReq) (*Room, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	admin_user := &user.User{
		ID: req.Admin,
	}

	new_channel_id, _ := gocql.RandomUUID()
	members := []gocql.UUID{(gocql.UUID)(admin_user.ID)}
	created_at := time.Now()

	default_channel := &Channel{
		ID:              new_channel_id,
		ChannelName:     "general",
		Members:         members,
		CreatedAt:       created_at,
		IsDirectChannel: false,
		Messages:        []int{},
	}

	new_room_id, _ := gocql.RandomUUID()
	new_room := &Room{
		ID:       new_room_id,
		RoomName: req.RoomName,
		Admin:    req.Admin,
		Channels: []gocql.UUID{default_channel.ID},
		Members: map[gocql.UUID][]gocql.UUID{
			default_channel.ID: members,
		},
		CreatedAt: created_at,
	}

	repo, err := s.REPOSITORY.CreateRoom(ctx, new_room, default_channel)

	if err != nil {
		return nil, err
	}

	return repo, nil

}

func (s *Service) JoinRoom(c context.Context, req *JoinRoomReq) (*Room, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	repo, err := s.REPOSITORY.JoinRoom(ctx, req.ID, req.User)

	if err != nil {
		return nil, err
	}

	return repo, nil

}