package ws

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
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

	if create_room_req.Admin.String() == "" {
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

	if create_channel_req.Admin.String() == "" {
		http.Error(w, "Admin ID not provided", http.StatusBadRequest)
		return
	}

	if create_channel_req.ChannelName == "" {
		http.Error(w, "Channel name not provided", http.StatusBadRequest)
		return
	}

	res, err := h.SERVICE.CreateChannel(r.Context(), &create_channel_req)
	if err != nil {
		http.Error(w, fmt.Sprintln(err.Error()), http.StatusInternalServerError)
		return
	}

	h.Hub.Rooms[res.ID] = res

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)

}
