package ws

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/gocql/gocql"
	"github.com/gorilla/websocket"
)

type Handler struct {
	SERVICE
	Hub *Hub
}

func NewHandler(hub *Hub, s SERVICE) *Handler {
	return &Handler{
		Hub:     hub,
		SERVICE: s,
	}
}

func (h *Handler) CreateRoom(w http.ResponseWriter, r *http.Request) {
	var create_room_req CreateRoomReq

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &create_room_req)
	if err != nil {
		http.Error(w, "Error unmarshalling the request body", http.StatusBadRequest)
		return
	}

	if create_room_req.Admin == "" {
		http.Error(w, "Admin not provided", http.StatusBadRequest)
		return
	}

	if create_room_req.RoomName == "" {
		http.Error(w, "RoomName not provided", http.StatusBadRequest)
		return
	}

	res, err := h.SERVICE.CreateRoom(r.Context(), &create_room_req)
	if err != nil {
		http.Error(w, fmt.Sprintln(err.Error()), http.StatusInternalServerError)
		return
	}

	h.Hub.Rooms[res.ID] = res

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)

}

func (h *Handler) JoinRoom(w http.ResponseWriter, r *http.Request) {
	var join_room_req JoinOrLeaveRoomReq

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &join_room_req)
	if err != nil {
		http.Error(w, "Error unmarshalling the request body", http.StatusBadRequest)
		return
	}

	if join_room_req.UserID.String() == "" {
		http.Error(w, "UserID not provided", http.StatusBadRequest)
		return
	}

	if join_room_req.ID.String() == "" {
		http.Error(w, "RoomID not provided", http.StatusBadRequest)
		return
	}

	if join_room_req.Username == "" {
		http.Error(w, "Username not provided", http.StatusBadRequest)
		return
	}

	if join_room_req.Email == "" {
		http.Error(w, "Email not provided", http.StatusBadRequest)
		return
	}

	res, err := h.SERVICE.JoinRoom(r.Context(), &join_room_req)
	if err != nil {
		http.Error(w, fmt.Sprintln(err.Error()), http.StatusInternalServerError)
		return
	}

	h.Hub.Rooms[res.ID] = res

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)

}

func (h *Handler) LeaveRoom(w http.ResponseWriter, r *http.Request) {

	var leave_room_req JoinOrLeaveRoomReq

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &leave_room_req)
	if err != nil {
		http.Error(w, "Error unmarshalling the request body", http.StatusBadRequest)
		return
	}

	if leave_room_req.UserID.String() == "" {
		http.Error(w, "UserID not provided", http.StatusBadRequest)
		return
	}

	if leave_room_req.ID.String() == "" {
		http.Error(w, "RoomID not provided", http.StatusBadRequest)
		return
	}

	if leave_room_req.Username == "" {
		http.Error(w, "Username not provided", http.StatusBadRequest)
		return
	}

	if leave_room_req.Email == "" {
		http.Error(w, "Email not provided", http.StatusBadRequest)
		return
	}

	res, err := h.SERVICE.LeaveRoom(r.Context(), &leave_room_req)
	if err != nil {
		http.Error(w, fmt.Sprintln(err.Error()), http.StatusInternalServerError)
		return
	}

	h.Hub.Rooms[res.ID] = res

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("You are no longer the member of the room")
}

func (h *Handler) DeleteRoom(w http.ResponseWriter, r *http.Request) {

	var delete_room_req DeleteRoomReq

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &delete_room_req)
	if err != nil {
		http.Error(w, "Error unmarshalling the request body", http.StatusBadRequest)
		return
	}

	if delete_room_req.UserID.String() == "" {
		http.Error(w, "UserID not provided", http.StatusBadRequest)
		return
	}

	if delete_room_req.ID.String() == "" {
		http.Error(w, "RoomID not provided", http.StatusBadRequest)
		return
	}

	res, err := h.SERVICE.DeleteRoom(r.Context(), &delete_room_req)
	if err != nil {
		http.Error(w, fmt.Sprintln(err.Error()), http.StatusInternalServerError)
		return
	}

	delete(h.Hub.Rooms, res.ID)

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode("The room is successfully deleted.")
}

func (h *Handler) CreateChannel(w http.ResponseWriter, r *http.Request) {
	var create_channel_req CreateChannelReq

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &create_channel_req)

	if err != nil {
		http.Error(w, "Error unmarshalling the request body", http.StatusBadRequest)
		return
	}

	if create_channel_req.RoomID.String() == "" {
		http.Error(w, "Room ID not provided", http.StatusBadRequest)
		return
	}

	if create_channel_req.Admin == "" {
		http.Error(w, "Admin not provided", http.StatusBadRequest)
		return
	}

	if create_channel_req.ChannelName == "" {
		http.Error(w, "Channel name not provided", http.StatusBadRequest)
		return
	}

	res, chn, err := h.SERVICE.CreateChannel(r.Context(), &create_channel_req)
	if err != nil {
		http.Error(w, fmt.Sprintln(err.Error()), http.StatusInternalServerError)
		return
	}

	prev_chan_map := h.Hub.Rooms[res.ID].ChannelMap
	prev_chan_map[chn.ID] = chn
	res.ChannelMap = prev_chan_map

	h.Hub.Rooms[res.ID] = res

	// TODO: Selectively return data fields instead
	// of returning the entire res object

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)

}

func (h *Handler) CreateDirectChannel(w http.ResponseWriter, r *http.Request) {
	var create_channel_req CreateDirectChannelReq

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &create_channel_req)

	if err != nil {
		http.Error(w, "Error unmarshalling the request body", http.StatusBadRequest)
		return
	}

	if create_channel_req.Sender == "" {
		http.Error(w, "User not provided", http.StatusBadRequest)
		return
	}

	if create_channel_req.Reciever == "" {
		http.Error(w, "Reciever username not provided", http.StatusBadRequest)
		return
	}

	chn_new, err := h.SERVICE.CheckIfChannelExists(r.Context(), &create_channel_req)
	if err != nil {
		log.Println("error encountered at CheckIfChannelExists. err: ", err)
		http.Error(w, "Server error", http.StatusInternalServerError)
		return
	}

	if chn_new == nil {
		room_id, chn, err := h.SERVICE.CreateDirectChannel(r.Context(), &create_channel_req)
		if err != nil {
			http.Error(w, fmt.Sprintln(err.Error()), http.StatusInternalServerError)
			return
		}

		h.Hub.Rooms[room_id].Channels = append(h.Hub.Rooms[room_id].Channels, chn.ID)
		h.Hub.Rooms[room_id].ChannelMap[chn.ID] = chn
		chn_new = chn
	}

	w.WriteHeader(http.StatusSeeOther)
	json.NewEncoder(w).Encode(chn_new)
}

func (h *Handler) DeleteChannel(w http.ResponseWriter, r *http.Request) {
	var delete_channel_req DeleteChannelReq

	body, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()

	if err != nil {
		http.Error(w, "Error reading request body", http.StatusBadRequest)
		return
	}

	err = json.Unmarshal(body, &delete_channel_req)

	if err != nil {
		http.Error(w, "Error unmarshalling the request body", http.StatusBadRequest)
		return
	}

	if delete_channel_req.RoomID.String() == "" {
		http.Error(w, "Room ID not provided", http.StatusBadRequest)
		return
	}

	if delete_channel_req.Admin == "" {
		http.Error(w, "Admin ID not provided", http.StatusBadRequest)
		return
	}

	if delete_channel_req.ChannelID.String() == "" {
		http.Error(w, "Channel name not provided", http.StatusBadRequest)
		return
	}

	res, chn_id, err := h.SERVICE.DeleteChannel(r.Context(), &delete_channel_req)
	if err != nil {
		http.Error(w, fmt.Sprintln(err.Error()), http.StatusInternalServerError)
		return
	}

	prev_chan_map := h.Hub.Rooms[res.ID].ChannelMap
	delete(prev_chan_map, chn_id)

	res.ChannelMap = prev_chan_map
	h.Hub.Rooms[res.ID] = res

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// TODO: Figure-out a better number for ReadBufferSize and WriteBufferSize
// TODO: Check origin
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		// origin := r.Header.Get("Origin")
		// return origin == "http://localhost:3000"
		return true
	},
}

func (h *Handler) JoinChannel(w http.ResponseWriter, r *http.Request) {

	ChannelID, err := gocql.ParseUUID(chi.URLParam(r, "channel_id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "Channel ID not provided or incorrect", http.StatusBadRequest)
		return
	}

	Username := r.URL.Query().Get("username")

	UserID, err := gocql.ParseUUID(r.URL.Query().Get("user_id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "User ID not provided or incorrect", http.StatusBadRequest)
		return
	}

	direct_channel_string := r.URL.Query().Get("is_direct_channel")
	IsDirectChannel := false
	var RoomID gocql.UUID

	if direct_channel_string == "true" {
		IsDirectChannel = true
		room_id := r.URL.Query().Get("room_id")
		if room_id != "" {
			http.Error(w, "Room ID not required", http.StatusBadRequest)
			return
		}
	} else {
		IsDirectChannel = false
		room_id, err := gocql.ParseUUID(r.URL.Query().Get("room_id"))
		if err != nil {
			log.Println(err)
			http.Error(w, "Room ID not provided or incorrect", http.StatusBadRequest)
			return
		}

		RoomID = room_id
	}

	if Username == "" {
		log.Println("username not provided")
		http.Error(w, "Username not provided", http.StatusBadRequest)
		return
	}

	// for direct message channels all channels will
	// come under the room 00000000-0000-0000-0000-000000000000
	if IsDirectChannel == true {
		RoomID = gocql.UUID{}
	}

	join_channel_req := &JoinChannelReq{
		ChannelID:       ChannelID,
		RoomID:          RoomID,
		UserID:          UserID,
		Username:        Username,
		IsDirectChannel: IsDirectChannel,
	}

	// check the DB to see if roomid, user and channel exists or not
	is_present, err := h.SERVICE.CheckChannelMembership(r.Context(), join_channel_req)

	if !is_present && err == nil {
		http.Error(w, "User or channel doesn't exist.", http.StatusBadRequest)
		return
	}

	if err != nil {
		log.Println(`error occured which checking if room, channel, user exists in
                 JoinChannel function. err: `, err)
		http.Error(w, "Error occured.", http.StatusInternalServerError)
		return
	}

	// upgrade connection to websocket
	conn, err := upgrader.Upgrade(w, r, nil)

	if err != nil {
		log.Println("Error upgrading connection to websockets. ", err)
		http.Error(w, fmt.Sprintln("Cannot establish a websocket connection."), http.StatusInternalServerError)
		return
	}

	cl := &Client{
		Conn:      conn,
		ID:        join_channel_req.UserID,
		RoomID:    join_channel_req.RoomID,
		ChannelID: join_channel_req.ChannelID,
		Username:  join_channel_req.Username,
		Message:   make(chan *Message, 10), // TODO: Size of buffer?
	}

	msg := &Message{
		RoomID:         join_channel_req.RoomID,
		ChannelID:      join_channel_req.ChannelID,
		SenderID:       join_channel_req.UserID,
		SenderUsername: join_channel_req.Username,
		Timestamp:      time.Now(),
		Type:           "notification",
		Content:        fmt.Sprintf("%v joined the room.", join_channel_req.Username),
	}

	log.Println("joined register channel")
	h.Hub.Register <- cl

	h.Hub.Broadcast <- msg

	go cl.writeMessage()
	cl.readMessage(r.Context(), h)
}

func (h *Handler) FetchAllMessages(w http.ResponseWriter, r *http.Request) {
	requesting_user_id := r.Context().Value("requesting_user_id").(string)
	user_id, err := gocql.ParseUUID(requesting_user_id)
	if err != nil {
		log.Println(err)
		http.Error(w, "Error", http.StatusBadRequest)
		return
	}

	limitStr := r.URL.Query().Get("limit")
	log.Println(limitStr)

	limit, err := strconv.Atoi(limitStr)
	if err != nil {
		http.Error(w, "Invalid limit", http.StatusBadRequest)
		return
	}

	pg_state_string := r.URL.Query().Get("page_state")
	log.Println("incoming string", pg_state_string)

	pg_state, err := base64.URLEncoding.DecodeString(pg_state_string)
	log.Printf("buffer: %+v", pg_state)
	if err != nil {
		log.Println("Error decoding Base64 string:", err)
		http.Error(w, "Error decoding Base64 string", http.StatusBadRequest)
		return
	}

	chn_id, err := gocql.ParseUUID(chi.URLParam(r, "channel_id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "Channel ID not provided or incorrect", http.StatusBadRequest)
		return
	}

	res, pageState, err := h.SERVICE.FetchAllMessages(r.Context(), chn_id, user_id, limit, pg_state)
	if err != nil {
		log.Println(err)
		http.Error(w, fmt.Sprintln(err.Error()), http.StatusInternalServerError)
		return
	}

	log.Println(base64.URLEncoding.EncodeToString(pageState))

	extended_res := struct {
		Messages  []Message `json:"messages"`
		PageState string    `json:"page_state"`
	}{
		Messages:  res,
		PageState: base64.URLEncoding.EncodeToString(pageState),
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(extended_res)
}

func (h *Handler) GetRoomDetails(w http.ResponseWriter, r *http.Request) {

	RoomID, err := gocql.ParseUUID(chi.URLParam(r, "room_id"))
	if err != nil {
		log.Println(err)
		http.Error(w, "Room ID not provided or incorrect", http.StatusBadRequest)
		return
	}

	res, err := h.SERVICE.GetAllRoomDetails(r.Context(), RoomID)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}
