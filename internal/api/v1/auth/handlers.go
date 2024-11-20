package auth

import (
	"Games/internal/DTO/userDTO"
	"Games/internal/config"
	"Games/internal/database"
	"Games/internal/models"
	"Games/internal/validation"
	"Games/internal/validation/error_code"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"strings"
	"time"
)

// SignUpUser godoc
//
// @Description	 sign up user
// @Accept		 json
// @Produce		 json
// @Param        SignUpInput		body		models.SignUpInput		true   "SignUpInput"
// @Success		 200
// @Failure      400
// @Router		 /api/v1/auth/register [post]
func SignUpUser(c *fiber.Ctx) error {
	var payload *userDTO.SignUpInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	errors := validation.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{errors})
	}

	if payload.Password != payload.PasswordConfirm {
		passwordError := []*validation.ErrorResponse{
			{
				Code:    error_code.PASSWORD_DO_NOT_MATCH,
				Message: "Passwords do not match",
			}}
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{passwordError})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{
			[]*validation.ErrorResponse{
				{
					Code:    error_code.ERROR,
					Message: err.Error(),
				}}})
	}

	newUser := models.User{
		Name:     payload.Name,
		Email:    strings.ToLower(payload.Email),
		Password: hashedPassword,
	}
	result := database.DB.Create(&newUser)

	if result.Error != nil && strings.Contains(result.Error.Error(), "duplicate key value violates unique") {
		return c.Status(fiber.StatusConflict).JSON(validation.ApiError{
			[]*validation.ErrorResponse{
				{
					Code:    error_code.USER_ALREADY_EXIST,
					Message: "User with that email already exists",
				}}})
	} else if result.Error != nil {
		return c.Status(fiber.StatusBadGateway).JSON(validation.ApiError{
			[]*validation.ErrorResponse{
				{
					Code:    error_code.ERROR,
					Message: "Something bad happened",
				}}})
	}

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{"data": fiber.Map{"user": userDTO.FilterUserRecord(&newUser)}})
}

// SignInUser godoc
//
// @Description	sign in user
// @Accept      json
// @Param       SignInInput		body		models.SignInInput		true   "SignInInput"
// @Produce		json
// @Success		200
// @Failure     400
// @Router	    /api/v1/auth/login [post]
func SignInUser(c *fiber.Ctx) error {
	var payload *userDTO.SignInInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{
			[]*validation.ErrorResponse{
				{
					Code:    error_code.PARSE_ERROR,
					Message: err.Error(),
				}}})
	}

	errors := validation.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{errors})
	}

	var user models.User
	result := database.DB.First(&user, "email = ?", strings.ToLower(payload.Email))

	loginError := []*validation.ErrorResponse{
		{
			Code:    error_code.INVALID_LOGIN_DATA,
			Message: "Invalid email or Password",
		},
	}

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{loginError})
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{loginError})
	}

	config, _ := config.LoadConfig(".")
	tokenByte := jwt.New(jwt.SigningMethodHS256)

	now := time.Now().UTC()
	claims := tokenByte.Claims.(jwt.MapClaims)

	claims["sub"] = user.ID
	claims["exp"] = now.Add(config.JwtExpiresIn).Unix()
	claims["iat"] = now.Unix()
	claims["nbf"] = now.Unix()

	tokenString, err := tokenByte.SignedString([]byte(config.JwtSecret))

	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(validation.ApiError{
			[]*validation.ErrorResponse{{
				Code:    error_code.ERROR,
				Message: fmt.Sprintf("generating JWT Token failed: %v", err),
			}}})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "token",
		Value:    tokenString,
		Path:     "/",
		MaxAge:   config.JwtMaxAge * 60,
		Secure:   false,
		HTTPOnly: true,
		Domain:   "localhost",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"token": tokenString})
}

// LogoutUser godoc
//
// @Description	logout
// @Accept		json
// @Produce		json
// @Success	    200
// @Router		/api/v1/auth/logout [get]
func LogoutUser(c *fiber.Ctx) error {
	expired := time.Now().Add(-time.Hour * 24)
	c.Cookie(&fiber.Cookie{
		Name:    "token",
		Value:   "",
		Expires: expired,
	})
	c.Status(fiber.StatusOK)
	return nil
}
