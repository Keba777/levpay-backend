package storage

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/Keba777/levpay-backend/internal/config"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client

// InitMinio initializes the MinIO client
func InitMinio() {
	cfg := config.CFG.Minio
	endpoint := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	accessKeyID := cfg.User
	secretAccessKey := cfg.Pass
	useSSL := cfg.SSL

	// Initialize minio client object.
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}

	MinioClient = minioClient

	// Create default bucket if not exists
	ctx := context.Background()
	bucketName := cfg.Bucket

	err = minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
	if err != nil {
		// Check to see if we already own this bucket (which happens if it exists)
		exists, errBucketExists := minioClient.BucketExists(ctx, bucketName)
		if errBucketExists == nil && exists {
			log.Printf("We already own %s\n", bucketName)
		} else {
			log.Fatalln(err)
		}
	} else {
		log.Printf("Successfully created %s\n", bucketName)
	}

	// Set bucket policy to public (read-only) for avatars?
	// Ideally we want avatars to be publicly readable.
	policy := fmt.Sprintf(`{
		"Version": "2012-10-17",
		"Statement": [
			{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}
		]
	}`, bucketName)

	if err := minioClient.SetBucketPolicy(ctx, bucketName, policy); err != nil {
		log.Printf("Failed to set bucket policy: %v", err)
	}
}

// UploadFile uploads a file to MinIO
func UploadFile(objectName string, reader io.Reader, size int64, contentType string) (string, error) {
	ctx := context.Background()
	bucketName := config.CFG.Minio.Bucket

	info, err := MinioClient.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}

	log.Printf("Successfully uploaded %s of size %d\n", objectName, info.Size)

	// Construct public URL
	scheme := "http"
	if config.CFG.Minio.SSL {
		scheme = "https"
	}

	url := fmt.Sprintf("%s://%s/%s/%s", scheme, config.CFG.Minio.Endpoint, bucketName, objectName)
	return url, nil
}
