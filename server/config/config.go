// Package config 加载和管理应用配置。
package config

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/viper"
)

// Config 应用全局配置
type Config struct {
	Server   ServerConfig   `mapstructure:"server"`
	Auth     AuthConfig     `mapstructure:"auth"`
	Database DatabaseConfig `mapstructure:"database"`
	AI       AIConfig       `mapstructure:"ai"`
	Storage  StorageConfig  `mapstructure:"storage"`
	Schedule ScheduleConfig `mapstructure:"work_schedule"`
	TLS      TLSConfig      `mapstructure:"tls"`
	Security SecurityConfig `mapstructure:"security"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	EncryptionKey string `mapstructure:"encryption_key"` // 32 字节 hex，为空则不加密
}

// ServerConfig HTTP 服务配置
type ServerConfig struct {
	Addr         string `mapstructure:"addr"`
	InternalAddr string `mapstructure:"internal_addr"`
	Timezone     string `mapstructure:"timezone"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	Token string `mapstructure:"token"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	URL      string `mapstructure:"url"`
	MaxConns int32  `mapstructure:"max_conns"`
	MinConns int32  `mapstructure:"min_conns"`
}

// AIConfig AI 服务配置
type AIConfig struct {
	Provider     string `mapstructure:"provider"`
	Endpoint     string `mapstructure:"endpoint"`
	Model        string `mapstructure:"model"`
	APIKey       string `mapstructure:"api_key"`
	CustomPrompt string `mapstructure:"custom_prompt"`
}

// StorageConfig 截图存储配置
type StorageConfig struct {
	ScreenshotDir  string `mapstructure:"screenshot_dir"`
	RetentionDays  int    `mapstructure:"retention_days"`
	MaxStorageMB   int    `mapstructure:"max_storage_mb"`
	ThumbnailWidth int    `mapstructure:"thumbnail_width"`
}

// ScheduleConfig 工作时间配置
type ScheduleConfig struct {
	StartHour   int `mapstructure:"start_hour"`
	StartMinute int `mapstructure:"start_minute"`
	EndHour     int `mapstructure:"end_hour"`
	EndMinute   int `mapstructure:"end_minute"`
}

// TLSConfig TLS 配置
type TLSConfig struct {
	Enabled  bool   `mapstructure:"enabled"`
	CertFile string `mapstructure:"cert_file"`
	KeyFile  string `mapstructure:"key_file"`
}

// Load 加载配置（config.yaml + 环境变量）
func Load() *Config {
	v := viper.New()

	// 默认值
	v.SetDefault("server.addr", "0.0.0.0:8080")
	v.SetDefault("server.internal_addr", "127.0.0.1:8081")
	v.SetDefault("database.url", "postgres://daylens:daylens_secret@localhost:5432/daylens?sslmode=disable")
	v.SetDefault("database.max_conns", 20)
	v.SetDefault("database.min_conns", 5)
	v.SetDefault("ai.provider", "ollama")
	v.SetDefault("ai.endpoint", "http://localhost:11434")
	v.SetDefault("ai.model", "qwen2.5")
	v.SetDefault("server.timezone", "Asia/Shanghai")
	v.SetDefault("storage.screenshot_dir", "./data/screenshots")
	v.SetDefault("storage.retention_days", 30)
	v.SetDefault("storage.max_storage_mb", 2048)
	v.SetDefault("storage.thumbnail_width", 360)
	v.SetDefault("work_schedule.start_hour", 9)
	v.SetDefault("work_schedule.end_hour", 18)

	// 配置文件
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Printf("读取配置文件失败: %v，使用默认配置", err)
		}
	}

	// 环境变量覆盖（DAYLENS_DATABASE_URL 等）
	v.SetEnvPrefix("DAYLENS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("解析配置失败: %v", err)
	}

	// Token 为空时自动生成
	if cfg.Auth.Token == "" {
		cfg.Auth.Token = generateToken()
		fmt.Printf("⚠️  未配置 Token，已自动生成: %s\n", cfg.Auth.Token) // DayLens
	}

	return &cfg
}

// generateToken 生成 256-bit 随机 Token
func generateToken() string {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		log.Fatalf("生成 Token 失败: %v", err)
	}
	return hex.EncodeToString(b)
}
