package ws

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/adnanhashmi09/clique_server/internal/user"
	"github.com/adnanhashmi09/clique_server/utils"
	"github.com/gocql/gocql"
)

type Repository struct {
	db *gocql.Session
}

func NewRepository(db *gocql.Session) REPOSITORY {
	return &Repository{db: db}
}

// This function will be used to initialize the Rooms map in Hub
func (r *Repository) FetchAllRooms() (map[gocql.UUID]*Room, error) {
	roomQuery := "SELECT id, room_name, channels, members, created_at, admin FROM rooms"
	roomIter := r.db.Query(roomQuery).Iter()

	roomMap := make(map[gocql.UUID]*Room)

	var roomID gocql.UUID
	var roomName string
	var channels []gocql.UUID
	var members map[gocql.UUID][]gocql.UUID
	var createdAt time.Time
	var admin gocql.UUID

	for roomIter.Scan(&roomID, &roomName, &channels, &members, &createdAt, &admin) {
		room := &Room{
			ID:         roomID,
			RoomName:   roomName,
			Channels:   channels,
			Members:    members,
			CreatedAt:  createdAt,
			Admin:      admin,
			ChannelMap: make(map[gocql.UUID]*Channel),
		}

		roomMap[roomID] = room
	}

	if err := roomIter.Close(); err != nil {
		log.Println("err here")
		return nil, err
	}

	// Fetch channels for each room and populate ChannelMap
	channelQuery := "SELECT id, channel_name, room_id, members, is_direct_channel, created_at FROM channels WHERE room_id = ? ALLOW FILTERING"
	channelStmt := r.db.Query(channelQuery)

	for roomID, room := range roomMap {
		channelIter := channelStmt.Bind(roomID).Iter()

		var channelID gocql.UUID
		var channelName string
		var roomID gocql.UUID
		var channelMembers []gocql.UUID
		var isDirectChannel bool
		var createdAt time.Time

		for channelIter.Scan(&channelID, &channelName, &roomID, &channelMembers, &isDirectChannel, &createdAt) {
			channel := &Channel{
				ID:              channelID,
				ChannelName:     channelName,
				Room:            roomID,
				Members:         channelMembers,
				IsDirectChannel: isDirectChannel,
				CreatedAt:       createdAt,
				Clients:         make(map[gocql.UUID]*Client),
			}

			room.ChannelMap[channelID] = channel
		}

		if err := channelIter.Close(); err != nil {
			return nil, err
		}
	}

	return roomMap, nil
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
           (id, channel_name, room_id, is_direct_channel, members, created_at) 
           VALUES (?, ?, ?, ?, ?, ?)`,
		Args: []interface{}{default_channel.ID, default_channel.ChannelName, default_channel.Room, default_channel.IsDirectChannel,
			default_channel.Members, default_channel.CreatedAt},
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

	// Get the channel id and created_at of all channels in the room
	var (
		chn_id              gocql.UUID
		channel_created_at  time.Time
		channels_created_at map[gocql.UUID]time.Time
	)

	channels_created_at = make(map[gocql.UUID]time.Time)

	scanner := r.db.Query(`
            SELECT 
            id, created_at 
            FROM channels 
            WHERE room_id=?
            ALLOW FILTERING`,
		room_id).WithContext(ctx).Iter().Scanner()

	for scanner.Next() {
		err := scanner.Scan(&chn_id, &channel_created_at)
		if err != nil {
			log.Println("Error occured while executing query to get general channel id in join room repository function:", err)
			return nil, errors.New("Server error.")
		}
		channels_created_at[chn_id] = channel_created_at
	}

	// Add user to the room
	batch := r.db.NewBatch(gocql.LoggedBatch)

	// 1. Add user to all the channels -> Update channel table
	for key, chn := range room.Members {
		chn = append(chn, user_id)
		room.Members[key] = chn

		batch.Entries = append(batch.Entries, gocql.BatchEntry{
			Stmt: `UPDATE channels 
             SET members = members + {?} 
             where room_id=? 
             AND id=?
             AND created_at=?`,
			Args:       []interface{}{user_id, room_id, key, channels_created_at[key]},
			Idempotent: true,
		})
	}

	// 2. Update Rooms table
	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `UPDATE rooms 
           SET members=? 
           where id=? 
           AND admin=?
           AND created_at=?`,
		Args:       []interface{}{room.Members, room_id, room.Admin, room.CreatedAt},
		Idempotent: true,
	})

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

func (r *Repository) LeaveRoom(ctx context.Context, room_id gocql.UUID, user_id gocql.UUID, username string, email string) (*Room, error) {

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
			log.Println("Error occured while executing query to get if user exists or not in leave room repository function:", err)
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
			log.Println("Error occured while executing query to get if room exists or not in leave room repository function:", err)
			return nil, errors.New("Server error.")
		}
	}

	// check if user is`not a member of room
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
			log.Println("Error occured while executing query to check if user is already in the room in leave room repository function:", err)
			return nil, errors.New("Server error.")
		} else {
			return nil, errors.New("User is not a member of the room.")
		}
	}

	// Get channels data
	var chn Channel
	var channels []Channel
	scanner := r.db.Query(`
             SELECT 
             room_id, id, created_at, members 
             FROM channels 
             WHERE room_id=?
             ALLOW FILTERING    
    `, room_id).WithContext(ctx).Iter().Scanner()

	for scanner.Next() {
		err := scanner.Scan(&chn.Room, &chn.ID, &chn.CreatedAt, &chn.Members)
		if err != nil {
			log.Println("Error while scanning for all the channels in Leave room repository function", err)
			return nil, errors.New("Server error.")
		}

		channels = append(channels, chn)
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error while closing the scanner. ", err)
		return nil, errors.New("Server Error")
	}

	// Remove user from the room
	batch := gocql.NewBatch(gocql.LoggedBatch)

	for _, channel := range channels {
		// 1. Remove user from all the channels in the room
		batch.Entries = append(batch.Entries, gocql.BatchEntry{
			Stmt: `UPDATE channels
	           SET members = members - {?}
	           where room_id=?
	           AND id=?
	           AND created_at=?`,
			Args:       []interface{}{user_id, room_id, channel.ID, channel.CreatedAt},
			Idempotent: true,
		})

		// 2. Remove user from the room  table
		room.Members[channel.ID] = utils.RemoveElementFromArray(room.Members[channel.ID], user_id)
		batch.Entries = append(batch.Entries, gocql.BatchEntry{
			Stmt: `UPDATE rooms
             SET members[?] = ?
	           where id=?
	           AND admin=?
	           AND created_at=?`,
			Args:       []interface{}{channel.ID, room.Members[channel.ID], room_id, room.Admin, room.CreatedAt},
			Idempotent: true,
		})
	}

	// 3. Remove room from the users table
	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `UPDATE users
	         SET rooms = rooms - {?}
	         where id=?
	         AND username=?
	         AND email=?`,
		Args:       []interface{}{room_id, user_id, username, email},
		Idempotent: true,
	})

	err := r.db.ExecuteBatch(batch)
	if err != nil {
		log.Println("Error executing batch statement in leave room: ", err)
		return nil, errors.New("Server Error.")
	}

	return &room, nil
}

func (r *Repository) DeleteRoom(ctx context.Context, room_id gocql.UUID, user_id gocql.UUID) (*Room, error) {
	// check if user is the admin
	// this also checks if room exists or not
	var (
		uid             gocql.UUID
		room_created_at time.Time
	)
	if err := r.db.Query(`
            SELECT 
            admin, created_at 
            FROM rooms 
            WHERE id=?
            ALLOW FILTERING`,
		room_id).WithContext(ctx).Scan(&uid, &room_created_at); err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.New("Room doesn't exist.")
		} else {
			log.Println("Error occured while executing query to get if user is admin or not in delete room repository function:", err)
			return nil, errors.New("Server error.")
		}
	}

	if uid != user_id {
		return nil, errors.New("User is not the admin of this room.")
	}

	// Get channels data
	var chn Channel
	var channels []Channel
	scanner := r.db.Query(`
             SELECT 
             room_id, id, created_at, members 
             FROM channels 
             WHERE room_id=?
             ALLOW FILTERING    
    `, room_id).WithContext(ctx).Iter().Scanner()

	for scanner.Next() {
		err := scanner.Scan(&chn.Room, &chn.ID, &chn.CreatedAt, &chn.Members)
		if err != nil {
			log.Println("Error while scanning for all the channels in delete room repository function", err)
			return nil, errors.New("Server error.")
		}

		channels = append(channels, chn)
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error while closing the scanner. ", err)
		return nil, errors.New("Server Error")
	}

	// Get users data
	var usr user.User
	var users []user.User
	scanner = r.db.Query(`
             SELECT 
             id, username, email, rooms 
             FROM users 
             WHERE rooms CONTAINS ? 
             ALLOW FILTERING    
    `, room_id).WithContext(ctx).Iter().Scanner()

	for scanner.Next() {
		err := scanner.Scan(&usr.ID, &usr.Username, &usr.Email, &usr.Rooms)
		if err != nil {
			log.Println("Error while scanning for all the users in a room in delete room repository function", err)
			return nil, errors.New("Server error.")
		}

		users = append(users, usr)
	}

	if err := scanner.Err(); err != nil {
		log.Println("Error while closing the scanner. ", err)
		return nil, errors.New("Server Error")
	}

	batch := gocql.NewBatch(gocql.LoggedBatch)

	// Delete all the channels belonging to the room
	for _, channel := range channels {
		batch.Entries = append(batch.Entries, gocql.BatchEntry{
			Stmt: `DELETE 
             FROM channels
	           where room_id=?
	           AND id=?
	           AND created_at=?`,
			Args:       []interface{}{room_id, channel.ID, channel.CreatedAt},
			Idempotent: true,
		})
	}

	// Delete the room from rooms table
	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `DELETE 
           FROM rooms
	         where id=?
	         AND admin=?
	         AND created_at=?`,
		Args:       []interface{}{room_id, user_id, room_created_at},
		Idempotent: true,
	})

	// Delete rooms from the users table -> Update the users rooms column
	for _, u := range users {
		batch.Entries = append(batch.Entries, gocql.BatchEntry{
			Stmt: `UPDATE users 
             SET rooms = rooms - {?} 
             where id=?
             AND username=?
             AND email=?`,
			Args:       []interface{}{room_id, u.ID, u.Username, u.Email},
			Idempotent: true,
		})
	}

	err := r.db.ExecuteBatch(batch)
	if err != nil {
		log.Println("Error executing batch statement in delete room: ", err)
		return nil, errors.New("Server Error.")
	}

	return &Room{ID: room_id}, nil
}

func (r *Repository) CreateChannel(ctx context.Context, new_channel *Channel, admin gocql.UUID) (*Room, error) {

	// check if room exists and
	// admin given is actually the
	// admin of the room
	var room Room

	if err := r.db.Query(`
            SELECT 
            * 
            FROM rooms 
            WHERE id=?
            ALLOW FILTERING`,
		new_channel.Room).WithContext(ctx).Scan(&room.ID, &room.Admin, &room.CreatedAt, &room.Channels, &room.Members, &room.RoomName); err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.New("Room doesn't exist.")
		} else {
			log.Println("Error occured while executing query to get if user is admin or not in create channel repository function:", err)
			return nil, errors.New("Server error.")
		}
	}

	if room.Admin != admin {
		return nil, errors.New("User is not the admin of this room.")
	}

	// update members in new_channel with the members of any channel in the room
	for _, value := range room.Members {
		new_channel.Members = value
		break
	}

	batch := gocql.NewBatch(gocql.LoggedBatch)

	// create a channel in the channels table
	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `INSERT 
           INTO channels 
           (id, channel_name, room_id, is_direct_channel, members, created_at) 
           VALUES (?, ?, ?, ?, ?, ?)`,
		Args: []interface{}{new_channel.ID, new_channel.ChannelName, new_channel.Room, new_channel.IsDirectChannel,
			new_channel.Members, new_channel.CreatedAt},
		Idempotent: true,
	})

	// update the channels set and members map
	// in rooms table
	room.Channels = append(room.Channels, new_channel.ID)
	room.Members[new_channel.ID] = new_channel.Members
	room.ChannelMap[new_channel.ID] = new_channel

	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `UPDATE rooms
           SET channels = channels + {?}, 
           members=?
           where id=?
           AND admin=?
           AND created_at=?`,
		Args:       []interface{}{new_channel.ID, room.Members, room.ID, room.Admin, room.CreatedAt},
		Idempotent: true,
	})

	err := r.db.ExecuteBatch(batch)
	if err != nil {
		log.Println("Error executing batch statement in create channel: ", err)
		return nil, errors.New("Server Error.")
	}

	return &room, nil
}

func (r *Repository) CreateDirectChannel(ctx context.Context, new_channel *Channel, admin gocql.UUID, reciever string) (gocql.UUID, *Channel, error) {
	room_id := gocql.UUID{}
	var created_at time.Time

	if err := r.db.Query(`
            SELECT 
            created_at 
            FROM rooms 
            WHERE id=?
            ALLOW FILTERING`,
		room_id).WithContext(ctx).Scan(&created_at); err != nil {
		if err == gocql.ErrNotFound {
			return gocql.UUID{}, nil, errors.New("Room doesn't exist.")
		} else {
			log.Println("Error occured while trying to get created_at in createDirectChannel repo function.", err)
			return gocql.UUID{}, nil, errors.New("Server error.")
		}
	}

	// ceck whether sender id is legit or not
	var sender_id gocql.UUID
	var sender_username string
	var sender_email string

	if err := r.db.Query(`
           SELECT
           id, username, email
           FROM users
           where id=?
           ALLOW FILTERING
    `, new_channel.Members[0]).WithContext(ctx).Scan(&sender_id, &sender_username, &sender_email); err != nil {

		if err == gocql.ErrNotFound {
			return gocql.UUID{}, nil, errors.New("Sender doesn't exist.")
		} else {
			log.Println("Error occured while executing query to check if user is legit or not in CreateDirectChannel repository function:", err)
			return gocql.UUID{}, nil, errors.New("Server error.")
		}
	}

	// check whether reciever username is legit or not
	var reciever_id gocql.UUID
	var reciever_username string
	var reciever_email string

	if err := r.db.Query(`
           SELECT
           id, username, email
           FROM users
           where username=?
           ALLOW FILTERING
    `, reciever).WithContext(ctx).Scan(&reciever_id, &reciever_username, &reciever_email); err != nil {

		if err == gocql.ErrNotFound {
			return gocql.UUID{}, nil, errors.New("Receiver doesn't exist.")
		} else {
			log.Println("Error occured while executing query to check if reciever username is legit or not in CreateDirectChannel repository function:", err)
			return gocql.UUID{}, nil, errors.New("Server error.")
		}
	}

	new_channel.Members = append(new_channel.Members, reciever_id)

	batch := gocql.NewBatch(gocql.LoggedBatch)

	// create a channel in the channels table
	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `INSERT 
           INTO channels 
           (id, channel_name, room_id, is_direct_channel, members, created_at) 
           VALUES (?, ?, ?, ?, ?, ?)`,
		Args: []interface{}{new_channel.ID, new_channel.ChannelName, new_channel.Room, new_channel.IsDirectChannel,
			new_channel.Members, new_channel.CreatedAt},
		Idempotent: true,
	})

	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `UPDATE rooms SET channels = channels + {?}, 
           members=?
           where id=?
           AND admin=?
           AND created_at=?`,
		Args:       []interface{}{new_channel.ID, nil, room_id, room_id, created_at},
		Idempotent: true,
	})

	// add to first user
	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `UPDATE users SET direct_msg_channels = direct_msg_channels  + {?}, 
           where id=?
           AND username=?
           AND email=?`,
		Args:       []interface{}{new_channel.ID, sender_id, sender_username, sender_email},
		Idempotent: true,
	})

	// add to second user
	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `UPDATE users SET direct_msg_channels = direct_msg_channels  + {?}, 
           where id=?
           AND username=?
           AND email=?`,
		Args:       []interface{}{new_channel.ID, reciever_id, reciever_username, reciever_email},
		Idempotent: true,
	})

	err := r.db.ExecuteBatch(batch)
	if err != nil {
		log.Println("Error executing batch statement in create channel: ", err)
		return gocql.UUID{}, nil, errors.New("Server Error.")
	}

	return room_id, new_channel, nil
}

func (r *Repository) DeleteChannel(ctx context.Context, chn *Channel, admin gocql.UUID) (*Room, error) {

	// check if room exists and
	// admin given is actually the
	// admin of the room
	var room Room

	if err := r.db.Query(`
            SELECT 
            * 
            FROM rooms 
            WHERE id=?
            ALLOW FILTERING`,
		chn.Room).WithContext(ctx).Scan(&room.ID, &room.Admin, &room.CreatedAt, &room.Channels, &room.Members, &room.RoomName); err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.New("Room doesn't exist.")
		} else {
			log.Println("Error occured while executing query to get if user is admin or not in delete channel repository function:", err)
			return nil, errors.New("Server error.")
		}
	}

	if room.Admin != admin {
		return nil, errors.New("User is not the admin of this room.")
	}

	// check if channel exists or not
	var chn_created_at time.Time

	if err := r.db.Query(`
            SELECT 
            created_at 
            FROM channels 
            WHERE id=?
            ALLOW FILTERING`,
		chn.ID).WithContext(ctx).Scan(&chn_created_at); err != nil {
		if err == gocql.ErrNotFound {
			return nil, errors.New("Channel doesn't exist.")
		} else {
			log.Println("Error occured while executing query to get if channel exists or not in delete channel repository function:", err)
			return nil, errors.New("Server error.")
		}
	}

	if room.Admin != admin {
		return nil, errors.New("User is not the admin of this room.")
	}
	batch := gocql.NewBatch(gocql.LoggedBatch)

	// delete the channel in the channels table
	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `DELETE
           FROM channels
           WHERE id=?
           AND room_id=?
           AND created_at=?`,
		Args:       []interface{}{chn.ID, chn.Room, chn_created_at},
		Idempotent: true,
	})

	// remove the chn.ID from room.Members and
	// room.channels
	room.Channels = utils.RemoveElementFromArray(room.Channels, chn.ID)
	delete(room.ChannelMap, chn.ID)
	delete(room.Members, chn.ID)

	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `UPDATE rooms
           SET channels = channels - {?}, 
           members=?
           where id=?
           AND admin=?
           AND created_at=?`,
		Args:       []interface{}{chn.ID, room.Members, room.ID, room.Admin, room.CreatedAt},
		Idempotent: true,
	})

	err := r.db.ExecuteBatch(batch)
	if err != nil {
		log.Println("Error executing batch statement in delete channel: ", err)
		return nil, errors.New("Server Error.")
	}

	return &room, nil
}

func (r *Repository) WriteMessageToDB(ctx context.Context, msg *Message) error {

	log.Printf("message: %+v\n", msg)
	batch := r.db.NewBatch(gocql.LoggedBatch).WithContext(ctx)
	batch.Entries = append(batch.Entries, gocql.BatchEntry{
		Stmt: `INSERT 
           INTO messages 
           (channel_id, sender_id, room_id, content, timestamp, type) 
           VALUES (?, ?, ?, ?, ?, ?)`,
		Args:       []interface{}{msg.ChannelID, msg.SenderID, msg.RoomID, msg.Content, msg.Timestamp, msg.Type},
		Idempotent: true,
	})

	err := r.db.ExecuteBatch(batch)
	if err != nil {
		log.Println("Error executing batch statement in WriteMessageToDB: ", err)
		return errors.New("Server Error.")
	}

	return nil
}

func (r *Repository) CheckChannelMembership(ctx context.Context, roomID, channelID, userID gocql.UUID) (bool, error) {

	query := "SELECT members FROM channels WHERE room_id = ? AND id = ? ALLOW FILTERING"
	var members []gocql.UUID
	if err := r.db.Query(query, roomID, channelID).Scan(&members); err != nil {
		if err == gocql.ErrNotFound {
			return false, nil // Channel or room not found
		}
		return false, err
	}

	// Check if the user is a member of the channel
	for _, memberID := range members {
		if memberID == userID {
			return true, nil // User is a member of the channel
		}
	}

	return false, nil // User is not a member of the channel
}

func (r *Repository) CheckIfChannelExists(ctx context.Context, req *CreateDirectChannelReq) (*Channel, error) {
	query := "SELECT id from users where username=? ALLOW FILTERING"
	var uid gocql.UUID

	if err := r.db.Query(query, req.Reciever).WithContext(ctx).Scan(&uid); err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil //user not found
		}
		return nil, err
	}

	query = `
          SELECT 
          id, channel_name, room_id, members, 
          is_direct_channel, created_at 
          FROM channels 
          WHERE room_id = ? 
          AND members 
          CONTAINS ? AND members CONTAINS ?
          ALLOW FILTERING
          `
	channel := &Channel{}
	if err := r.db.Query(query, gocql.UUID{}, req.Sender, uid).Scan(
		&channel.ID, &channel.ChannelName, &channel.Room, &channel.Members,
		&channel.IsDirectChannel, &channel.CreatedAt,
	); err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil //user not found
		}
		return nil, err
	}

	return channel, nil
}

func (r *Repository) FetchAllMessages(ctx context.Context, chn_id gocql.UUID, user_id gocql.UUID, limit int, pg_state []byte) ([]Message, []byte, error) {
	query := "SELECT id from channels where id=? AND members CONTAINS ? ALLOW FILTERING"
	var uid gocql.UUID

	if err := r.db.Query(query, chn_id, user_id).WithContext(ctx).Scan(&uid); err != nil {
		if err == gocql.ErrNotFound {
			return nil, nil, errors.New("Channel not found") //user not found
		}
		return nil, nil, err
	}
	// TODO: Implement Paging
	// Fetch messages of the channel
	var messages []Message
	query = `SELECT 
           channel_id, sender_id, room_id, content, timestamp, type 
           FROM messages 
           WHERE channel_id = ? 
           ORDER BY timestamp DESC 
           `
	// var pageState []byte
	iter := r.db.Query(query, chn_id).PageSize(limit).PageState(pg_state).Iter()
	defer iter.Close()
	nextPageState := iter.PageState()

	var msg Message
	for iter.Scan(&msg.ChannelID, &msg.SenderID, &msg.RoomID, &msg.Content, &msg.Timestamp, &msg.Type) {
		messages = append(messages, msg)
	}
	if err := iter.Close(); err != nil {
		log.Println("Error at FetchAllMessages repo function: ", err)
		return nil, nil, err
	}

	return messages, nextPageState, nil
}
