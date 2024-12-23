package DTO

import (
	"Games/internal/models"
	"github.com/google/uuid"
	"time"
)

type TokenResponse struct {
	AccessToken string `json:"access_token"`
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
	ID                 uuid.UUID `json:"id"`
	Name               string    `json:"name"`
	Email              string    `json:"email"`
	Age                int       `json:"age"`
	Grade              int       `json:"grade"`
	Gender             string    `json:"gender"`
	ContinuousProgress string    `json:"continuous_progress"`
	CreatedAt          time.Time `json:"created_at"`
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
	Age    int    `json:"age" validate:"number,lt=100,gt=17"`
	Gender string `json:"gender"`
	Grade  int    `json:"grade" validate:"number,lt=10,gt=0"`
}

func FilterUserRecord(user *models.User) UserResponse {
	return UserResponse{
		ID:        *user.ID,
		Name:      user.Name,
		Email:     user.Email,
		Age:       user.Age,
		Gender:    user.Gender,
		Grade:     user.Grade,
		CreatedAt: *user.CreatedAt,
	}
}

type AddScoreToSkillInput struct {
	SkillName string `json:"skill_name" validate:"required"`
	Score     int    `json:"score" validate:"number,gt=0"`
}

type UserSkillsResponse struct {
	Skills []UserSkill `json:"skills"`
}

type UserSkill struct {
	Name  string `json:"name" validate:"required"`
	Score int    `json:"score" validate:"number,gt=0"`
}

func FilterUserSkill(userSkill models.UserSkill) UserSkill {
	return UserSkill{
		Name:  userSkill.SkillName,
		Score: userSkill.Score,
	}
}
