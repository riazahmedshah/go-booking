package gcs

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"cloud.google.com/go/storage"
	"github.com/riazahmedshah/stayz/internal/config"
	"google.golang.org/api/option"
)

type GCSClient struct {
	cfg    *config.Config
	client *storage.Client
}

func NewGCSClient(cfg *config.Config) (*GCSClient, error) {
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithAuthCredentialsFile(option.ServiceAccount, cfg.GCS.GoogleCredentialPath))
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}
	return &GCSClient{
		cfg:    cfg,
		client: client,
	}, nil
}

func (g *GCSClient) UploadFile(ctx context.Context, objectName string, reader []byte) error {
	wc := g.client.Bucket(g.cfg.GCS.GCSBucketName).Object(objectName).NewWriter(ctx)

	if _, err := io.Copy(wc, bytes.NewReader(reader)); err != nil {
		return fmt.Errorf("failed to copy data: %w", err)
	}

	if err := wc.Close(); err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	return nil
}

func (g *GCSClient) DeleteFile(ctx context.Context, objectName string) error {
	if err := g.client.Bucket(g.cfg.GCS.GCSBucketName).Object(objectName).Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}
