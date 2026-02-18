package domain

import "time"

type User struct {
	ID        int64     `json:"id"`
	Email     string    `json:"email"`
	Password  string    `json:"password,omitempty"` // omitempty so we don't return it in JSON
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
}

type AuthResponse struct {
	User  User   `json:"user"`
	Token string `json:"token"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

type RegisterRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
	Name     string `json:"name" binding:"required"`
}
