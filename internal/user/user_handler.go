package user

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Handler struct {
	SERVICE
}

func NewHandler(s SERVICE) *Handler {
	return &Handler{
		SERVICE: s,
	}
}

func (h *Handler) CreateUser(w http.ResponseWriter, r *http.Request) {
	var u CreateUserReq

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &u)
	if err != nil {
		http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}

	// Check for null or empty fields
	if u.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}
	if u.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	if u.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if u.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	fmt.Println(u)

	res, err := h.SERVICE.CreateUser(r.Context(), &u)

	if err != nil {
		http.Error(w, fmt.Sprintln(err.Error()), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	var u LoginUserReq
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	err = json.Unmarshal(body, &u)
	if err != nil {
		http.Error(w, "Error unmarshalling request body", http.StatusBadRequest)
		return
	}

	// Check for null or empty fields
	if u.Email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}
	if u.Password == "" {
		http.Error(w, "Password is required", http.StatusBadRequest)
		return
	}

	response, err := h.SERVICE.Login(r.Context(), &u)

	if err != nil {
		http.Error(w, fmt.Sprintln("Invalid credentials."), http.StatusInternalServerError)
		return
	}

	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    response.accessToken,
		HttpOnly: true,
		Secure:   false,
		MaxAge:   3600,
	}

	res := &LoginUserRes{
		ID:       response.ID,
		Username: response.Username,
		Email:    response.Email,
		Name:     response.Name,
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *Handler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie := &http.Cookie{
		Name:     "jwt",
		Value:    "",
		HttpOnly: true,
		Secure:   false,
		MaxAge:   -1,
	}

	http.SetCookie(w, cookie)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Logout Successful",
	})
}
