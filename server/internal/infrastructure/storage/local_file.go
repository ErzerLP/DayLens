// Package storage 本地文件存储适配器，实现 port.FileStorage。
package storage

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	_ "image/png" // PNG 解码支持
	"os"
	"path/filepath"
	"strings"
	"time"

	"golang.org/x/image/draw"

	"daylens-server/config"
	"daylens-server/internal/domain/activity"
)

// LocalFileStorage 本地磁盘截图存储
type LocalFileStorage struct {
	baseDir        string
	maxStorageMB   int
	retentionDays  int
	thumbnailWidth int
}

// NewLocalFileStorage 创建本地文件存储
func NewLocalFileStorage(cfg *config.StorageConfig) *LocalFileStorage {
	// 确保目录存在
	_ = os.MkdirAll(cfg.ScreenshotDir, 0o755)
	return &LocalFileStorage{
		baseDir:        cfg.ScreenshotDir,
		maxStorageMB:   cfg.MaxStorageMB,
		retentionDays:  cfg.RetentionDays,
		thumbnailWidth: cfg.ThumbnailWidth,
	}
}

// Save 保存文件到磁盘
func (s *LocalFileStorage) Save(_ context.Context, key string, data []byte) error {
	fullPath := filepath.Join(s.baseDir, filepath.FromSlash(key))
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("create dir: %w", err)
	}
	if err := os.WriteFile(fullPath, data, 0o644); err != nil {
		return fmt.Errorf("write file: %w", err)
	}
	return nil
}

// Get 读取文件原始内容
func (s *LocalFileStorage) Get(_ context.Context, key string) ([]byte, error) {
	fullPath := filepath.Join(s.baseDir, filepath.FromSlash(key))
	data, err := os.ReadFile(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", key)
		}
		return nil, fmt.Errorf("read file: %w", err)
	}
	return data, nil
}

// GetThumbnail 生成并返回缩略图
func (s *LocalFileStorage) GetThumbnail(_ context.Context, key string, width int) ([]byte, error) {
	if width <= 0 {
		width = s.thumbnailWidth
	}

	// 先检查缓存的缩略图
	thumbKey := thumbPath(key, width)
	thumbFull := filepath.Join(s.baseDir, filepath.FromSlash(thumbKey))
	if data, err := os.ReadFile(thumbFull); err == nil {
		return data, nil
	}

	// 读取原图
	origPath := filepath.Join(s.baseDir, filepath.FromSlash(key))
	origData, err := os.ReadFile(origPath)
	if err != nil {
		return nil, fmt.Errorf("read original: %w", err)
	}

	// 解码
	src, _, err := image.Decode(bytes.NewReader(origData))
	if err != nil {
		return nil, fmt.Errorf("decode image: %w", err)
	}

	// 计算缩略图尺寸（等比缩放）
	bounds := src.Bounds()
	origW := bounds.Dx()
	origH := bounds.Dy()
	if origW <= width {
		return origData, nil // 原图已够小
	}
	newH := origH * width / origW
	dst := image.NewRGBA(image.Rect(0, 0, width, newH))

	// Lanczos 高质量缩放
	draw.CatmullRom.Scale(dst, dst.Bounds(), src, src.Bounds(), draw.Over, nil)

	// 编码为 JPEG
	var buf bytes.Buffer
	if err := jpeg.Encode(&buf, dst, &jpeg.Options{Quality: 80}); err != nil {
		return nil, fmt.Errorf("encode thumbnail: %w", err)
	}

	// 缓存缩略图
	thumbDir := filepath.Dir(thumbFull)
	_ = os.MkdirAll(thumbDir, 0o755)
	_ = os.WriteFile(thumbFull, buf.Bytes(), 0o644)

	return buf.Bytes(), nil
}

// Delete 删除文件
func (s *LocalFileStorage) Delete(_ context.Context, key string) error {
	fullPath := filepath.Join(s.baseDir, filepath.FromSlash(key))
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("delete file: %w", err)
	}
	// 同时删除缩略图
	thumbFull := filepath.Join(s.baseDir, filepath.FromSlash(thumbPath(key, s.thumbnailWidth)))
	_ = os.Remove(thumbFull)
	return nil
}

// Stats 获取存储统计
func (s *LocalFileStorage) Stats(_ context.Context) (*activity.StorageStats, error) {
	var totalSize int64
	var fileCount int64
	oldestDate := time.Now().Format("2006-01-02")

	_ = filepath.Walk(s.baseDir, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		// 跳过缩略图
		if strings.Contains(path, "_thumb_") {
			return nil
		}
		totalSize += info.Size()
		fileCount++
		if info.ModTime().Format("2006-01-02") < oldestDate {
			oldestDate = info.ModTime().Format("2006-01-02")
		}
		return nil
	})

	return &activity.StorageStats{
		ScreenshotCount:    fileCount,
		DiskUsageMB:        totalSize / (1024 * 1024),
		MaxStorageMB:       int64(s.maxStorageMB),
		OldestActivityDate: oldestDate,
		RetentionDays:      s.retentionDays,
	}, nil
}

// thumbPath 生成缩略图存储路径
func thumbPath(key string, width int) string {
	ext := filepath.Ext(key)
	base := strings.TrimSuffix(key, ext)
	return fmt.Sprintf("%s_thumb_%d%s", base, width, ext)
}
