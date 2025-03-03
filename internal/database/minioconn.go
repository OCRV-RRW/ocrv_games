package database

import (
	"Games/internal/config"
	"Games/internal/models"
	"context"
	"errors"
	"fmt"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"log"
	"mime/multipart"
	"path/filepath"
	"strings"
)

var MinioClient *minio.Client
var minioCtx context.Context
var appBucket string

var (
	S3ErrorIncorrectFormat = errors.New("incorrect format")
)

func ConnectMinio(config *config.Config) {
	minioCtx = context.Background()
	endpoint := config.MinioHost
	accessKeyID := config.MinioAccessKey
	secretAccessKey := config.MinioSecretKey
	appBucket = config.AppBucket

	var err error
	// Initialize minio client object.
	MinioClient, err = minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: config.MinioSecure,
	})
	if err != nil {
		log.Fatalln(err)
	}

	log.Printf("%#v\n", MinioClient) // minioClient is now set up

	log.Println("✅ MinioClient client connected successfully...")
}

func PutObject(objectName string, fileReader multipart.File) (minio.UploadInfo, error) {
	info, err := MinioClient.PutObject(
		minioCtx, appBucket,
		objectName, fileReader,
		16, minio.PutObjectOptions{})
	return info, err
}

func PutGamePreviewer(game *models.Game, objectName string, fileReader multipart.File) (*minio.UploadInfo, error) {
	ext := filepath.Ext(objectName)
	ext = strings.ToLower(ext)
	if ext != ".png" && ext != ".jpg" {
		return nil, S3ErrorIncorrectFormat
	}

	info, err := MinioClient.PutObject(
		minioCtx, appBucket,
		fmt.Sprintf("/games/%s/preview%s", game.Name, ext), fileReader,
		-1, minio.PutObjectOptions{})

	return &info, err
}
