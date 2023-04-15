package types

type User struct {
	ID                 int      `json:"id"`
	Username           string   `json:"username"`
	Email              string   `json:"email"`
	Name               string   `json:"name"`
	Password           string   `json:"password"`
	Rooms              []string `json:"rooms"`
	DirectMsgChannelID []int    `json:"direct_msg_channel_id"`
}
