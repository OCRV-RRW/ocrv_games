package userDTO

import (
	"Games/internal/models"
	"github.com/google/uuid"
	"time"
)

type SignUpInput struct {
	Name            string `json:"name" validate:"required"`
	Email           string `json:"email" validate:"required,email"`
	Password        string `json:"password" validate:"required"`
	PasswordConfirm string `json:"password_confirm" validate:"required"`
}

type SignInInput struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordInput struct {
	Password        string `json:"password" validate:"required"`
	ConfirmPassword string `json:"confirm_password" validate:"required"`
}

func FilterUserRecord(user *models.User) UserResponse {
	return UserResponse{
		ID:        *user.ID,
		Name:      user.Name,
		Email:     user.Email,
		CreatedAt: *user.CreatedAt,
	}
}
