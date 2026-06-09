package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"path"
	"time"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Config MinIO 配置
type Config struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	Bucket          string
	UseSSL          bool
	PublicURLPrefix string // 用于拼接对外可访问的 URL（可选）
}

// Client 是 MinIO 客户端的薄封装，便于在 handler 与测试中替换
type Client struct {
	cfg    Config
	client *minio.Client
	// 当 disabled 为 true 时表示未配置 MinIO，所有上传将返回错误
	disabled bool
}

// NewClient 创建 MinIO 客户端；若 endpoint 为空，则返回 disabled 的客户端
func NewClient(cfg Config) (*Client, error) {
	if cfg.Endpoint == "" {
		log.Println("[storage] MinIO endpoint 未配置，文件上传功能将被禁用")
		return &Client{cfg: cfg, disabled: true}, nil
	}

	cli, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("init minio client: %w", err)
	}

	c := &Client{cfg: cfg, client: cli}
	if err := c.ensureBucket(); err != nil {
		// 不阻塞服务启动，但提示开发者
		log.Printf("[storage] ensure bucket failed: %v", err)
		c.disabled = true
	}
	return c, nil
}

func (c *Client) ensureBucket() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exists, err := c.client.BucketExists(ctx, c.cfg.Bucket)
	if err != nil {
		return err
	}
	if !exists {
		return c.client.MakeBucket(ctx, c.cfg.Bucket, minio.MakeBucketOptions{})
	}
	return nil
}

// Disabled 暴露当前客户端是否可用，便于 handler 提示
func (c *Client) Disabled() bool { return c.disabled }

// Upload 上传文件到 MinIO，自动生成对象键
func (c *Client) Upload(ctx context.Context, category, fileName string, size int64, reader io.Reader, contentType string) (objectKey, url string, err error) {
	if c.disabled {
		return "", "", fmt.Errorf("minio storage disabled")
	}

	ext := path.Ext(fileName)
	objectKey = fmt.Sprintf("members/%s/%s%s", category, uuid.NewString(), ext)

	_, err = c.client.PutObject(ctx, c.cfg.Bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", "", fmt.Errorf("put object: %w", err)
	}

	if c.cfg.PublicURLPrefix != "" {
		url = fmt.Sprintf("%s/%s/%s", c.cfg.PublicURLPrefix, c.cfg.Bucket, objectKey)
	} else {
		url = fmt.Sprintf("%s/%s/%s", c.cfg.Endpoint, c.cfg.Bucket, objectKey)
	}
	return objectKey, url, nil
}
