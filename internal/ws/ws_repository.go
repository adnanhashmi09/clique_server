package ws

import (
	"context"
	"errors"
	"log"
	"time"

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

	if err := r.db.Query(`
            SELECT 
            id, username, email 
            FROM users 
            WHERE id=? 
            ALLOW FILTERING`,
		room.Admin).WithContext(ctx).Scan(&uid, &admin_username, &admin_email); err != nil {
		if err != gocql.ErrNotFound {
			log.Println("Error occured while executing query to get if admin exists or not in create room repository function:", err)
			return nil, errors.New("Server error.")
		} else {
			return nil, errors.New("Admin id doesn't exist.")
		}
	}

	// Check if room name already exists for that admin
	var room_id string

	if err := r.db.Query(`
            SELECT 
            id 
            FROM rooms 
            WHERE admin=? AND room_name=? 
            ALLOW FILTERING`,
		room.Admin, room.RoomName).WithContext(ctx).Scan(&room_id); err != nil {
		if err != gocql.ErrNotFound {
			log.Println("Error occured while executing query to get if admin exists or not in create room repository function:", err)
			return nil, errors.New("Server error.")
		}
	}

	if room_id != "" {
		log.Println("Room id: ", room_id)
		return nil, errors.New("Room already exists.")
	}

	batch := r.db.NewBatch(gocql.LoggedBatch).WithContext(ctx)
	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `INSERT 
           INTO rooms 
           (id, room_name, channels, members, admin, created_at) 
           VALUES (?, ?, ?, ?, ?, ?)`,
		Args:       []interface{}{room.ID, room.RoomName, room.Channels, room.Members, room.Admin, room.CreatedAt},
		Idempotent: true,
	})

	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `INSERT 
           INTO channels 
           (id, channel_name, room_id, is_direct_channel, members, created_at, messages) 
           VALUES (?, ?, ?, ?, ?, ?, ?)`,
		Args: []interface{}{default_channel.ID, default_channel.ChannelName, default_channel.Room, default_channel.IsDirectChannel,
			default_channel.Members, default_channel.CreatedAt, default_channel.Messages},
		Idempotent: true,
	})

	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `UPDATE users 
           SET rooms = rooms + {?} 
           where id=? AND username=? AND email=?`,
		Args:       []interface{}{room.ID, room.Admin, admin_username, admin_email},
		Idempotent: true,
	})

	err := r.db.ExecuteBatch(batch)
	if err != nil {
		log.Println("Error executing batch statement in create room: ", err)
		return nil, errors.New("Server Error.")
	}

	return room, nil
}

// TODO: FINISH this function
func (r *Repository) JoinRoom(ctx context.Context, room_id gocql.UUID, user_id gocql.UUID, username string, email string) (*Room, error) {

	// check if user doesn't exist
	var uid gocql.UUID
	if err := r.db.Query(`
            SELECT 
            id 
            FROM users 
            WHERE id=?
            ALLOW FILTERING`,
		user_id).WithContext(ctx).Scan(&uid); err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.New("User doesn't exist.")
		} else {
			log.Println("Error occured while executing query to get if user exists or not in join room repository function:", err)
			return nil, errors.New("Server error.")
		}
	}

	// check if room doesn't exist
	var room Room
	if err := r.db.Query(`
            SELECT 
            * 
            FROM rooms 
            WHERE id=?
            ALLOW FILTERING`,
		room_id).WithContext(ctx).Scan(&room.ID, &room.Admin, &room.CreatedAt, &room.Channels, &room.Members, &room.RoomName); err != nil {

		if err == gocql.ErrNotFound {
			return nil, errors.New("Room id doesn't exist.")
		} else {
			log.Println("Error occured while executing query to get if room exists or not in join room repository function:", err)
			return nil, errors.New("Server error.")
		}
	}

	// check if user is already a member of room
	var uid_mem string
	if err := r.db.Query(`
            SELECT 
            id 
            FROM users 
            WHERE id=?
            AND
            rooms CONTAINS ?
            ALLOW FILTERING`,
		user_id, room_id).WithContext(ctx).Scan(&uid_mem); err != nil {

		if err != gocql.ErrNotFound {
			log.Println("Error occured while executing query to check if user is already in the room in join room repository function:", err)
			return nil, errors.New("Server error.")
		}

	}

	if uid_mem != "" {
		return nil, errors.New("User is already a member of the room.")
	}

	// Get the channel id of general channel
	var (
		chn_id             gocql.UUID
		channel_created_at time.Time
	)
	if err := r.db.Query(`
            SELECT 
            id, created_at 
            FROM channels 
            WHERE room_id=?
            AND
            channel_name='general'
            ALLOW FILTERING`,
		room_id).WithContext(ctx).Scan(&chn_id, &channel_created_at); err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.New("Channel id doesn't exist.")
		} else {
			log.Println("Error occured while executing query to get general channel id in join room repository function:", err)
			return nil, errors.New("Server error.")
		}
	}

	// Add user to the room
	// 1. Add user to general channel -> Update channel table
	batch := r.db.NewBatch(gocql.LoggedBatch)

	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `UPDATE channels 
           SET members = members + {?} 
           where room_id=? 
           AND id=?
           AND created_at=?`,
		Args:       []interface{}{user_id, room_id, chn_id, channel_created_at},
		Idempotent: true,
	})

	// 2. Update Rooms table
	room.Members[chn_id] = append(room.Members[chn_id], user_id)

	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `UPDATE rooms 
           SET members=? 
           where id=? 
           AND admin=?
           AND created_at=?`,
		Args:       []interface{}{room.Members, room_id, room.Admin, room.CreatedAt},
		Idempotent: true,
	})

	room.Members[chn_id] = append(room.Members[chn_id], user_id)

	// 3. Update users table
	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `UPDATE users 
           SET rooms = rooms + {?} 
           where id=?
           AND username=?
           AND email=?`,
		Args:       []interface{}{room_id, user_id, username, email},
		Idempotent: true,
	})

	err := r.db.ExecuteBatch(batch)
	if err != nil {
		log.Println("Error executing batch statement in join room: ", err)
		return nil, errors.New("Server Error.")
	}

	return &room, nil
}
