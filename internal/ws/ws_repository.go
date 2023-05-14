package ws

import (
	"context"
	"errors"
	"log"

	"github.com/gocql/gocql"
)

type Repository struct {
	db *gocql.Session
}

func NewRepository(db *gocql.Session) REPOSITORY {
	return &Repository{db: db}
}

func (r *Repository) CreateRoom(ctx context.Context, room *Room, default_channel *Channel) (*Room, error) {

	// Check if admin exists or not
	// TODO: Remove allow filtering if possible
	var (
		uid            gocql.UUID
		admin_username string
		admin_email    string
	)

	if err := r.db.Query(`SELECT id, username, email FROM users WHERE id=? ALLOW FILTERING`,
		room.Admin).WithContext(ctx).Scan(&uid, &admin_username, &admin_email); err != nil {
		if err != gocql.ErrNotFound {
			log.Println("Error occured while executing query to get if admin exists or not in create room repository function:", err)
			return nil, errors.New("Server error.")
		} else {
			return nil, errors.New("Admin id doesn't exist.")
		}
	}

	// Check if room name already exists for that admin
	var room_id gocql.UUID

	if err := r.db.Query(`SELECT id FROM rooms WHERE admin=? AND room_name=? ALLOW FILTERING`,
		room.Admin, room.RoomName).WithContext(ctx).Scan(&room_id); err != nil {
		if err != gocql.ErrNotFound {
			log.Println("Error occured while executing query to get if admin exists or not in create room repository function:", err)
			return nil, errors.New("Server error.")
		}
	}

	if room_id.String() != "" {
		log.Println("Room id: ", room_id)
		return nil, errors.New("Room already exists.")
	}

	// Insert into rooms table
	if err := r.db.Query(`INSERT INTO rooms (id, room_name, channels, members, admin, created_at)
                       VALUES (?, ?, ?, ?, ?, ?)`, room.ID, room.RoomName, room.Channels,
		room.Members, room.Admin, room.CreatedAt).WithContext(ctx).Exec(); err != nil {
		log.Println("Error occured while executing query to create room. ", err)
		return nil, errors.New("Admin id doesn't exist.")
	}

	// Insert the default_channel into channels table
	if err := r.db.Query(`INSERT INTO channels (id, channel_name, is_direct_channel, members, created_at, messages)
                       VALUES (?, ?, ?, ?, ?, ?)`, default_channel.ID, default_channel.ChannelName, default_channel.IsDirectChannel,
		default_channel.Members, default_channel.CreatedAt, default_channel.Messages).WithContext(ctx).Exec(); err != nil {
		log.Println("Error occured while executing query to create default channel.", err)
		return nil, errors.New("Admin id doesn't exist.")
	}

	// Add the room in users table
	if err := r.db.Query(`UPDATE users SET rooms = rooms + {?} where id=? AND username=? AND email=?`,
		room.ID, room.Admin, admin_username, admin_email).WithContext(ctx).Exec(); err != nil {
		log.Println("Error occured while executing query update user room set to include newly created room. ", err)
		return nil, errors.New("Admin id doesn't exist.")
	}

	return room, nil
}

// TODO: FINISH this function
func (r *Repository) JoinRoom(ctx context.Context, room_id gocql.UUID, user_id gocql.UUID) (*Room, error) {

	// check if user doesn't exist
	var uid gocql.UUID
	if err := r.db.Query(`SELECT ID FROM users WHERE id=?`, user_id).WithContext(ctx).Scan(&uid); err != nil {
		log.Println("Error occured while executing query to get if user exists or not in join room repository function:", err)
		return nil, errors.New("User id doesn't exist.")
	}

	// check if room doesn't exist
	var rid gocql.UUID
	if err := r.db.Query(`SELECT ID FROM rooms WHERE id=?`, room_id).WithContext(ctx).Scan(&rid); err != nil {
		log.Println("Error occured while executing query to get if room exists or not in join room repository function:", err)
		return nil, errors.New("Room id doesn't exist.")
	}

	// check if user is already a member of room

	// Add user to the room

	return nil, nil
}
