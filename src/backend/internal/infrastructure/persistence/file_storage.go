// Package persistence 实现数据访问层
// 包含所有Repository接口的具体实现
package persistence

import (
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/domain/service"
)

// LocalFileStorage 本地文件存储实现
// 实现 service.FileService 接口
type LocalFileStorage struct {
	baseDir string // 基础目录，如: "./uploads"
}

// NewLocalFileStorage 创建本地文件存储实例
// 参数：
//   - baseDir: 基础目录路径
// 返回：
//   - service.FileService: 文件服务接口
func NewLocalFileStorage(baseDir string) service.FileService {
	return &LocalFileStorage{baseDir: baseDir}
}

// SaveUploadedFile 保存上传文件
// 参数：
//   - file: 上传的文件
//   - dir: 保存目录（相对路径）
// 返回：
//   - string: 文件相对路径
//   - error: 错误信息
func (s *LocalFileStorage) SaveUploadedFile(file *multipart.FileHeader, dir string) (string, error) {
	// 1. 生成唯一文件名（UUID + 原扩展名）
	ext := filepath.Ext(file.Filename)
	filename := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// 2. 构建完整路径: uploads/models/uuid.obj
	fullPath := filepath.Join(s.baseDir, dir, filename)

	// 3. 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 4. 打开上传文件
	src, err := file.Open()
	if err != nil {
		return "", fmt.Errorf("打开上传文件失败: %w", err)
	}
	defer src.Close()

	// 5. 创建目标文件
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("创建文件失败: %w", err)
	}
	defer dst.Close()

	// 6. 复制文件内容
	if _, err = io.Copy(dst, src); err != nil {
		return "", fmt.Errorf("复制文件内容失败: %w", err)
	}

	// 7. 返回相对路径（存入数据库）
	return filepath.Join(dir, filename), nil
}

// SaveImage 保存图片文件（快捷方法）
// 参数：
//   - file: 上传的图片文件
//   - dir: 保存目录（相对路径）
// 返回：
//   - string: 文件相对路径
//   - error: 错误信息
func (s *LocalFileStorage) SaveImage(file *multipart.FileHeader, dir string) (string, error) {
	return s.SaveUploadedFile(file, dir)
}

// SaveResultImage 保存结果图（从Base64）
// 参数：
//   - base64Data: Base64编码的图片数据
//   - dir: 保存目录（相对路径）
// 返回：
//   - string: 文件相对路径
//   - error: 错误信息
func (s *LocalFileStorage) SaveResultImage(base64Data string, dir string) (string, error) {
	// 1. 解码Base64数据
	imageData, err := base64.StdEncoding.DecodeString(base64Data)
	if err != nil {
		return "", fmt.Errorf("Base64解码失败: %w", err)
	}

	// 2. 生成唯一文件名（默认.jpg扩展名）
	filename := fmt.Sprintf("%s.jpg", uuid.New().String())

	// 3. 构建完整路径
	fullPath := filepath.Join(s.baseDir, dir, filename)

	// 4. 确保目录存在
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("创建目录失败: %w", err)
	}

	// 5. 写入文件
	if err := os.WriteFile(fullPath, imageData, 0644); err != nil {
		return "", fmt.Errorf("写入文件失败: %w", err)
	}

	// 6. 返回相对路径
	return filepath.Join(dir, filename), nil
}

// DeleteFile 删除文件
// 参数：
//   - path: 文件相对路径
// 返回：
//   - error: 错误信息
func (s *LocalFileStorage) DeleteFile(path string) error {
	fullPath := filepath.Join(s.baseDir, path)
	if err := os.Remove(fullPath); err != nil {
		// 文件不存在不算错误
		if os.IsNotExist(err) {
			return nil
		}
		return fmt.Errorf("删除文件失败: %w", err)
	}
	return nil
}

// ValidateFileFormat 验证文件格式
// 参数：
//   - file: 上传的文件
//   - allowedExts: 允许的扩展名列表
// 返回：
//   - error: 格式不支持时返回错误
func (s *LocalFileStorage) ValidateFileFormat(file *multipart.FileHeader, allowedExts []string) error {
	ext := filepath.Ext(file.Filename)
	for _, allowed := range allowedExts {
		if ext == allowed {
			return nil
		}
	}
	return fmt.Errorf("不支持的文件格式: %s，允许的格式: %v", ext, allowedExts)
}

// ValidateFileSize 验证文件大小
// 参数：
//   - file: 上传的文件
//   - maxSize: 最大文件大小（字节）
// 返回：
//   - error: 超过大小限制时返回错误
func (s *LocalFileStorage) ValidateFileSize(file *multipart.FileHeader, maxSize int64) error {
	if file.Size > maxSize {
		return fmt.Errorf("文件大小超过限制: %.2f MB (最大 %.2f MB)",
			float64(file.Size)/1024/1024,
			float64(maxSize)/1024/1024)
	}
	return nil
}
