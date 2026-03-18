// Package service 定义领域服务层
// 包含核心业务逻辑和领域规则
package service

import "mime/multipart"

// FileService 文件服务接口
// 定义文件上传、删除、验证等操作
type FileService interface {
	// SaveUploadedFile 保存上传的文件
	// 参数：
	//   - file: 上传的文件
	//   - dir: 保存目录（相对路径，如 "models"）
	// 返回：
	//   - string: 文件相对路径（如 "models/uuid.obj"）
	//   - error: 错误信息
	SaveUploadedFile(file *multipart.FileHeader, dir string) (string, error)

	// SaveImage 保存图片文件（快捷方法）
	// 参数：
	//   - file: 上传的图片文件
	//   - dir: 保存目录（相对路径，如 "images"）
	// 返回：
	//   - string: 文件相对路径
	//   - error: 错误信息
	SaveImage(file *multipart.FileHeader, dir string) (string, error)

	// SaveResultImage 保存结果图（从Base64）
	// 参数：
	//   - base64Data: Base64编码的图片数据
	//   - dir: 保存目录（相对路径，如 "results"）
	// 返回：
	//   - string: 文件相对路径
	//   - error: 错误信息
	SaveResultImage(base64Data string, dir string) (string, error)

	// DeleteFile 删除文件
	// 参数：
	//   - path: 文件相对路径
	// 返回：
	//   - error: 错误信息
	DeleteFile(path string) error

	// ValidateFileFormat 验证文件格式
	// 参数：
	//   - file: 上传的文件
	//   - allowedExts: 允许的扩展名列表（如 []string{".obj", ".fbx"}）
	// 返回：
	//   - error: 格式不支持时返回错误
	ValidateFileFormat(file *multipart.FileHeader, allowedExts []string) error

	// ValidateFileSize 验证文件大小
	// 参数：
	//   - file: 上传的文件
	//   - maxSize: 最大文件大小（字节）
	// 返回：
	//   - error: 超过大小限制时返回错误
	ValidateFileSize(file *multipart.FileHeader, maxSize int64) error
}
