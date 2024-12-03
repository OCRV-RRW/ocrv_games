package auth

import (
	"Games/internal/DTO/userDTO"
	"Games/internal/config"
	"Games/internal/database"
	"Games/internal/models"
	"Games/internal/repository"
	"Games/internal/token"
	"Games/internal/utils"
	"Games/internal/validation"
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
	"strings"
	"time"
)

// SignUpUser godoc
//
// @Description	 sign up user
// @Accept		 json
// @Produce		 json
// @Param        SignUpInput		body		userDTO.SignUpInput		true   "SignUpInput"
// @Success		 201
// @Failure      400
// @Failure      409
// @Failure      502
// @Router		 /api/v1/auth/register [post]
func SignUpUser(c *fiber.Ctx) error {
	var payload *userDTO.SignUpInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	errors := validation.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "errors": errors})
	}

	if payload.Password != payload.PasswordConfirm {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "passwords do not match"})
	}

	hashedPassword, err := utils.HashPassword(payload.Password)

	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	verificationCode := utils.GenerateCode(20)

	newUser := models.User{
		Name:             payload.Name,
		Email:            strings.ToLower(payload.Email),
		Password:         hashedPassword,
		Verified:         false,
		VerificationCode: verificationCode,
	}

	r := repository.NewUserRepository()
	err = r.Create(&newUser)

	if err != nil && strings.Contains(err.Error(), "duplicate key value violates unique") {
		return c.Status(fiber.StatusConflict).JSON(
			fiber.Map{"status": "fail", "message": "User with that email already exists"})
	} else if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	//Send verification code.
	config, _ := config.LoadConfig(".")
	utils.SendEmail(&newUser, &utils.EmailData{
		URL:       config.ClientOrigin + "/api/v1/auth/verifyemail/" + verificationCode,
		FirstName: newUser.Name,
		Subject:   "Your account verification code",
	}, "verificationCode.html")

	return c.Status(fiber.StatusCreated).JSON(fiber.Map{
		"status":  "success",
		"message": "We sent an email with a verification code to " + newUser.Email,
		"data":    fiber.Map{"user": userDTO.FilterUserRecord(&newUser)}})
}

// SignInUser godoc
//
// @Description	sign in user
// @Accept      json
// @Param       SignInInput		body		userDTO.SignInInput		true   "SignInInput"
// @Produce		json
// @Success		200
// @Failure     400
// @Failure     422
// @Router	    /api/v1/auth/login [post]
func SignInUser(c *fiber.Ctx) error {
	var payload *userDTO.SignInInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	errors := validation.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "errors": errors})
	}

	var user models.User
	result := database.DB.First(&user, "email = ?", strings.ToLower(payload.Email))

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid email or password"})
	}

	err := utils.VerifyPassword(user.Password, payload.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid email or password"})
	}

	if !user.Verified {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Please verify your email"})
	}

	config, _ := config.LoadConfig(".")
	accessTokenDetails, err := token.CreateToken(user.ID.String(), config.AccessTokenExpiresIn, config.AccessTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	refreshTokenDetails, err := token.CreateToken(user.ID.String(), config.RefreshTokenExpiresIn, config.RefreshTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}
	tokenRepo := token.NewAuthTokenRepository(database.RedisClient)
	now := time.Now()
	errAccess := tokenRepo.SaveToken(
		user.ID.String(),
		accessTokenDetails,
		time.Unix(*accessTokenDetails.ExpiresIn, 0).Sub(now))

	if errAccess != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": errAccess.Error()})
	}

	errRefresh := tokenRepo.SaveToken(
		user.ID.String(),
		refreshTokenDetails,
		time.Unix(*refreshTokenDetails.ExpiresIn, 0).Sub(now))

	if errRefresh != nil {
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
// @Failure     403
// @Failure     502
// @Router		/api/v1/auth/logout [get]
func LogoutUser(c *fiber.Ctx) error {
	message := "Token is invalid or session has expired"

	refresh_token := c.Cookies("refresh_token")

	if refresh_token == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": message})
	}

	config, _ := config.LoadConfig(".")

	tokenClaims, err := token.ValidateToken(refresh_token, config.RefreshTokenPublicKey)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	accessTokenUuid := c.Locals("access_token_uuid").(string)
	tokenRepo := token.NewAuthTokenRepository(database.RedisClient)
	err = tokenRepo.RemoveTokenByTokenUuid(accessTokenUuid, tokenClaims.TokenUuid)
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

// RefreshAccessToken godoc
//
// @Description	refresh access token
// @Accept		json
// @Success	    200
// @Failure     403
// @Failure     502
// @Failure     422
// @Router		/api/v1/auth/refresh [post]
func RefreshAccessToken(c *fiber.Ctx) error {
	message := "could not refresh access token"

	refresh_token := c.Cookies("refresh_token")

	if refresh_token == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": message})
	}
	config, _ := config.LoadConfig(".")

	// Validate refresh token, and search token in redis
	tokenClaims, err := token.ValidateToken(refresh_token, config.RefreshTokenPublicKey)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}
	tokenRepo := token.NewAuthTokenRepository(database.RedisClient)
	userid, err := tokenRepo.GetUserIdByTokenUuid(tokenClaims.TokenUuid)
	if errors.Is(err, redis.Nil) || userid == "" {
		return c.Status(fiber.StatusForbidden).JSON(fiber.Map{"status": "fail", "message": message})
	}

	var user models.User
	err = database.DB.First(&user, "id = ?", userid).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "fail",
				"message": "the user belonging to this token no logger exists"})
		} else {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
		}
	}

	// Create and save new access token
	now := time.Now()
	accessTokenDetails, err := token.CreateToken(user.ID.String(), config.AccessTokenExpiresIn, config.AccessTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}
	errAccess := tokenRepo.SaveToken(
		user.ID.String(),
		accessTokenDetails,
		time.Unix(*accessTokenDetails.ExpiresIn, 0).Sub(now))
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

// VerifyEmail godoc
//
// @Description	 verify user email
// @Produce      json
// @Param        verify_code   path string true "Verification code"
// @Success		 200
// @Failure      400
// @Failure      409
// @Router		 /api/v1/auth/verify-email [post]
func VerifyEmail(c *fiber.Ctx) error {
	verificationCode := c.Params("verificationCode")

	var updatedUser models.User
	result := database.DB.First(&updatedUser, "verification_code = ?", verificationCode)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid verification code or user doesn't exists"})
	}

	if updatedUser.Verified {
		return c.Status(fiber.StatusConflict).JSON(fiber.Map{"status": "fail", "message": "User already verified"})
	}

	updatedUser.VerificationCode = ""
	updatedUser.Verified = true
	database.DB.Save(&updatedUser)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Email verified successfully"})
}

// ForgotPassword godoc
//
// @Description	 forgot password
// @Accept		 json
// @Produce      json
// @Param        ForgotPasswordInput		body		userDTO.ForgotPasswordInput		true   "ForgotPasswordInput"
// @Success		 200
// @Failure      400
// @Failure      401 nil nil  "User email is not verified"
// @Failure      404
// @Failure      502
// @Router		 /api/v1/auth/forgot-password [post]
func ForgotPassword(c *fiber.Ctx) error {
	repo := repository.NewUserRepository()
	var payload userDTO.ForgotPasswordInput

	repository.NewUserRepository()
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	errors := validation.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "errors": errors})
	}
	user, err := repo.GetByEmail(payload.Email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"status": "fail", "message": "Invalid email"})
	}

	if !user.Verified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Account not verified"})
	}

	config, _ := config.LoadConfig(".")

	// Generate and send to email verification Code
	resetToken := utils.GenerateCode(30)
	ctx := context.TODO()
	database.RedisClient.Set(ctx, resetToken, user.ID.String(), config.ResetPasswordTokenExpiredIn)
	utils.SendEmail(user, &utils.EmailData{
		URL:       config.ClientOrigin + "/api/v1/auth/resetpassword/" + resetToken,
		FirstName: user.Name,
		Subject:   "Your password reset token (valid for 10min)",
	}, "resetPassword.html")

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":  "success",
		"message": "We sent an email with a reset code to " + user.Email,
	})
}

// ResetPassword godoc
//
// @Description	 reset user password
// @Accept		 json
// @Produce		 json
// @Param        reset_code   path string true "reset code"
// @Param        ResetPasswordInput		body		userDTO.ResetPasswordInput		true   "ResetPasswordInput"
// @Success		 200
// @Failure      400
// @Failure      502
// @Router		 /api/v1/auth/reset-password [patch]
func ResetPassword(c *fiber.Ctx) error {
	message := "could not reset password."
	var payload *userDTO.ResetPasswordInput
	resetToken := c.Params("resetToken")

	userRepo := repository.NewUserRepository()
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	userErrors := validation.ValidateStruct(payload)
	if userErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "userErrors": userErrors})
	}

	if payload.Password != payload.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Passwords do not match"})
	}

	tokenRepo := token.NewAuthTokenRepository(database.RedisClient)

	ctx := context.TODO()
	userid, err := database.RedisClient.Get(ctx, resetToken).Result()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "The reset token is invalid or has expired"})
	}
	user, err := userRepo.GetUserById(userid)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": message})
	}

	hashedPassword, _ := utils.HashPassword(payload.Password)

	user.Password = hashedPassword
	err = userRepo.Update(user)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": message})
	}

	_, err = database.RedisClient.Del(ctx, resetToken).Result()
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	err = tokenRepo.RemoveAllUserToken(user.ID.String())
	if err != nil {
		log.Warnf("Couldn't reset user token error: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Password data updated successfully"})
}
