package config

import (
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	AppPort int `mapstructure:"APP_PORT"`

	ClientOrigin string `mapstructure:"CLIENT_ORIGIN"`

	//Postgres
	DBHost         string `mapstructure:"POSTGRES_HOST"`
	DBUserName     string `mapstructure:"POSTGRES_USER"`
	DBUserPassword string `mapstructure:"POSTGRES_PASSWORD"`
	DBName         string `mapstructure:"POSTGRES_DB"`
	DBPort         string `mapstructure:"POSTGRES_PORT"`

	//MINIO
	MinioPublicHost string `mapstructure:"MINIO_PUBLIC_HOST"`
	MinioSecure     bool   `mapstructure:"MINIO_SECURE"`
	MinioHost       string `mapstructure:"MINIO_HOST"`
	MinioAccessKey  string `mapstructure:"MINIO_ACCESS_KEY"`
	MinioSecretKey  string `mapstructure:"MINIO_SECRET_KEY"`
	AppBucket       string `mapstructure:"MINIO_BUCKET"`

	//Redis
	RedisPort string `mapstructure:"REDIS_PORT"`
	RedisHost string `mapstructure:"REDIS_HOST"`

	ResetPasswordTokenExpiredIn time.Duration `mapstructure:"RESET_PASSWORD_TOKEN_EXPIRED_IN"`

	// Token
	AccessTokenPrivateKey  string        `mapstructure:"ACCESS_TOKEN_PRIVATE_KEY"`
	AccessTokenPublicKey   string        `mapstructure:"ACCESS_TOKEN_PUBLIC_KEY"`
	RefreshTokenPrivateKey string        `mapstructure:"REFRESH_TOKEN_PRIVATE_KEY"`
	RefreshTokenPublicKey  string        `mapstructure:"REFRESH_TOKEN_PUBLIC_KEY"`
	AccessTokenExpiresIn   time.Duration `mapstructure:"ACCESS_TOKEN_EXPIRED_IN"`
	RefreshTokenExpiresIn  time.Duration `mapstructure:"REFRESH_TOKEN_EXPIRED_IN"`
	AccessTokenMaxAge      int           `mapstructure:"ACCESS_TOKEN_MAXAGE"`
	RefreshTokenMaxAge     int           `mapstructure:"REFRESH_TOKEN_MAXAGE"`

	//SMTP
	EmailFrom string `mapstructure:"EMAIL_FROM"`
	SMTPHost  string `mapstructure:"SMTP_HOST"`
	SMTPPass  string `mapstructure:"SMTP_PASS"`
	SMTPPort  int    `mapstructure:"SMTP_PORT"`
	SMTPUser  string `mapstructure:"SMTP_USER"`
}

func LoadConfig(path string) (config Config, err error) {
	viper.AddConfigPath(path)
	viper.SetConfigType("env")
	viper.SetConfigName(".env")

	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}

	err = viper.Unmarshal(&config)
	return
}
