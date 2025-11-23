package user

// RegisterRequest defines the shape of incoming registration JSON
type RegisterRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// AuthResponse defines what we send back on success
type AuthResponse struct {
	Token    string `json:"token"`
	UserID   string `json:"user_id"`
	UserName string `json:"user_name"`
}
