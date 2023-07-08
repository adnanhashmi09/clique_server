package ws

import (
	"context"
	"log"
	"time"

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

	new_channel_id, _ := gocql.RandomUUID()
	members := []string{req.Admin}
	created_at := time.Now()
	new_room_id, _ := gocql.RandomUUID()

	default_channel := &Channel{
		ID:              new_channel_id,
		ChannelName:     "general",
		Room:            new_room_id,
		Members:         members,
		CreatedAt:       created_at,
		IsDirectChannel: false,
	}

	new_room := &Room{
		ID:       new_room_id,
		RoomName: req.RoomName,
		Admin:    req.Admin,
		Channels: []gocql.UUID{default_channel.ID},
		Members: map[gocql.UUID][]string{
			default_channel.ID: members,
		},
		CreatedAt: created_at,
		ChannelMap: map[gocql.UUID]*Channel{
			default_channel.ID: default_channel,
		},
	}

	repo, err := s.REPOSITORY.CreateRoom(ctx, new_room, default_channel)

	if err != nil {
		return nil, err
	}

	return repo, nil

}

func (s *Service) JoinRoom(c context.Context, req *JoinOrLeaveRoomReq) (*Room, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	room, err := s.REPOSITORY.JoinRoom(ctx, req.ID, req.UserID, req.Username, req.Email)

	if err != nil {
		return nil, err
	}

	return room, nil

}

func (s *Service) LeaveRoom(c context.Context, req *JoinOrLeaveRoomReq) (*Room, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	room, err := s.REPOSITORY.LeaveRoom(ctx, req.ID, req.UserID, req.Username, req.Email)

	if err != nil {
		return nil, err
	}

	return room, nil

}

func (s *Service) DeleteRoom(c context.Context, req *DeleteRoomReq) (*Room, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	room, err := s.REPOSITORY.DeleteRoom(ctx, req.ID, req.UserID)

	if err != nil {
		return nil, err
	}

	return room, nil
}

func (s *Service) CreateChannel(c context.Context, req *CreateChannelReq) (*Room, *Channel, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	created_at := time.Now()
	new_channel_id, _ := gocql.RandomUUID()

	new_channel := Channel{
		ID:              new_channel_id,
		ChannelName:     req.ChannelName,
		Room:            req.RoomID,
		IsDirectChannel: false,
		CreatedAt:       created_at,
	}

	log.Println(req.Admin)
	room, chn, err := s.REPOSITORY.CreateChannel(ctx, &new_channel, req.Admin)

	if err != nil {
		return nil, nil, err
	}

	return room, chn, nil
}

func (s *Service) CreateDirectChannel(c context.Context, req *CreateDirectChannelReq) (gocql.UUID, *Channel, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	created_at := time.Now()
	new_channel_id, _ := gocql.RandomUUID()
	room_id := gocql.UUID{}
	admin := gocql.UUID{}
	members := []string{req.Sender, req.Reciever}

	new_channel := Channel{
		ID:              new_channel_id,
		ChannelName:     "",
		Room:            room_id,
		IsDirectChannel: true,
		CreatedAt:       created_at,
		Members:         members,
	}

	log.Println(admin)
	room_id, chn, err := s.REPOSITORY.CreateDirectChannel(ctx, &new_channel, admin, req.Reciever)

	if err != nil {
		return room_id, nil, err
	}

	return room_id, chn, nil
}

func (s *Service) DeleteChannel(c context.Context, req *DeleteChannelReq) (*Room, gocql.UUID, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	delete_channel := Channel{
		ID:              req.ChannelID,
		Room:            req.RoomID,
		IsDirectChannel: false,
	}

	log.Println(req.Admin)
	room, chn_id, err := s.REPOSITORY.DeleteChannel(ctx, &delete_channel, req.Admin)

	if err != nil {
		return nil, chn_id, err
	}

	return room, chn_id, nil
}

func (s *Service) WriteMessage(c context.Context, msg *Message) error {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	err := s.REPOSITORY.WriteMessageToDB(ctx, msg)

	if err != nil {
		return err
	}

	return nil
}

func (s *Service) CheckChannelMembership(c context.Context, join_channel_req *JoinChannelReq) (bool, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()
	return s.REPOSITORY.CheckChannelMembership(ctx, join_channel_req.Username, join_channel_req.RoomID, join_channel_req.ChannelID)
}

func (s *Service) CheckIfChannelExists(c context.Context, req *CreateDirectChannelReq) (*Channel, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()
	return s.REPOSITORY.CheckIfChannelExists(ctx, req)
}

func (s *Service) FetchAllMessages(c context.Context, chn_id gocql.UUID, user_id gocql.UUID, limit int, pg_state []byte) ([]Message, []byte, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()
	return s.REPOSITORY.FetchAllMessages(ctx, chn_id, user_id, limit, pg_state)
}

func (s *Service) GetAllRoomDetails(c context.Context, room_id gocql.UUID) (*RoomDetails, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	response, err := s.REPOSITORY.GetAllRoomDetails(ctx, room_id)

	if err != nil {
		return nil, err
	}

	return response, nil
}
