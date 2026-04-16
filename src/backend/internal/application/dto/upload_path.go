package dto

import "strings"

// NormalizeUploadPublicPath 将存储路径标准化为对外可访问的上传路径。
// 例如: "models/a.obj" -> "/uploads/models/a.obj"
func NormalizeUploadPublicPath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}

	normalized := strings.ReplaceAll(path, "\\", "/")

	if strings.HasPrefix(normalized, "http://") || strings.HasPrefix(normalized, "https://") {
		return normalized
	}
	if strings.HasPrefix(normalized, "/uploads/") {
		return normalized
	}
	if strings.HasPrefix(normalized, "uploads/") {
		return "/" + normalized
	}
	if strings.HasPrefix(normalized, "/") {
		return normalized
	}

	return "/uploads/" + strings.TrimLeft(normalized, "/")
}
