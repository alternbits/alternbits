package utils

import (
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/dariubs/altern/app/config"
	"github.com/google/uuid"
)

type R2Service struct {
	client    *s3.Client
	bucket    string
	region    string
	accountID string
	publicURL string
}

func NewR2Service() (*R2Service, error) {
	if !config.C.R2Enabled() {
		return nil, fmt.Errorf("R2 configuration is incomplete")
	}

	r2 := config.C.CloudflareR2
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", r2.AccountID)

	cfg, err := awsconfig.LoadDefaultConfig(context.TODO(),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			r2.AccessKeyID, r2.SecretAccessKey, "",
		)),
		awsconfig.WithRegion(r2.Region),
	)
	if err != nil {
		return nil, fmt.Errorf("load r2 config: %w", err)
	}

	client := s3.NewFromConfig(cfg, func(o *s3.Options) {
		o.BaseEndpoint = aws.String(endpoint)
	})

	return &R2Service{
		client:    client,
		bucket:    r2.Bucket,
		region:    r2.Region,
		accountID: r2.AccountID,
		publicURL: r2.PublicURL,
	}, nil
}

// UploadFile uploads to R2 under folder/ and returns a browser-accessible URL.
func (r *R2Service) UploadFile(file *multipart.FileHeader, folder string) (string, error) {
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("open upload: %w", err)
	}
	defer src.Close()

	ext := strings.ToLower(filepath.Ext(file.Filename))
	key := fmt.Sprintf("%s/%s%s", folder, uuid.New().String(), ext)

	_, err = r.client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:      aws.String(r.bucket),
		Key:         aws.String(key),
		Body:        src,
		ContentType: aws.String(file.Header.Get("Content-Type")),
	})
	if err != nil {
		return "", fmt.Errorf("put object: %w", err)
	}

	return r.urlFor(key), nil
}

// DeleteByURL removes an object previously returned by UploadFile.
func (r *R2Service) DeleteByURL(fileURL string) error {
	key := r.keyFromURL(fileURL)
	if key == "" {
		return fmt.Errorf("cannot derive object key from URL")
	}
	_, err := r.client.DeleteObject(context.TODO(), &s3.DeleteObjectInput{
		Bucket: aws.String(r.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object: %w", err)
	}
	return nil
}

func (r *R2Service) urlFor(key string) string {
	if r.publicURL != "" {
		return r.publicURL + "/" + key
	}
	return fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/%s", r.accountID, r.bucket, key)
}

func (r *R2Service) keyFromURL(fileURL string) string {
	if r.publicURL != "" {
		if key, ok := strings.CutPrefix(fileURL, r.publicURL+"/"); ok {
			return key
		}
	}
	prefix := fmt.Sprintf("https://%s.r2.cloudflarestorage.com/%s/", r.accountID, r.bucket)
	if key, ok := strings.CutPrefix(fileURL, prefix); ok {
		return key
	}
	return ""
}
