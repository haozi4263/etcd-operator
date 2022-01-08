package file

import (
	"context"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type s3Uploader struct {
	Endpoint        string
	AccessKeyId     string
	SecretAccessKey string
}

func News3Uploader(Endpoint, AK, SK string) *s3Uploader {
	return &s3Uploader{
		Endpoint:        Endpoint,
		AccessKeyId:     AK,
		SecretAccessKey: SK,
	}
}

func (s *s3Uploader) InitClient() (*minio.Client, error) {
	return minio.New(s.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(s.AccessKeyId, s.SecretAccessKey, ""),
		Secure: true,
	})
}

func (s *s3Uploader) Uploader(ctx context.Context, bucketName, objectName, filePath string) (int64, error) {
	minioClient, err := s.InitClient()
	if err != nil {
		return 0, err
	}

	uploaderInfo, err := minioClient.FPutObject(ctx, bucketName, objectName, filePath, minio.PutObjectOptions{})
	if err != nil {
		return 0, err
	}

	return uploaderInfo.Size, nil
}
