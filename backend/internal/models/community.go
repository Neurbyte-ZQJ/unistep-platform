package models

import "time"

// 队伍类型常量
const (
	TeamTypeAutonomy  = "autonomy"  // 自治队伍
	TeamTypeVolunteer = "volunteer" // 志愿服务队
	TeamTypeDuty      = "duty"      // 值班队伍
)

// 成员角色常量
const (
	MemberRoleLeader   = "leader"   // 负责人
	MemberRoleVice     = "vice"     // 副负责人
	MemberRoleMember   = "member"   // 成员
	MemberRoleTrainee  = "trainee"  // 实习成员
	MemberRoleRetired  = "retired"  // 已换届退出
)

// 成员状态常量
const (
	MemberStatusActive  = "active"  // 在职
	MemberStatusPending = "pending" // 待审核
	MemberStatusLeft    = "left"    // 已退出
)

// 值班状态常量
const (
	DutyStatusScheduled = "scheduled" // 已排班
	DutyStatusActive    = "active"    // 值班中
	DutyStatusCompleted = "completed" // 已完成
	DutyStatusAbsent    = "absent"    // 缺勤
)

// CommunityTeam 学生社区/自治队伍
type CommunityTeam struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	Name        string     `gorm:"size:128;not null;index" json:"name"`              // 队伍名称
	TeamType    string     `gorm:"size:32;not null;index" json:"teamType"`           // 队伍类型
	Description string     `gorm:"type:text" json:"description"`                     // 队伍简介
	Quota       int        `gorm:"not null;default:0" json:"quota"`                  // 编制人数
	Location    string     `gorm:"size:255" json:"location"`                         // 活动地点/值班室
	ContactInfo string     `gorm:"size:255" json:"contactInfo"`                      // 联系方式
	Status      string     `gorm:"size:16;not null;default:active;index" json:"status"` // active/disbanded
	CreatedBy   uint       `gorm:"not null;index" json:"createdBy"`                  // 创建人
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`

	// 关联
	Members []TeamMember `gorm:"foreignKey:TeamID" json:"members,omitempty"`
}

// TeamMember 队伍成员（含纳新换届）
type TeamMember struct {
	ID         uint       `gorm:"primarykey" json:"id"`
	TeamID     uint       `gorm:"index;not null" json:"teamId"`
	UserID     uint       `gorm:"index;not null" json:"userId"`
	Name       string     `gorm:"size:64;not null" json:"name"`                     // 姓名
	StudentNo  string     `gorm:"size:32;not null" json:"studentNo"`                // 学号
	Role       string     `gorm:"size:32;not null;default:member" json:"role"`      // 角色
	Status     string     `gorm:"size:16;not null;default:active" json:"status"`    // 状态
	JoinDate   string     `gorm:"size:32;not null" json:"joinDate"`                 // 入队日期
	LeaveDate  *string    `gorm:"size:32" json:"leaveDate"`                         // 离队日期
	TermStart  string     `gorm:"size:32" json:"termStart"`                         // 届次开始
	TermEnd    string     `gorm:"size:32" json:"termEnd"`                           // 届次结束
	Remark     string     `gorm:"type:text" json:"remark"`                          // 备注
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// DutySchedule 值班安排
type DutySchedule struct {
	ID        uint      `gorm:"primarykey" json:"id"`
	TeamID    uint      `gorm:"index;not null" json:"teamId"`
	Date      string    `gorm:"size:10;not null;index" json:"date"`               // 值班日期 yyyy-MM-dd
	StartTime string    `gorm:"size:5;not null" json:"startTime"`                 // 开始时间 HH:mm
	EndTime   string    `gorm:"size:5;not null" json:"endTime"`                   // 结束时间 HH:mm
	Location  string    `gorm:"size:255" json:"location"`                         // 值班地点
	Status    string    `gorm:"size:16;not null;default:scheduled" json:"status"` // 值班状态
	CreatedBy uint      `gorm:"not null" json:"createdBy"`                        // 排班人
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// 关联
	Records []DutyRecord `gorm:"foreignKey:ScheduleID" json:"records,omitempty"`
}

// DutyRecord 值班记录（签到/签退）
type DutyRecord struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	ScheduleID  uint       `gorm:"index;not null" json:"scheduleId"`
	TeamID      uint       `gorm:"index;not null" json:"teamId"`
	UserID      uint       `gorm:"index;not null" json:"userId"`
	Name        string     `gorm:"size:64;not null" json:"name"`                     // 姓名
	CheckinTime *time.Time `json:"checkinTime"`                                       // 签到时间
	CheckoutTime *time.Time `json:"checkoutTime"`                                     // 签退时间
	Duration    *float64   `json:"duration"`                                          // 时长(小时)，签退时自动计算
	Status      string     `gorm:"size:16;not null;default=scheduled" json:"status"` // scheduled/completed/absent
	Remark      string     `gorm:"type:text" json:"remark"`                          // 备注
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// VolunteerService 志愿服务记录
type VolunteerService struct {
	ID          uint       `gorm:"primarykey" json:"id"`
	TeamID      uint       `gorm:"index;not null" json:"teamId"`
	UserID      uint       `gorm:"index;not null" json:"userId"`
	Name        string     `gorm:"size:64;not null" json:"name"`                     // 姓名
	StudentNo   string     `gorm:"size:32;not null" json:"studentNo"`                // 学号
	Title       string     `gorm:"size:255;not null" json:"title"`                   // 服务名称
	Date        string     `gorm:"size:10;not null;index" json:"date"`               // 服务日期
	Hours       float64    `gorm:"not null;check:hours > 0" json:"hours"`            // 服务时长(小时)
	Category    string     `gorm:"size:32;not null" json:"category"`                 // 服务类别
	Description string     `gorm:"type:text" json:"description"`                     // 服务描述
	Verified    bool       `gorm:"not null;default:false" json:"verified"`           // 是否已核实
	VerifiedBy  *uint      `json:"verifiedBy"`                                        // 核实人
	VerifiedAt  *time.Time `json:"verifiedAt"`                                        // 核实时间
	CreatedBy   uint       `gorm:"not null" json:"createdBy"`                        // 记录人
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}
