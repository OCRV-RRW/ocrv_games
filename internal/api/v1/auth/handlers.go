package auth

import (
	"Games/internal/DTO/userDTO"
	"Games/internal/config"
	"Games/internal/database"
	"Games/internal/models"
	"Games/internal/utils"
	"Games/internal/validation"
	"Games/internal/validation/error_code"
	"context"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
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
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{Errors: errors})
	}

	if payload.Password != payload.PasswordConfirm {
		passwordError := []*validation.ErrorResponse{
			{
				Code:    error_code.PASSWORD_DO_NOT_MATCH,
				Message: "Passwords do not match",
			}}
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{Errors: passwordError})
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(payload.Password), bcrypt.DefaultCost)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{
			Errors: []*validation.ErrorResponse{
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
			Errors: []*validation.ErrorResponse{
				{
					Code:    error_code.ALREADY_EXIST,
					Message: "User with that email already exists",
				}}})
	} else if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(validation.ApiError{
			Errors: []*validation.ErrorResponse{
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
			Errors: []*validation.ErrorResponse{
				{
					Code:    error_code.PARSE_ERROR,
					Message: err.Error(),
				}}})
	}

	errors := validation.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{Errors: errors})
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
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{Errors: loginError})
	}

	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(payload.Password))
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(validation.ApiError{Errors: loginError})
	}

	config, _ := config.LoadConfig(".")
	accessTokenDetails, err := utils.CreateToken(user.ID.String(), config.AccessTokenExpiresIn, config.AccessTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	refreshTokenDetails, err := utils.CreateToken(user.ID.String(), config.RefreshTokenExpiresIn, config.RefreshTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	ctx := context.TODO()
	now := time.Now()

	errAccess := database.RedisClient.Set(ctx, accessTokenDetails.TokenUuid, user.ID.String(), time.Unix(*accessTokenDetails.ExpiresIn, 0).Sub(now)).Err()
	if errAccess != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": errAccess.Error()})
	}

	errRefresh := database.RedisClient.Set(ctx, refreshTokenDetails.TokenUuid, user.ID.String(), time.Unix(*refreshTokenDetails.ExpiresIn, 0).Sub(now)).Err()
	if errAccess != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": errRefresh.Error()})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    *accessTokenDetails.Token,
		Path:     "/",
		MaxAge:   config.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: true,
		Domain:   "localhost",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    *refreshTokenDetails.Token,
		Path:     "/",
		MaxAge:   config.RefreshTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: true,
		Domain:   "localhost",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "logged_in",
		Value:    "true",
		Path:     "/",
		MaxAge:   config.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: false,
		Domain:   "localhost",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "access_token": accessTokenDetails.Token})
}

// LogoutUser godoc
//
// @Description	logout
// @Accept		json
// @Produce		json
// @Success	    200
// @Router		/api/v1/auth/logout [get]
func LogoutUser(c *fiber.Ctx) error {
	message := "Token is invalid or session has expired"

	refresh_token := c.Cookies("refresh_token")

	if refresh_token == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": message})
	}

	config, _ := config.LoadConfig(".")
	ctx := context.TODO()

	tokenClaims, err := utils.ValidateToken(refresh_token, config.RefreshTokenPublicKey)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	access_token_uuid := c.Locals("access_token_uuid").(string)
	_, err = database.RedisClient.Del(ctx, tokenClaims.TokenUuid, access_token_uuid).Result()
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	expired := time.Now().Add(-time.Hour * 24)
	c.Cookie(&fiber.Cookie{
		Name:    "access_token",
		Value:   "",
		Expires: expired,
	})
	c.Cookie(&fiber.Cookie{
		Name:    "refresh_token",
		Value:   "",
		Expires: expired,
	})
	c.Cookie(&fiber.Cookie{
		Name:    "logged_in",
		Value:   "",
		Expires: expired,
	})
	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success"})
}

func RefreshAccessToken(c *fiber.Ctx) error {
	message := "could not refresh access token"

	refresh_token := c.Cookies("refresh_token")

	if refresh_token == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": message})
	}

	config, _ := config.LoadConfig(".")
	ctx := context.TODO()
	tokenClaims, err := utils.ValidateToken(refresh_token, config.RefreshTokenPublicKey)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	userid, err := database.RedisClient.Get(ctx, tokenClaims.TokenUuid).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": message})
	}

	var user models.User
	err = database.DB.First(&user, "id = ?", userid).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": "the user belonging to this token no logger exists"})
		} else {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
		}
	}

	accessTokenDetails, err := utils.CreateToken(user.ID.String(), config.AccessTokenExpiresIn, config.AccessTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	now := time.Now()

	errAccess := database.RedisClient.Set(ctx, accessTokenDetails.TokenUuid, user.ID.String(), time.Unix(*accessTokenDetails.ExpiresIn, 0).Sub(now)).Err()
	if errAccess != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": errAccess.Error()})
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    *accessTokenDetails.Token,
		Path:     "/",
		MaxAge:   config.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: true,
		Domain:   "localhost",
	})

	c.Cookie(&fiber.Cookie{
		Name:     "logged_in",
		Value:    "true",
		Path:     "/",
		MaxAge:   config.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: false,
		Domain:   "localhost",
	})

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "access_token": accessTokenDetails.Token})
}
