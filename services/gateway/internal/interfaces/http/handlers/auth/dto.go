package auth

// RegisterRequest — тело запроса POST /register.
type RegisterRequest struct {
	Email    string `json:"email"     validate:"required,email"`
	Password string `json:"password"  validate:"required,min=8"`
	FullName string `json:"full_name" validate:"required,min=2,max=50"`
}

// LoginRequest — тело запроса POST /login.
type LoginRequest struct {
	Email    string `json:"email"    validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// RefreshRequest — тело запроса POST /refresh.
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LogoutRequest — тело запроса POST /logout.
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}
