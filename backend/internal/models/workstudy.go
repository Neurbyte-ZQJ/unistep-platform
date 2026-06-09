package models

import "time"

// 岗位状态常量
const (
	JobDraft     = "draft"     // 草稿
	JobPublished = "published" // 已发布
	JobClosed    = "closed"    // 已关闭
	JobCompleted = "completed" // 已完成
)

// 报名状态
const (
	AppApplied   = "applied"   // 已报名
	AppAccepted  = "accepted"  // 已录用
	AppRejected  = "rejected"  // 已拒绝
	AppCancelled = "cancelled" // 已取消
)

// 薪资状态
const (
	SalaryPending   = "pending"   // 待发放
	SalaryPaid      = "paid"      // 已发放
	SalaryCancelled = "cancelled" // 已取消
)

// WorkStudyJob 勤工助学岗位
type WorkStudyJob struct {
	ID            uint       `gorm:"primarykey" json:"id"`
	Title         string     `gorm:"size:255;not null" json:"title"`                     // 岗位名称
	Department    string     `gorm:"size:128;not null;index" json:"department"`          // 用工部门
	Location      string     `gorm:"size:255;not null" json:"location"`                  // 工作地点
	Description   string     `gorm:"type:text;not null" json:"description"`              // 岗位描述
	Quota         int        `gorm:"not null;check:quota > 0" json:"quota"`              // 招聘人数
	SalaryPerHour float64    `gorm:"type:numeric(10,2);not null" json:"salaryPerHour"`   // 时薪(元)
	StartTime     time.Time  `gorm:"not null" json:"startTime"`                          // 开始时间
	EndTime       time.Time  `gorm:"not null" json:"endTime"`                            // 结束时间
	ContactPerson string     `gorm:"size:64;not null" json:"contactPerson"`              // 联系人
	ContactPhone  string     `gorm:"size:20;not null" json:"contactPhone"`               // 联系电话
	Status        string     `gorm:"size:32;not null;default:draft;index" json:"status"` // 状态
	CreatedBy     uint       `gorm:"not null;index" json:"createdBy"`                    // 创建人
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`

	// 关联
	Applications []JobApplication `gorm:"foreignKey:JobID" json:"applications,omitempty"`
	Attendances  []WorkAttendance `gorm:"foreignKey:JobID" json:"attendances,omitempty"`
	Salaries     []SalaryRecord   `gorm:"foreignKey:JobID" json:"salaries,omitempty"`
	Files        []WorkStudyFile  `gorm:"foreignKey:JobID" json:"files,omitempty"`
}

// JobApplication 岗位报名
type JobApplication struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	JobID       uint       `gorm:"index;not null" json:"jobId"`
	StudentID   uint       `gorm:"not null;index" json:"studentId"`
	Status      string     `gorm:"size:16;not null;default:applied" json:"status"` // applied/accepted/rejected/cancelled
	Remark      string     `gorm:"type:text" json:"remark"`                        // 备注
	AppliedAt   time.Time  `json:"appliedAt"`
	AcceptedAt  *time.Time `json:"acceptedAt"`
	RejectedAt  *time.Time `json:"rejectedAt"`
	CancelledAt *time.Time `json:"cancelledAt"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// WorkAttendance 考勤记录
type WorkAttendance struct {
	ID           uint       `gorm:"primarykey" json:"id"`
	JobID        uint       `gorm:"index;not null" json:"jobId"`
	StudentID    uint       `gorm:"not null;index" json:"studentId"`
	Date         string     `gorm:"size:10;not null" json:"date"`                    // YYYY-MM-DD
	CheckinTime  time.Time  `gorm:"not null" json:"checkinTime"`
	CheckoutTime *time.Time `json:"checkoutTime"`
	Hours        float64    `gorm:"type:numeric(4,1);default:0" json:"hours"`        // 工时
	Method       string     `gorm:"size:16;not null;default:manual" json:"method"`   // qr/manual
	Remark       string     `gorm:"type:text" json:"remark"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
}

// SalaryRecord 薪资记录
type SalaryRecord struct {
	ID        uint       `gorm:"primarykey" json:"id"`
	JobID     uint       `gorm:"index;not null" json:"jobId"`
	StudentID uint       `gorm:"not null;index" json:"studentId"`
	Month     string     `gorm:"size:7;not null" json:"month"`                    // YYYY-MM
	Hours     float64    `gorm:"type:numeric(8,1);not null" json:"hours"`
	Amount    float64    `gorm:"type:numeric(10,2);not null" json:"amount"`
	Status    string     `gorm:"size:16;not null;default:pending" json:"status"` // pending/paid/cancelled
	PaidAt    *time.Time `json:"paidAt"`
	Remark    string     `gorm:"type:text" json:"remark"`
	CreatedAt time.Time  `json:"createdAt"`
	UpdatedAt time.Time  `json:"updatedAt"`
}

// WorkStudyFile 岗位附件
type WorkStudyFile struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	JobID      uint      `gorm:"index;not null" json:"jobId"`
	FileName   string    `gorm:"size:255;not null" json:"fileName"`
	ObjectKey  string    `gorm:"size:255;not null" json:"objectKey"`
	URL        string    `gorm:"size:512" json:"url"`
	FileType   string    `gorm:"size:32;not null;default:image" json:"fileType"` // image/document
	Size       int64     `json:"size"`
	UploadedBy uint      `json:"uploadedBy"`
	CreatedAt  time.Time `json:"createdAt"`
}
