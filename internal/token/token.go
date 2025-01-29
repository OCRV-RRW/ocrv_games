package token

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"strings"
	"time"
)

type TokenDetails struct {
	Token     *string
	TokenUuid string
	UserID    string
	ExpiresIn *int64
}

type AuthTokenRepository struct {
	rdb *redis.Client
}

func NewAuthTokenRepository(rdb *redis.Client) *AuthTokenRepository {
	return &AuthTokenRepository{rdb}
}

func (c *AuthTokenRepository) removeTokensByPattern(pattens ...string) error {
	ctx := context.TODO()
	for _, pattern := range pattens {
		iter := c.rdb.Scan(ctx, 0, pattern, 0).Iterator()
		for iter.Next(ctx) {
			c.rdb.Del(ctx, iter.Val())
		}
		if err := iter.Err(); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func (c *AuthTokenRepository) RemoveAllUserToken(userId string) error {
	pattern := fmt.Sprintf("%s:*", userId)
	return c.removeTokensByPattern(pattern)
}

func (c *AuthTokenRepository) RemoveTokenByTokenUuid(tokenUuids ...string) error {
	for _, tokenUuid := range tokenUuids {
		pattern := fmt.Sprintf("*:%s", tokenUuid)
		if err := c.removeTokensByPattern(pattern); err != nil {
			return err
		}
	}
	return nil
}

func (c *AuthTokenRepository) GetUserIdByTokenUuid(tokenUuid string) (string, error) {
	ctx := context.TODO()
	pattern := fmt.Sprintf("*:%s", tokenUuid)
	var cursor uint64
	for {
		var keys []string
		var err error
		keys, cursor, err = c.rdb.Scan(ctx, cursor, pattern, 1).Result()
		if err != nil {
			return "", err
		}

		if len(keys) >= 1 {
			userId := strings.Split(keys[0], ":")[0]
			if userId == "" {
				return "", errors.New("not found")
			}
			return userId, nil
		}

		if cursor == 0 { // no more keys
			break
		}
	}
	return "", redis.Nil
}

func (c *AuthTokenRepository) SaveToken(userId string, token *TokenDetails, expiration time.Duration) error {
	ctx := context.TODO()
	key := fmt.Sprintf("%s:%s", userId, token.TokenUuid)
	return c.rdb.Set(ctx, key, userId, expiration).Err()
}

func CreateToken(userid string, ttl time.Duration, privateKey string) (*TokenDetails, error) {
	now := time.Now().UTC()
	td := &TokenDetails{
		ExpiresIn: new(int64),
		Token:     new(string),
	}
	*td.ExpiresIn = now.Add(ttl).Unix()
	token, err := uuid.NewV6()

	if err != nil {
		return nil, fmt.Errorf("Error generating token: %v", err)
	}
	td.TokenUuid = token.String()
	td.UserID = userid

	decodedPrivateKey, err := base64.StdEncoding.DecodeString(privateKey)
	if err != nil {
		return nil, fmt.Errorf("could not decode token private key: %w", err)
	}
	key, err := jwt.ParseRSAPrivateKeyFromPEM(decodedPrivateKey)

	if err != nil {
		return nil, fmt.Errorf("create: parse token private key: %w", err)
	}
	atClaims := make(jwt.MapClaims)
	atClaims["sub"] = td.UserID
	atClaims["token_uuid"] = td.TokenUuid
	atClaims["iat"] = now.Unix()
	atClaims["nbf"] = now.Unix()

	*td.Token, err = jwt.NewWithClaims(jwt.SigningMethodRS256, atClaims).SignedString(key)

	if err != nil {
		return nil, fmt.Errorf("create: sign token: %w", err)
	}

	return td, nil
}

func ValidateToken(token string, publicKey string) (*TokenDetails, error) {
	decodedPublicKey, err := base64.StdEncoding.DecodeString(publicKey)
	if err != nil {
		return nil, fmt.Errorf("could not decode: %w", err)
	}

	key, err := jwt.ParseRSAPublicKeyFromPEM(decodedPublicKey)

	if err != nil {
		return nil, fmt.Errorf("validate: parse key: %w", err)
	}

	parsedToken, err := jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected method: %s", t.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		return nil, fmt.Errorf("validate: %w", err)
	}

	claims, ok := parsedToken.Claims.(jwt.MapClaims)
	if !ok || !parsedToken.Valid {
		return nil, fmt.Errorf("validate: invalid token")
	}

	return &TokenDetails{
		TokenUuid: fmt.Sprint(claims["token_uuid"]),
		UserID:    fmt.Sprint(claims["sub"]),
	}, nil
}
