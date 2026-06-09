package models

import "time"

// 团员发展阶段常量
const (
	StageApplicant       = "applicant"        // 入团申请人
	StageActivist        = "activist"         // 积极分子
	StageDevelopTarget   = "develop_target"   // 发展对象
	StagePoliticalReview = "political_review" // 政审备案
	StageLeagueMember    = "league_member"    // 正式团员
)

// MemberProfile 团员电子档案，覆盖团员发展全流程的核心字段
type MemberProfile struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	UserID     uint   `gorm:"index;not null" json:"userId"`                   // 关联 users 表
	Name       string `gorm:"size:64;not null" json:"name"`                   // 姓名
	StudentNo  string `gorm:"uniqueIndex;size:32;not null" json:"studentNo"`  // 学号
	Gender     string `gorm:"size:8" json:"gender"`                           // 性别
	Birthday   string `gorm:"size:16" json:"birthday"`                        // 生日 yyyy-MM-dd
	IDCard     string `gorm:"size:32" json:"idCard"`                          // 身份证号
	Nation     string `gorm:"size:32" json:"nation"`                          // 民族
	Phone      string `gorm:"size:32" json:"phone"`                           // 联系电话
	College    string `gorm:"size:64" json:"college"`                         // 学院
	Major      string `gorm:"size:64" json:"major"`                           // 专业
	ClassName  string `gorm:"size:64" json:"className"`                       // 班级
	Stage      string `gorm:"size:32;index;default:applicant" json:"stage"`   // 当前阶段
	JoinDate   string `gorm:"size:16" json:"joinDate"`                        // 入团日期
	Remark     string `gorm:"type:text" json:"remark"`                        // 备注
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`

	// 关联
	Applications     []LeagueApplication  `gorm:"foreignKey:ProfileID" json:"applications,omitempty"`
	ActivistRecords  []ActivistRecord     `gorm:"foreignKey:ProfileID" json:"activistRecords,omitempty"`
	DevelopRecords   []DevelopTargetRecord `gorm:"foreignKey:ProfileID" json:"developRecords,omitempty"`
	PoliticalRecords []PoliticalReview    `gorm:"foreignKey:ProfileID" json:"politicalRecords,omitempty"`
	Attachments      []MemberAttachment   `gorm:"foreignKey:ProfileID" json:"attachments,omitempty"`
}

// LeagueApplication 入团申请记录
type LeagueApplication struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	ProfileID   uint      `gorm:"index;not null" json:"profileId"`
	ApplyDate   string    `gorm:"size:16;not null" json:"applyDate"`   // 申请日期
	Motivation  string    `gorm:"type:text" json:"motivation"`         // 入团动机
	Introducer  string    `gorm:"size:64" json:"introducer"`           // 介绍人
	Status      string    `gorm:"size:16;default:pending" json:"status"` // pending/approved/rejected
	ReviewNote  string    `gorm:"type:text" json:"reviewNote"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ActivistRecord 积极分子培养记录
type ActivistRecord struct {
	ID          uint      `gorm:"primarykey" json:"id"`
	ProfileID   uint      `gorm:"index;not null" json:"profileId"`
	StartDate   string    `gorm:"size:16;not null" json:"startDate"`
	Trainer     string    `gorm:"size:64" json:"trainer"`     // 培养联系人
	TrainPlan   string    `gorm:"type:text" json:"trainPlan"` // 培养计划
	Evaluation  string    `gorm:"type:text" json:"evaluation"`
	Score       float32   `json:"score"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// DevelopTargetRecord 发展对象管理记录
type DevelopTargetRecord struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	ProfileID     uint      `gorm:"index;not null" json:"profileId"`
	ConfirmedDate string    `gorm:"size:16;not null" json:"confirmedDate"`
	Mentor        string    `gorm:"size:64" json:"mentor"`
	PublicityNote string    `gorm:"type:text" json:"publicityNote"` // 公示说明
	Conclusion    string    `gorm:"size:255" json:"conclusion"`
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// PoliticalReview 政审备案
type PoliticalReview struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	ProfileID     uint      `gorm:"index;not null" json:"profileId"`
	ReviewDate    string    `gorm:"size:16;not null" json:"reviewDate"`
	Reviewer      string    `gorm:"size:64" json:"reviewer"`
	FamilyMembers string    `gorm:"type:text" json:"familyMembers"` // 直系亲属情况
	Conclusion    string    `gorm:"size:255" json:"conclusion"`
	Status        string    `gorm:"size:16;default:filed" json:"status"` // filed/recheck
	CreatedAt     time.Time `json:"createdAt"`
	UpdatedAt     time.Time `json:"updatedAt"`
}

// MemberAttachment 团员档案附件（存储在 MinIO 中）
type MemberAttachment struct {
	ID         uint      `gorm:"primarykey" json:"id"`
	ProfileID  uint      `gorm:"index;not null" json:"profileId"`
	Category   string    `gorm:"size:32;not null" json:"category"` // application/activist/develop/political/other
	FileName   string    `gorm:"size:255;not null" json:"fileName"`
	ObjectKey  string    `gorm:"size:255;not null" json:"objectKey"`
	URL        string    `gorm:"size:512" json:"url"`
	Size       int64     `json:"size"`
	UploadedBy uint      `json:"uploadedBy"`
	CreatedAt  time.Time `json:"createdAt"`
}
