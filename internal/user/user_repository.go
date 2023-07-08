package user

import (
	"context"
	"errors"
	"fmt"
	"log"

	"github.com/gocql/gocql"
)

type DBTX interface {
}

type Repository struct {
	db *gocql.Session
}

func NewRepository(db *gocql.Session) REPOSITORY {
	return &Repository{db: db}
}

func (r *Repository) CreateUser(ctx context.Context, user *User) (*User, error) {

	// Check if username already exists
	var uid gocql.UUID
	duplicate := true
	if err := r.db.Query(`
            SELECT id 
            FROM users 
            WHERE username=? 
            ALLOW FILTERING`,
		user.Username).WithContext(ctx).Scan(&uid); err != nil {

		if err != gocql.ErrNotFound {
			log.Println("Error occured while executing query to get if username exists or not in create user repository function:", err)
			return nil, errors.New("server error.")
		} else {
			duplicate = false
		}
	}

	if duplicate == true {
		return nil, errors.New("username already exists.")
	}

	// Check if email already exists
	duplicate = true
	var email_check string
	if err := r.db.Query(`
            SELECT ID 
            FROM users 
            WHERE email=? 
            ALLOW FILTERING`,
		user.Email).WithContext(ctx).Scan(&email_check); err != nil {

		if err != gocql.ErrNotFound {
			log.Println("Error occured while executing query to get if email exists or not in create user repository function:", err)
			return nil, errors.New("server error.")
		} else {
			duplicate = false
		}
	}

	if duplicate == true {
		return nil, errors.New("Email already exists.")
	}

	newId, _ := gocql.RandomUUID()

	if err := r.db.Query(`
            INSERT INTO 
            users (id, username, email, name, password) 
            VALUES (?, ?, ?, ?, ?)`,
		newId, user.Username, user.Email, user.Name, user.Password).WithContext(ctx).Exec(); err != nil {
		log.Fatal(err)
	}

	user.ID = newId
	return user, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (*User, error) {
	u := User{}
	scanner := r.db.Query(`
             SELECT 
             id, username, email, name, password 
             FROM users 
             WHERE email=? 
             ALLOW FILTERING`,
		email).WithContext(ctx).Iter().Scanner()
	scanner.Next()

	err := scanner.Scan(&u.ID, &u.Username, &u.Email, &u.Name, &u.Password)
	if err != nil {
		return nil, fmt.Errorf("Cannot scan: %v", err)
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("Error while closing the scanner: %v", err)
	}

	return &u, nil
}
func (r *Repository) FetchAllInformation(ctx context.Context, user_id gocql.UUID, username, email string) ([]*Channel, []*Room, *User, error) {
	query := `SELECT id, username, email, name, rooms, direct_msg_channels FROM users WHERE id=? AND username=? AND email=? LIMIT 1`

	var user User
	err := r.db.Query(query, user_id, username, email).WithContext(ctx).Scan(
		&user.ID, &user.Username, &user.Email, &user.Name, &user.Rooms, &user.DirectMsgChannels,
	)

	if err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil, nil, errors.New("user not found")
		}
		return nil, nil, nil, err
	}

	query = `SELECT id, channel_name, room_id, members, is_direct_channel, created_at FROM channels WHERE id IN ? ALLOW FILTERING`

	var channels []*Channel
	if len(user.DirectMsgChannels) != 0 {
		iter := r.db.Query(query, user.DirectMsgChannels).WithContext(ctx).Iter()
		channel := &Channel{}
		for iter.Scan(
			&channel.ID, &channel.ChannelName, &channel.Room, &channel.Members, &channel.IsDirectChannel, &channel.CreatedAt,
		) {
			channels = append(channels, channel)
			channel = &Channel{}
		}
		if err := iter.Close(); err != nil {
			return nil, nil, nil, err
		}
	}

	query = `SELECT id, room_name, channels, members, created_at, admin FROM rooms WHERE id IN ? ALLOW FILTERING`
	var rooms []*Room
	if len(user.Rooms) != 0 {
		iter := r.db.Query(query, user.Rooms).WithContext(ctx).Iter()
		room := &Room{}
		for iter.Scan(
			&room.ID, &room.RoomName, &room.Channels, &room.Members, &room.CreatedAt, &room.Admin,
		) {
			room.ChannelMap = make(map[gocql.UUID]*Channel)
			rooms = append(rooms, room)
			room = &Room{}
		}
		if err := iter.Close(); err != nil {
			return nil, nil, nil, err
		}
	}

	return channels, rooms, &user, nil
}
