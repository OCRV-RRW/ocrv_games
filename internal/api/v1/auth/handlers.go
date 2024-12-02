package auth

import (
	"Games/internal/DTO/userDTO"
	"Games/internal/config"
	"Games/internal/database"
	"Games/internal/models"
	"Games/internal/repository"
	"Games/internal/utils"
	"Games/internal/validation"
	"context"
	"github.com/gofiber/fiber/v2"
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
// @Param        SignUpInput		body		models.SignUpInput		true   "SignUpInput"
// @Success		 200
// @Failure      400
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
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"status": "fail", "message": err.Error()})
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
// @Param       SignInInput		body		models.SignInInput		true   "SignInInput"
// @Produce		json
// @Success		200
// @Failure     400
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
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "fail",
				"message": "the user belonging to this token no logger exists"})
		} else {
			return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
		}
	}

	accessTokenDetails, err := utils.CreateToken(user.ID.String(), config.AccessTokenExpiresIn, config.AccessTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	now := time.Now()

	errAccess := database.RedisClient.Set(
		ctx, accessTokenDetails.TokenUuid,
		user.ID.String(),
		time.Unix(*accessTokenDetails.ExpiresIn, 0).Sub(now)).Err()
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
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Invalid email"})
	}

	if !user.Verified {
		return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"status": "fail", "message": "Account not verified"})
	}

	config, _ := config.LoadConfig(".")
	//
	// Generate Verification Code
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

func ResetPassword(c *fiber.Ctx) error {
	message := "could not reset password."
	var payload *userDTO.ResetPasswordInput
	resetToken := c.Params("resetToken")

	repo := repository.NewUserRepository()
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	errors := validation.ValidateStruct(payload)
	if errors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "errors": errors})
	}

	if payload.Password != payload.ConfirmPassword {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "Passwords do not match"})
	}

	ctx := context.TODO()
	userid, err := database.RedisClient.Get(ctx, resetToken).Result()
	if err == redis.Nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": "The reset token is invalid or has expired"})
	}
	user, err := repo.GetUserById(userid)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"status": "fail", "message": message})
	}

	hashedPassword, _ := utils.HashPassword(payload.Password)

	user.Password = hashedPassword
	err = repo.Update(user)
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": message})
	}

	_, err = database.RedisClient.Del(ctx, resetToken).Result()
	if err != nil {
		return c.Status(fiber.StatusBadGateway).JSON(fiber.Map{"status": "fail", "message": err.Error()})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Password data updated successfully"})
}
