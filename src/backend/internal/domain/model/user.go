// Package model 定义领域模型（数据库实体）
// 包含所有业务实体的结构定义
package model

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

// User 用户实体
// 对应数据库 users 表
type User struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`                // 用户ID（主键）
	Username  string    `gorm:"type:varchar(50);uniqueIndex;not null" json:"username"` // 用户名（唯一）
	Password  string    `gorm:"type:varchar(255);not null" json:"-"`               // 密码（bcrypt加密，不返回给前端）
	RealName  string    `gorm:"type:varchar(50);not null" json:"real_name"`        // 真实姓名
	Phone     string    `gorm:"type:varchar(20)" json:"phone"`                     // 手机号
	Email     string    `gorm:"type:varchar(100);uniqueIndex;not null" json:"email"` // 邮箱（唯一）
	Role      string    `gorm:"type:varchar(20);default:'user'" json:"role"`       // 角色: user/admin
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`                  // 创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`                  // 更新时间
}

// TableName 指定表名
// GORM 会使用这个方法返回的表名，而不是默认的复数形式
func (User) TableName() string {
	return "users"
}

// IsAdmin 判断是否为管理员
// 返回值：
//   - bool: true 表示是管理员，false 表示普通用户
func (u *User) IsAdmin() bool {
	return u.Role == "admin"
}

// Bridge 桥梁实体
// 对应数据库 bridges 表
type Bridge struct {
	ID          uint           `gorm:"primaryKey;autoIncrement" json:"id"`                    // 桥梁ID（主键）
	BridgeName  string         `gorm:"type:varchar(100);not null" json:"bridge_name"`         // 桥梁名称
	BridgeCode  string         `gorm:"type:varchar(50);uniqueIndex;not null" json:"bridge_code"` // 桥梁编号（唯一）
	Address     string         `gorm:"type:varchar(255)" json:"address"`                      // 详细地址
	Longitude   float64        `gorm:"type:decimal(10,6)" json:"longitude"`                   // 经度坐标
	Latitude    float64        `gorm:"type:decimal(10,6)" json:"latitude"`                    // 纬度坐标
	BridgeType  string         `gorm:"type:varchar(50)" json:"bridge_type"`                   // 桥梁类型（梁桥/拱桥/斜拉桥）
	BuildYear   int            `gorm:"type:int" json:"build_year"`                            // 建造年份
	Length      float64        `gorm:"type:decimal(10,2)" json:"length"`                      // 桥梁长度（米）
	Width       float64        `gorm:"type:decimal(10,2)" json:"width"`                       // 桥梁宽度（米）
	Status      string         `gorm:"type:varchar(20);default:'正常'" json:"status"`          // 桥梁状态（正常/维修中/停用）
	Model3DPath string         `gorm:"column:model_3d_path;type:varchar(255)" json:"model_3d_path"` // 3D模型文件路径
	Remark      string         `gorm:"type:text" json:"remark"`                               // 备注信息
	UserID      uint           `gorm:"not null;index" json:"user_id"`                         // 所属用户ID（外键）
	User        *User          `gorm:"foreignKey:UserID" json:"user,omitempty"`               // 关联用户（延迟加载）
	Defects     []Defect       `gorm:"foreignKey:BridgeID" json:"defects,omitempty"`          // 关联缺陷记录
	CreatedAt   time.Time      `gorm:"autoCreateTime" json:"created_at"`                      // 创建时间
	UpdatedAt   time.Time      `gorm:"autoUpdateTime" json:"updated_at"`                      // 更新时间
	DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`                                        // 软删除时间
}

// TableName 指定表名
func (Bridge) TableName() string {
	return "bridges"
}

// IsOwnedBy 判断桥梁是否属于指定用户
// 返回值：
//   - bool: true 表示属于该用户，false 表示不属于
func (b *Bridge) IsOwnedBy(userID uint) bool {
	return b.UserID == userID
}

// Drone 无人机实体
// 对应数据库 drones 表
type Drone struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`      // 无人机ID（主键）
	Name      string    `gorm:"type:varchar(100);not null" json:"name"`  // 无人机名称
	Model     string    `gorm:"type:varchar(100)" json:"model"`          // 型号
	StreamURL string    `gorm:"type:varchar(255)" json:"stream_url"`     // 视频流地址
	UserID    uint      `gorm:"index;not null" json:"user_id"`           // 所属用户ID（外键）
	User      *User     `gorm:"foreignKey:UserID" json:"user,omitempty"` // 关联用户（延迟加载）
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`        // 创建时间
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`        // 更新时间
}

// TableName 指定表名
func (Drone) TableName() string {
	return "drones"
}

// IsOwnedBy 判断无人机是否属于指定用户
// 返回值：
//   - bool: true 表示属于该用户，false 表示不属于
func (d *Drone) IsOwnedBy(userID uint) bool {
	return d.UserID == userID
}

// Defect 缺陷检测结果实体
// 对应数据库 defects 表
type Defect struct {
	ID         uint           `gorm:"primaryKey;autoIncrement" json:"id"`                // 缺陷ID（主键）
	BridgeID   uint           `gorm:"index;not null" json:"bridge_id"`                   // 所属桥梁ID（外键）
	Bridge     *Bridge        `gorm:"foreignKey:BridgeID" json:"bridge,omitempty"`       // 关联桥梁（延迟加载）
	DefectType string         `gorm:"type:varchar(50);not null" json:"defect_type"`      // 缺陷类型（如：裂缝、剥落等）
	ImagePath  string         `gorm:"type:varchar(255);not null" json:"image_path"`      // 原始图片路径
	ResultPath string         `gorm:"type:varchar(255)" json:"result_path"`              // 检测结果图片路径
	BBox       string         `gorm:"type:text" json:"bbox"`                             // 边界框坐标（JSON格式）
	Length     float64        `gorm:"type:decimal(10,4)" json:"length"`                  // 缺陷长度（米）
	Width      float64        `gorm:"type:decimal(10,4)" json:"width"`                   // 缺陷宽度（米）
	Area       float64        `gorm:"type:decimal(10,4)" json:"area"`                    // 缺陷面积（平方米）
	Confidence float64        `gorm:"type:decimal(5,4)" json:"confidence"`               // 置信度（0-1）
	DetectedAt time.Time      `gorm:"type:datetime;index;not null" json:"detected_at"`   // 检测时间
	CreatedAt  time.Time      `gorm:"autoCreateTime" json:"created_at"`                  // 创建时间
	DeletedAt  gorm.DeletedAt `gorm:"index" json:"-"`                                    // 软删除时间
}

// TableName 指定表名
func (Defect) TableName() string {
	return "defects"
}

// ReportType 报表类型
type ReportType string

const (
	ReportTypeBridgeInspection  ReportType = "bridge_inspection"  // 桥梁检测报告
	ReportTypeDefectAnalysis    ReportType = "defect_analysis"    // 缺陷分析报告
	ReportTypeHealthComparison  ReportType = "health_comparison"  // 健康对比报告
)

// ReportStatus 报表状态
type ReportStatus string

const (
	ReportStatusGenerating ReportStatus = "generating" // 生成中
	ReportStatusCompleted  ReportStatus = "completed"  // 已完成
	ReportStatusFailed     ReportStatus = "failed"     // 失败
)

// BridgeIDs 桥梁ID列表（用于JSON序列化）
// 用于多桥梁对比报表
type BridgeIDs []uint

// Scan 实现sql.Scanner接口，从数据库读取JSON数据
func (b *BridgeIDs) Scan(value interface{}) error {
	if value == nil {
		*b = nil
		return nil
	}

	// 类型断言为[]byte
	bytes, ok := value.([]byte)
	if !ok {
		return gorm.ErrInvalidData
	}

	// 解析JSON
	var ids []uint
	if err := json.Unmarshal(bytes, &ids); err != nil {
		return err
	}

	*b = ids
	return nil
}

// Value 实现driver.Valuer接口，将数据写入数据库
func (b BridgeIDs) Value() (driver.Value, error) {
	if len(b) == 0 {
		return nil, nil
	}

	// 序列化为JSON
	bytes, err := json.Marshal(b)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// Report 报表实体
// 对应数据库 reports 表
type Report struct {
	ID            uint           `gorm:"primaryKey;autoIncrement" json:"id"`                                              // 报表ID（主键）
	ReportName    string         `gorm:"type:varchar(200);not null" json:"report_name"`                                   // 报表名称
	ReportType    ReportType     `gorm:"type:enum('bridge_inspection','defect_analysis','health_comparison');not null" json:"report_type"` // 报表类型
	UserID        uint           `gorm:"not null;index" json:"user_id"`                                                   // 创建用户ID（外键）
	BridgeID      *uint          `gorm:"index" json:"bridge_id,omitempty"`                                                // 关联桥梁ID（单桥梁报表）
	BridgeIDs     BridgeIDs      `gorm:"type:json" json:"bridge_ids,omitempty"`                                           // 关联桥梁ID列表（多桥梁报表）
	StartTime     time.Time      `gorm:"not null" json:"start_time"`                                                      // 报表开始时间
	EndTime       time.Time      `gorm:"not null" json:"end_time"`                                                        // 报表结束时间
	FilePath      string         `gorm:"type:varchar(500)" json:"file_path,omitempty"`                                    // PDF文件路径
	FileSize      int64          `json:"file_size"`                                                                       // 文件大小（字节）
	Status        ReportStatus   `gorm:"type:enum('generating','completed','failed');default:'generating';index" json:"status"` // 生成状态
	ErrorMessage  string         `gorm:"type:text" json:"error_message,omitempty"`                                        // 错误信息（失败时）
	TotalPages    int            `json:"total_pages"`                                                                     // 总页数
	DefectCount   int            `json:"defect_count"`                                                                    // 缺陷数量
	HighRiskCount int            `json:"high_risk_count"`                                                                 // 高危缺陷数量
	HealthScore   float64        `json:"health_score"`                                                                    // 健康度评分
	CreatedAt     time.Time      `gorm:"index" json:"created_at"`                                                         // 创建时间
	UpdatedAt     time.Time      `json:"updated_at"`                                                                      // 更新时间
	DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`                                                                  // 软删除时间

	// 关联
	User   *User   `gorm:"foreignKey:UserID" json:"user,omitempty"`   // 关联用户（延迟加载）
	Bridge *Bridge `gorm:"foreignKey:BridgeID" json:"bridge,omitempty"` // 关联桥梁（延迟加载）
}

// TableName 指定表名
func (Report) TableName() string {
	return "reports"
}

// IsOwnedBy 检查报表是否属于指定用户
// 返回值：
//   - bool: true 表示属于该用户，false 表示不属于
func (r *Report) IsOwnedBy(userID uint) bool {
	return r.UserID == userID
}

// IsCompleted 检查报表是否生成完成
// 返回值：
//   - bool: true 表示已完成，false 表示未完成
func (r *Report) IsCompleted() bool {
	return r.Status == ReportStatusCompleted
}

// IsFailed 检查报表是否生成失败
// 返回值：
//   - bool: true 表示失败，false 表示未失败
func (r *Report) IsFailed() bool {
	return r.Status == ReportStatusFailed
}

// MarkAsCompleted 标记报表为已完成
// 参数：
//   - filePath: PDF文件路径
//   - fileSize: 文件大小（字节）
//   - totalPages: 总页数
func (r *Report) MarkAsCompleted(filePath string, fileSize int64, totalPages int) {
	r.Status = ReportStatusCompleted
	r.FilePath = filePath
	r.FileSize = fileSize
	r.TotalPages = totalPages
	r.ErrorMessage = ""
}

// MarkAsFailed 标记报表为失败
// 参数：
//   - errorMsg: 错误信息
func (r *Report) MarkAsFailed(errorMsg string) {
	r.Status = ReportStatusFailed
	r.ErrorMessage = errorMsg
}
