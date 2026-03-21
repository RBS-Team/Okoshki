package dto

//go:generate easyjson $GOFILE

//easyjson:json
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Role     string `json:"role"`
}

//easyjson:json
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

//easyjson:json
type LoginResponse struct {
	ID   string `json:"id"`
	Role string `json:"role"`
}

//easyjson:json
type RegisterResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
	Token string `json:"token,omitempty"`
}
