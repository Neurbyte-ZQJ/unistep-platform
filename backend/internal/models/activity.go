package models

import "time"

// 活动状态常量
const (
	ActivityDraft      = "draft"       // 草稿
	ActivityPending    = "pending"     // 待审批
	ActivityRejected   = "rejected"    // 已驳回
	ActivityRegOpen    = "reg_open"    // 报名开放
	ActivityRegClosed  = "reg_closed"  // 报名截止
	ActivityInProgress = "in_progress" // 进行中
	ActivityCompleted  = "completed"   // 已完成
	ActivityArchived   = "archived"    // 已归档
)

// ClubActivity 社团活动
type ClubActivity struct {
	ID              uint       `gorm:"primarykey" json:"id"`
	ClubName        string     `gorm:"size:128;not null;index" json:"clubName"`        // 社团名称
	Title           string     `gorm:"size:255;not null" json:"title"`                 // 活动标题
	StartTime       time.Time  `gorm:"not null" json:"startTime"`                      // 开始时间
	EndTime         time.Time  `gorm:"not null" json:"endTime"`                        // 结束时间
	Location        string     `gorm:"size:255;not null" json:"location"`              // 活动地点
	Capacity        int        `gorm:"not null;check:capacity > 0" json:"capacity"`    // 容量
	Description     string     `gorm:"type:text;not null" json:"description"`          // 活动描述
	Budget          *float64   `gorm:"type:numeric(10,2);check:budget IS NULL OR budget >= 0" json:"budget"` // 预算
	Status          string     `gorm:"size:32;not null;default:draft;index" json:"status"` // 状态
	ApprovalOpinion string     `gorm:"type:text" json:"approvalOpinion"`               // 审批意见
	Summary         string     `gorm:"type:text" json:"summary"`                       // 活动总结
	CreatedBy       uint       `gorm:"not null;index" json:"createdBy"`                // 创建人
	ApprovedBy      *uint      `gorm:"index" json:"approvedBy"`                        // 审批人
	ApprovedAt      *time.Time `json:"approvedAt"`                                      // 审批时间
	CreatedAt       time.Time  `json:"createdAt"`
	UpdatedAt       time.Time  `json:"updatedAt"`

	// 关联
	Registrations []ActivityRegistration `gorm:"foreignKey:ActivityID" json:"registrations,omitempty"`
	Checkins      []ActivityCheckin      `gorm:"foreignKey:ActivityID" json:"checkins,omitempty"`
	Files         []ActivityFile         `gorm:"foreignKey:ActivityID" json:"files,omitempty"`
}

// ActivityRegistration 活动报名
type ActivityRegistration struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	ActivityID   uint       `gorm:"index;not null" json:"activityId"`
	StudentID    uint       `gorm:"not null" json:"studentId"`
	Status       string     `gorm:"size:16;not null;default:registered" json:"status"` // registered/cancelled
	RegisteredAt time.Time  `json:"registeredAt"`
	CancelledAt  *time.Time `json:"cancelledAt"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// ActivityCheckin 活动签到
type ActivityCheckin struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	ActivityID    uint      `gorm:"index;not null" json:"activityId"`
	StudentID     uint      `gorm:"not null" json:"studentId"`
	CheckinTime   time.Time `gorm:"not null;default:now" json:"checkinTime"`
	CheckinMethod string    `gorm:"size:16;not null;default:manual" json:"checkinMethod"` // qr/manual
	CreatedAt     time.Time `json:"createdAt"`
}

// ActivityFile 活动附件/图片
type ActivityFile struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	ActivityID uint      `gorm:"index;not null" json:"activityId"`
	FileName   string    `gorm:"size:255;not null" json:"fileName"`
	ObjectKey  string    `gorm:"size:255;not null" json:"objectKey"`
	URL        string    `gorm:"size:512" json:"url"`
	FileType   string    `gorm:"size:32;not null;default:image" json:"fileType"` // image/document/summary
	Size       int64     `json:"size"`
	UploadedBy uint      `json:"uploadedBy"`
	CreatedAt  time.Time `json:"createdAt"`
}
