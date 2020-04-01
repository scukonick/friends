package dispatcher

// Request - request in our system
// {"user_id": 1, "friends": [2, 3, 4]}
type Request struct {
	UserID  int   `json:"user_id"`
	Friends []int `json:"friends"`
}

// Response - not actual response
// but a message which clients receive
type Response struct {
	Online bool `json:"online"`
}
