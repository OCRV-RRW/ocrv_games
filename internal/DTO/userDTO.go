package DTO

import (
	"Games/internal/models"
	"github.com/google/uuid"
	"time"
)

type TokenResponse struct {
	AccessToken string     `json:"access_token"`
	ExpiredIn   *time.Time `json:"expired_in"`
}

type UserResponseDTO struct {
	User UserResponse `json:"user"`
}

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
	ID                 uuid.UUID           `json:"id"`
	Name               string              `json:"name"`
	Email              string              `json:"email"`
	IsAdmin            bool                `json:"is_admin"`
	Birthdate          time.Time           `json:"birthdate"`
	Grade              int                 `json:"grade"`
	Gender             string              `json:"gender"`
	ContinuousProgress string              `json:"continuous_progress"`
	Skills             []UserSkillResponse `json:"skills"`
	CreatedAt          time.Time           `json:"created_at"`
}

type UsersResponse struct {
	Users []UserResponse `json:"users"`
}

type ForgotPasswordInput struct {
	Email string `json:"email" validate:"required,email"`
}

type ResetPasswordInput struct {
	Password        string `json:"password" validate:"required"`
	PasswordConfirm string `json:"password_confirm" validate:"required"`
}

type UpdateUserInput struct {
	Age     int    `json:"age" validate:"number,lt=100,gt=17"`
	Gender  string `json:"gender"`
	Grade   int    `json:"grade" validate:"number,lt=10,gt=-1"`
	IsAdmin bool   `json:"is_admin"`
}

func FilterUserRecord(user *models.User, userSkills []UserSkillResponse) UserResponse {
	return UserResponse{
		ID:        *user.ID,
		Name:      user.Name,
		Email:     user.Email,
		IsAdmin:   user.IsAdmin,
		Birthdate: user.Birthdate,
		Gender:    user.Gender,
		Grade:     user.Grade,
		Skills:    userSkills,
		CreatedAt: *user.CreatedAt,
	}
}

type AddScoreToSkillInput struct {
	SkillName string `json:"skill_name" validate:"required"`
	Score     int    `json:"score" validate:"number,gt=0"`
}

type UserSkillsResponse struct {
	Skills []UserSkillResponse `json:"skills"`
}

type UserSkillResponse struct {
	Name         string `json:"name" validate:"required"`
	FriendlyName string `json:"friendly_name" validate:"required"`
	Description  string `json:"description" validate:"required"`
	Score        int    `json:"score" validate:"number,gt=0"`
}
