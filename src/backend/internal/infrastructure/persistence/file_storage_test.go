// Package persistence_test 文件存储单元测试
package persistence_test

import (
	"mime/multipart"
	"testing"

	"github.com/xinyiis/BridgeDefectDetectionSys/src/backend/internal/infrastructure/persistence"
)

// TestFileService_ValidateFileFormat 测试文件格式验证
func TestFileService_ValidateFileFormat(t *testing.T) {
	fs := persistence.NewLocalFileStorage("./uploads")

	tests := []struct {
		name      string
		filename  string
		allowed   []string
		wantError bool
	}{
		{
			name:      "允许的格式 - .obj",
			filename:  "model.obj",
			allowed:   []string{".obj", ".fbx", ".gltf", ".glb"},
			wantError: false,
		},
		{
			name:      "允许的格式 - .fbx",
			filename:  "model.fbx",
			allowed:   []string{".obj", ".fbx", ".gltf", ".glb"},
			wantError: false,
		},
		{
			name:      "允许的格式 - .gltf",
			filename:  "model.gltf",
			allowed:   []string{".obj", ".fbx", ".gltf", ".glb"},
			wantError: false,
		},
		{
			name:      "允许的格式 - .glb",
			filename:  "model.glb",
			allowed:   []string{".obj", ".fbx", ".gltf", ".glb"},
			wantError: false,
		},
		{
			name:      "不允许的格式 - .txt",
			filename:  "model.txt",
			allowed:   []string{".obj", ".fbx", ".gltf", ".glb"},
			wantError: true,
		},
		{
			name:      "不允许的格式 - .pdf",
			filename:  "model.pdf",
			allowed:   []string{".obj", ".fbx", ".gltf", ".glb"},
			wantError: true,
		},
		{
			name:      "不允许的格式 - .exe",
			filename:  "model.exe",
			allowed:   []string{".obj", ".fbx", ".gltf", ".glb"},
			wantError: true,
		},
		{
			name:      "无扩展名",
			filename:  "model",
			allowed:   []string{".obj", ".fbx", ".gltf", ".glb"},
			wantError: true,
		},
		{
			name:      "大写扩展名（不匹配）",
			filename:  "model.OBJ",
			allowed:   []string{".obj", ".fbx"},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &multipart.FileHeader{Filename: tt.filename}
			err := fs.ValidateFileFormat(file, tt.allowed)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateFileFormat() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestFileService_ValidateFileSize 测试文件大小验证
func TestFileService_ValidateFileSize(t *testing.T) {
	fs := persistence.NewLocalFileStorage("./uploads")

	tests := []struct {
		name      string
		fileSize  int64
		maxSize   int64
		wantError bool
	}{
		{
			name:      "文件大小正常 - 1MB",
			fileSize:  1 * 1024 * 1024,
			maxSize:   50 * 1024 * 1024,
			wantError: false,
		},
		{
			name:      "文件大小正常 - 10MB",
			fileSize:  10 * 1024 * 1024,
			maxSize:   50 * 1024 * 1024,
			wantError: false,
		},
		{
			name:      "文件大小刚好等于上限",
			fileSize:  50 * 1024 * 1024,
			maxSize:   50 * 1024 * 1024,
			wantError: false,
		},
		{
			name:      "文件大小超过上限 - 51MB",
			fileSize:  51 * 1024 * 1024,
			maxSize:   50 * 1024 * 1024,
			wantError: true,
		},
		{
			name:      "文件大小超过上限 - 100MB",
			fileSize:  100 * 1024 * 1024,
			maxSize:   50 * 1024 * 1024,
			wantError: true,
		},
		{
			name:      "零字节文件",
			fileSize:  0,
			maxSize:   50 * 1024 * 1024,
			wantError: false,
		},
		{
			name:      "极小文件 - 1KB",
			fileSize:  1024,
			maxSize:   50 * 1024 * 1024,
			wantError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			file := &multipart.FileHeader{
				Filename: "test.obj",
				Size:     tt.fileSize,
			}
			err := fs.ValidateFileSize(file, tt.maxSize)
			if (err != nil) != tt.wantError {
				t.Errorf("ValidateFileSize() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

// TestFileService_ValidateFileFormat_EdgeCases 测试边界情况
func TestFileService_ValidateFileFormat_EdgeCases(t *testing.T) {
	fs := persistence.NewLocalFileStorage("./uploads")

	t.Run("空文件名", func(t *testing.T) {
		file := &multipart.FileHeader{Filename: ""}
		err := fs.ValidateFileFormat(file, []string{".obj"})
		if err == nil {
			t.Error("Empty filename should return error")
		}
	})

	t.Run("只有点的文件名", func(t *testing.T) {
		file := &multipart.FileHeader{Filename: "."}
		err := fs.ValidateFileFormat(file, []string{".obj"})
		if err == nil {
			t.Error("Filename with only dot should return error")
		}
	})

	t.Run("多个点的文件名", func(t *testing.T) {
		file := &multipart.FileHeader{Filename: "my.model.obj"}
		err := fs.ValidateFileFormat(file, []string{".obj"})
		if err != nil {
			t.Errorf("Filename with multiple dots should be valid, got error: %v", err)
		}
	})

	t.Run("空的允许列表", func(t *testing.T) {
		file := &multipart.FileHeader{Filename: "model.obj"}
		err := fs.ValidateFileFormat(file, []string{})
		if err == nil {
			t.Error("Empty allowed list should return error")
		}
	})
}

// TestFileService_ValidateFileSize_EdgeCases 测试大小验证边界情况
func TestFileService_ValidateFileSize_EdgeCases(t *testing.T) {
	fs := persistence.NewLocalFileStorage("./uploads")

	t.Run("负数文件大小", func(t *testing.T) {
		file := &multipart.FileHeader{
			Filename: "test.obj",
			Size:     -1,
		}
		// 注意：实际上文件大小不应该是负数，但我们测试一下
		err := fs.ValidateFileSize(file, 50*1024*1024)
		if err != nil {
			t.Errorf("Negative file size should pass (treated as 0), got error: %v", err)
		}
	})

	t.Run("零大小限制", func(t *testing.T) {
		file := &multipart.FileHeader{
			Filename: "test.obj",
			Size:     100,
		}
		err := fs.ValidateFileSize(file, 0)
		if err == nil {
			t.Error("File size exceeding zero limit should return error")
		}
	})
}

// TestFileService_CombinedValidation 测试组合验证
func TestFileService_CombinedValidation(t *testing.T) {
	fs := persistence.NewLocalFileStorage("./uploads")

	// 模拟真实场景：同时验证格式和大小
	t.Run("合法的3D模型文件", func(t *testing.T) {
		file := &multipart.FileHeader{
			Filename: "bridge_model.obj",
			Size:     30 * 1024 * 1024, // 30MB
		}

		// 验证格式
		if err := fs.ValidateFileFormat(file, []string{".obj", ".fbx", ".gltf", ".glb"}); err != nil {
			t.Errorf("Valid 3D model format should pass, got error: %v", err)
		}

		// 验证大小
		if err := fs.ValidateFileSize(file, 50*1024*1024); err != nil {
			t.Errorf("Valid file size should pass, got error: %v", err)
		}
	})

	t.Run("格式合法但大小超限", func(t *testing.T) {
		file := &multipart.FileHeader{
			Filename: "large_model.obj",
			Size:     60 * 1024 * 1024, // 60MB
		}

		// 验证格式 - 应该通过
		if err := fs.ValidateFileFormat(file, []string{".obj", ".fbx", ".gltf", ".glb"}); err != nil {
			t.Errorf("Valid format should pass, got error: %v", err)
		}

		// 验证大小 - 应该失败
		if err := fs.ValidateFileSize(file, 50*1024*1024); err == nil {
			t.Error("Oversized file should fail validation")
		}
	})

	t.Run("格式不合法但大小正常", func(t *testing.T) {
		file := &multipart.FileHeader{
			Filename: "model.txt",
			Size:     10 * 1024 * 1024, // 10MB
		}

		// 验证格式 - 应该失败
		if err := fs.ValidateFileFormat(file, []string{".obj", ".fbx", ".gltf", ".glb"}); err == nil {
			t.Error("Invalid format should fail validation")
		}

		// 验证大小 - 应该通过
		if err := fs.ValidateFileSize(file, 50*1024*1024); err != nil {
			t.Errorf("Valid size should pass, got error: %v", err)
		}
	})
}
