package config

import "os"

type Config struct {
	Port         string
	DatabasePath string
	JWTSecret    string
	FrontendURL  string

	// MinIO 配置（可选；未配置时附件上传接口将返回 STORAGE_DISABLED）
	MinioEndpoint  string
	MinioAccessKey string
	MinioSecretKey string
	MinioBucket    string
	MinioUseSSL    bool
	MinioPublicURL string
}

func Load() Config {
	return Config{
		Port:         getEnv("PORT", "8080"),
		DatabasePath: getEnv("DATABASE_PATH", "data/unistep.db"),
		JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
		FrontendURL:  getEnv("FRONTEND_URL", "http://localhost:5173"),

		MinioEndpoint:  getEnv("MINIO_ENDPOINT", ""),
		MinioAccessKey: getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:    getEnv("MINIO_BUCKET", "unistep"),
		MinioUseSSL:    getEnv("MINIO_USE_SSL", "false") == "true",
		MinioPublicURL: getEnv("MINIO_PUBLIC_URL", ""),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
