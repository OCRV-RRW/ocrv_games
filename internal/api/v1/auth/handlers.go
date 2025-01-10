package auth

import (
	"Games/internal/DTO"
	"Games/internal/api"
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
	"strings"
	"time"
)

// SignUpUser godoc
//
// @Description	 sign up user
// @Tags       Auth
// @Accept		 json
// @Produce		 json
// @Param        SignUpInput		body		DTO.SignUpInput		true   "SignUpInput"
// @Success		 201 {object} api.SuccessResponse[DTO.UserResponseDTO]
// @Failure      400 {object} api.ErrorResponse
// @Failure      409 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Failure      422 {object} api.ErrorResponse
// @Router		 /api/v1/auth/register [post]
func SignUpUser(c *fiber.Ctx) error {
	var payload *DTO.SignUpInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.UnprocessableEntity, Message: err.Error()},
		}))
	}

	userErrors := validation.ValidateStruct(payload)
	if userErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse(userErrors))
	}

	if payload.Password != payload.PasswordConfirm {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.IncorrectParameter, Parameter: "password", Message: "password and confirm password don't match"},
			{Code: api.IncorrectParameter, Parameter: "confirm_password", Message: "password and confirm password don't match"},
		}))
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

	sendEmailError := errors.New("Something went wrong on sending email")

	err = r.CreateUserWithTransaction(&newUser, func(user *models.User) error {
		//Send verification code.
		config, _ := config.LoadConfig(".")
		err = utils.SendEmail(&newUser, &utils.EmailData{
			URL:       config.ClientOrigin + "/register/verify/" + verificationCode,
			FirstName: newUser.Name,
			Subject:   "Your account verification code",
		}, "verificationCode.html")
		if err != nil {
			return sendEmailError
		}
		return nil
	})

	if errors.Is(err, sendEmailError) {
		param := validation.GetJSONTag(payload, "Email")
		userErrors = []*api.Error{validation.GetErrorResponse(param, "email")}
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse(userErrors))
	} else if errors.Is(err, repository.ErrDuplicatedKey) {
		return c.Status(fiber.StatusConflict).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.EmailAlreadyExists, Message: "email already exists"},
		}))
	} else if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: err.Error()},
		}))
	}

	return c.Status(fiber.StatusCreated).JSON(api.NewSuccessResponse(
		DTO.UserResponseDTO{DTO.FilterUserRecord(&newUser)},
		"We sent an email with a verification code to "+newUser.Email))
}

// SignInUser godoc
//
// @Description	sign in user
// @Tags        Auth
// @Accept      json
// @Param       SignInInput		body		DTO.SignInInput		true   "SignInInput"
// @Produce		json
// @Success		200 {object} api.SuccessResponse[DTO.TokenResponse]
// @Failure     400 {object} api.ErrorResponse
// @Failure     422 {object} api.ErrorResponse
// @Router	    /api/v1/auth/login [post]
func SignInUser(c *fiber.Ctx) error {
	var payload *DTO.SignInInput

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.UnprocessableEntity, Message: err.Error()},
		}))
	}

	userErrors := validation.ValidateStruct(payload)
	if userErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse(userErrors))
	}

	var user models.User
	result := database.DB.First(&user, "email = ?", strings.ToLower(payload.Email))

	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.InvalidEmailOrPassword, Message: "invalid email or password"},
		}))
	}

	err := utils.VerifyPassword(user.Password, payload.Password)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.InvalidEmailOrPassword, Message: "invalid email or password"},
		}))
	}

	if !user.Verified {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.InvalidEmailOrPassword, Message: "invalid email or password"},
		}))
	}

	config, _ := config.LoadConfig(".")
	accessTokenDetails, err := token.CreateToken(user.ID.String(), config.AccessTokenExpiresIn, config.AccessTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: ""},
		}))
	}

	refreshTokenDetails, err := token.CreateToken(user.ID.String(), config.RefreshTokenExpiresIn, config.RefreshTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: ""},
		}))
	}

	tokenRepo := token.NewAuthTokenRepository(database.RedisClient)
	now := time.Now()
	errAccess := tokenRepo.SaveToken(
		user.ID.String(),
		accessTokenDetails,
		time.Unix(*accessTokenDetails.ExpiresIn, 0).Sub(now))

	if errAccess != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: ""},
		}))
	}

	errRefresh := tokenRepo.SaveToken(
		user.ID.String(),
		refreshTokenDetails,
		time.Unix(*refreshTokenDetails.ExpiresIn, 0).Sub(now))

	if errRefresh != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: ""},
		}))
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    *accessTokenDetails.Token,
		Path:     "/",
		MaxAge:   config.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: true,
		//Domain:   config.Domen,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "refresh_token",
		Value:    *refreshTokenDetails.Token,
		Path:     "/",
		MaxAge:   config.RefreshTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: true,
		//Domain:   config.Domen,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "logged_in",
		Value:    "true",
		Path:     "/",
		MaxAge:   config.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: false,
		//Domain:   config.Domen,
	})

	return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(
		fiber.Map{"access_token": accessTokenDetails.Token},
		"success authorize"))
}

// LogoutUser godoc
//
// @Description	logout
// @Tags         Auth
// @Accept		json
// @Produce		json
// @Success	    200 {object} api.Response
// @Failure     403 {object} api.ErrorResponse
// @Failure     500 {object} api.ErrorResponse
// @Router		/api/v1/auth/logout [get]
func LogoutUser(c *fiber.Ctx) error {
	message := "Token is invalid or session has expired"

	refresh_token := c.Cookies("refresh_token")
	if refresh_token == "" {
		return c.Status(fiber.StatusForbidden).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.Forbidden, Message: message},
		}))
	}

	config, _ := config.LoadConfig(".")

	tokenClaims, err := token.ValidateToken(refresh_token, config.RefreshTokenPublicKey)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.Forbidden, Message: message},
		}))
	}

	accessTokenUuid := c.Locals("access_token_uuid").(string)
	tokenRepo := token.NewAuthTokenRepository(database.RedisClient)
	err = tokenRepo.RemoveTokenByTokenUuid(accessTokenUuid, tokenClaims.TokenUuid)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError},
		}))
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
// @Tags        Auth
// @Accept		json
// @Success	    200 {object} api.SuccessResponse[DTO.TokenResponse]
// @Failure     403 {object} api.ErrorResponse
// @Failure     500 {object} api.ErrorResponse
// @Failure     422 {object} api.ErrorResponse
// @Router		/api/v1/auth/refresh [post]
func RefreshAccessToken(c *fiber.Ctx) error {
	message := "could not refresh access token"

	refresh_token := c.Cookies("refresh_token")

	if refresh_token == "" {
		return c.Status(fiber.StatusForbidden).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.Forbidden, Message: message},
		}))
	}
	config, _ := config.LoadConfig(".")

	// Validate refresh token, and search token in redis
	tokenClaims, err := token.ValidateToken(refresh_token, config.RefreshTokenPublicKey)
	if err != nil {
		return c.Status(fiber.StatusForbidden).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.Forbidden, Message: message},
		}))
	}
	tokenRepo := token.NewAuthTokenRepository(database.RedisClient)
	userid, err := tokenRepo.GetUserIdByTokenUuid(tokenClaims.TokenUuid)
	if errors.Is(err, redis.Nil) || userid == "" {
		return c.Status(fiber.StatusForbidden).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.Forbidden, Message: message},
		}))
	}

	rep := repository.NewUserRepository()
	user, err := rep.GetUserById(userid)

	if err != nil {
		if errors.Is(err, repository.ErrRecordNotFound) {
			return c.Status(fiber.StatusForbidden).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.Forbidden, Message: "the user belonging to this token no logger exists"},
			}))
		} else {
			return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
				{Code: api.ServerError, Message: err.Error()},
			}))
		}
	}

	// Create and save new access token
	now := time.Now()
	accessTokenDetails, err := token.CreateToken(user.ID.String(), config.AccessTokenExpiresIn, config.AccessTokenPrivateKey)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: err.Error()},
		}))
	}
	errAccess := tokenRepo.SaveToken(
		user.ID.String(),
		accessTokenDetails,
		time.Unix(*accessTokenDetails.ExpiresIn, 0).Sub(now))
	if errAccess != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError, Message: err.Error()},
		}))
	}

	c.Cookie(&fiber.Cookie{
		Name:     "access_token",
		Value:    *accessTokenDetails.Token,
		Path:     "/",
		MaxAge:   config.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: true,
		//Domain:   config.Domen,
	})

	c.Cookie(&fiber.Cookie{
		Name:     "logged_in",
		Value:    "true",
		Path:     "/",
		MaxAge:   config.AccessTokenMaxAge * 60,
		Secure:   false,
		HTTPOnly: false,
		//Domain:   config.Domen,
	})
	return c.Status(fiber.StatusOK).JSON(api.NewSuccessResponse(
		fiber.Map{"access_token": accessTokenDetails.Token},
		"success refresh token"))
}

// VerifyEmail godoc
//
// @Description	 verify user email
// @Tags         Auth
// @Produce      json
// @Param        verify_code   path string true "Verification code"
// @Success		 200 {object} api.Response
// @Failure      400 {object} api.ErrorResponse
// @Failure      409 {object} api.ErrorResponse
// @Router		 /api/v1/auth/verify-email [post]
func VerifyEmail(c *fiber.Ctx) error {
	verificationCode := c.Params("verificationCode")

	var updatedUser models.User
	result := database.DB.First(&updatedUser, "verification_code = ?", verificationCode)
	if result.Error != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.InvalidVerificationCode, Message: "Invalid verification code or user doesn't exists"},
		}))
	}

	if updatedUser.Verified {
		return c.Status(fiber.StatusConflict).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.InvalidVerificationCode, Message: "email already verified"},
		}))
	}

	updatedUser.VerificationCode = ""
	updatedUser.Verified = true
	database.DB.Save(&updatedUser)

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Email verified successfully"})
}

// ForgotPassword godoc
//
// @Description	 forgot password
// @Tags         Auth
// @Accept		 json
// @Produce      json
// @Param        ForgotPasswordInput		body		DTO.ForgotPasswordInput		true   "ForgotPasswordInput"
// @Success		 200 {object} api.Response
// @Failure      400 {object} api.ErrorResponse
// @Failure      401 {object} api.ErrorResponse  "User email is not verified"
// @Failure      404 {object} api.ErrorResponse
// @Router		 /api/v1/auth/forgot-password [post]
func ForgotPassword(c *fiber.Ctx) error {
	repo := repository.NewUserRepository()
	var payload DTO.ForgotPasswordInput

	repository.NewUserRepository()

	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.UnprocessableEntity, Message: err.Error()},
		}))
	}

	userErrors := validation.ValidateStruct(payload)
	if userErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse(userErrors))
	}

	user, err := repo.GetByEmail(payload.Email)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.IncorrectParameter, Parameter: "email", Message: "invalid email"},
		}))
	}

	if !user.Verified {
		return c.Status(fiber.StatusUnauthorized).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.Unauthorized, Message: "unauthorized"},
		}))
	}

	config, _ := config.LoadConfig(".")

	// Generate and send to email verification Code
	resetToken := utils.GenerateCode(30)
	ctx := context.TODO()
	database.RedisClient.Set(ctx, resetToken, user.ID.String(), config.ResetPasswordTokenExpiredIn)
	utils.SendEmail(user, &utils.EmailData{
		URL:       config.ClientOrigin + "/forgot_password/reset/" + resetToken,
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
// @Tags         Auth
// @Accept		 json
// @Produce		 json
// @Param        reset_code   path string true "reset code"
// @Param        ResetPasswordInput		body		DTO.ResetPasswordInput		true   "ResetPasswordInput"
// @Success		 200 {object} api.Response
// @Failure      400 {object} api.ErrorResponse
// @Failure      500 {object} api.ErrorResponse
// @Router		 /api/v1/auth/reset-password [patch]
func ResetPassword(c *fiber.Ctx) error {
	var payload *DTO.ResetPasswordInput
	resetToken := c.Params("resetToken")

	userRepo := repository.NewUserRepository()
	if err := c.BodyParser(&payload); err != nil {
		return c.Status(fiber.StatusUnprocessableEntity).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.UnprocessableEntity, Message: err.Error()},
		}))
	}
	userErrors := validation.ValidateStruct(payload)
	if userErrors != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse(userErrors))
	}

	if payload.Password != payload.PasswordConfirm {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.IncorrectParameter, Parameter: "password", Message: "password and password_confirm do not match"},
			{Code: api.IncorrectParameter, Parameter: "password_confirm", Message: "password and password_confirm do not match"},
		}))
	}

	tokenRepo := token.NewAuthTokenRepository(database.RedisClient)

	ctx := context.TODO()
	userid, err := database.RedisClient.Get(ctx, resetToken).Result()
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.TokenInvalidOrExpired, Message: "The reset token is invalid or has expired"},
		}))
	}
	user, err := userRepo.GetUserById(userid)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.NotFound},
		}))
	}

	hashedPassword, _ := utils.HashPassword(payload.Password)

	user.Password = hashedPassword
	err = userRepo.Update(user)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError},
		}))
	}

	_, err = database.RedisClient.Del(ctx, resetToken).Result()
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(api.NewErrorResponse([]*api.Error{
			{Code: api.ServerError},
		}))
	}

	err = tokenRepo.RemoveAllUserToken(user.ID.String())
	if err != nil {
		log.Warnf("Couldn't reset user token error: %v", err)
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{"status": "success", "message": "Password data updated successfully"})
}
