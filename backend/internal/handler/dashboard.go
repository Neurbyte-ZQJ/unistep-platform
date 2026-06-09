package handler

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/models"
	"unistep-platform/backend/internal/response"
)

// DashboardHandler 统计分析仪表盘处理器
type DashboardHandler struct {
	DB *gorm.DB
}

// NewDashboardHandler 创建仪表盘处理器
func NewDashboardHandler(db *gorm.DB) *DashboardHandler {
	return &DashboardHandler{DB: db}
}

// Overview 返回全平台统计概览
func (h *DashboardHandler) Overview(c *gin.Context) {
	// ---- 团员人数统计 ----
	var totalMembers int64
	h.DB.Model(&models.MemberProfile{}).Count(&totalMembers)

	type StageCount struct {
		Stage string `json:"stage"`
		Count int64  `json:"count"`
	}
	var stageCounts []StageCount
	h.DB.Model(&models.MemberProfile{}).
		Select("stage, count(*) as count").
		Group("stage").
		Scan(&stageCounts)

	// ---- 活动统计 ----
	var totalActivities int64
	h.DB.Model(&models.ClubActivity{}).Count(&totalActivities)

	type ActivityStatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var activityStatusCounts []ActivityStatusCount
	h.DB.Model(&models.ClubActivity{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&activityStatusCounts)

	var totalRegistrations int64
	h.DB.Model(&models.ActivityRegistration{}).Where("status = ?", "registered").Count(&totalRegistrations)

	var totalCheckins int64
	h.DB.Model(&models.ActivityCheckin{}).Count(&totalCheckins)

	// ---- 服务时长统计 ----
	var totalServiceHours float64
	h.DB.Model(&models.VolunteerService{}).
		Where("verified = ?", true).
		Select("COALESCE(SUM(hours), 0)").
		Scan(&totalServiceHours)

	var totalDutyHours float64
	h.DB.Model(&models.DutyRecord{}).
		Where("status = ?", models.DutyStatusCompleted).
		Select("COALESCE(SUM(duration), 0)").
		Scan(&totalDutyHours)

	var totalVolunteerCount int64
	h.DB.Model(&models.VolunteerService{}).Where("verified = ?", true).Count(&totalVolunteerCount)

	var totalDutyRecordCount int64
	h.DB.Model(&models.DutyRecord{}).Where("status = ?", models.DutyStatusCompleted).Count(&totalDutyRecordCount)

	// ---- 勤工助学岗位统计 ----
	var totalJobs int64
	h.DB.Model(&models.WorkStudyJob{}).Count(&totalJobs)

	type JobStatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var jobStatusCounts []JobStatusCount
	h.DB.Model(&models.WorkStudyJob{}).
		Select("status, count(*) as count").
		Group("status").
		Scan(&jobStatusCounts)

	var totalJobApplications int64
	h.DB.Model(&models.JobApplication{}).Count(&totalJobApplications)

	var totalAccepted int64
	h.DB.Model(&models.JobApplication{}).Where("status = ?", models.AppAccepted).Count(&totalAccepted)

	var totalSalaryPaid float64
	h.DB.Model(&models.SalaryRecord{}).
		Where("status = ?", models.SalaryPaid).
		Select("COALESCE(SUM(amount), 0)").
		Scan(&totalSalaryPaid)

	var totalWorkHours float64
	h.DB.Model(&models.WorkAttendance{}).
		Where("checkout_time IS NOT NULL").
		Select("COALESCE(SUM(hours), 0)").
		Scan(&totalWorkHours)

	response.OK(c, gin.H{
		"members": gin.H{
			"total":         totalMembers,
			"stageBreakdown": stageCounts,
		},
		"activities": gin.H{
			"total":            totalActivities,
			"statusBreakdown":  activityStatusCounts,
			"totalRegistrations": totalRegistrations,
			"totalCheckins":    totalCheckins,
		},
		"services": gin.H{
			"totalServiceHours":    totalServiceHours,
			"totalDutyHours":       totalDutyHours,
			"totalHours":           totalServiceHours + totalDutyHours,
			"totalVolunteerCount":  totalVolunteerCount,
			"totalDutyRecordCount": totalDutyRecordCount,
		},
		"workstudy": gin.H{
			"totalJobs":           totalJobs,
			"statusBreakdown":     jobStatusCounts,
			"totalApplications":   totalJobApplications,
			"totalAccepted":       totalAccepted,
			"totalSalaryPaid":     totalSalaryPaid,
			"totalWorkHours":      totalWorkHours,
		},
	})
}

// MemberTrend 团员发展趋势（按月统计新增人数）
func (h *DashboardHandler) MemberTrend(c *gin.Context) {
	type MonthCount struct {
		Month string `json:"month"`
		Count int64  `json:"count"`
	}
	var monthly []MonthCount
	h.DB.Model(&models.MemberProfile{}).
		Select("strftime('%Y-%m', created_at) as month, count(*) as count").
		Group("month").
		Order("month").
		Scan(&monthly)

	response.OK(c, gin.H{
		"trend": monthly,
	})
}

// ActivityTrend 活动趋势（按月统计）
func (h *DashboardHandler) ActivityTrend(c *gin.Context) {
	type MonthCount struct {
		Month string `json:"month"`
		Count int64  `json:"count"`
	}
	var monthly []MonthCount
	h.DB.Model(&models.ClubActivity{}).
		Select("strftime('%Y-%m', created_at) as month, count(*) as count").
		Group("month").
		Order("month").
		Scan(&monthly)

	response.OK(c, gin.H{
		"trend": monthly,
	})
}

// ServiceTrend 服务时长趋势（按月统计）
func (h *DashboardHandler) ServiceTrend(c *gin.Context) {
	type MonthHours struct {
		Month string  `json:"month"`
		Hours float64 `json:"hours"`
	}
	var volunteerMonthly []MonthHours
	h.DB.Model(&models.VolunteerService{}).
		Where("verified = ?", true).
		Select("strftime('%Y-%m', date) as month, COALESCE(SUM(hours), 0) as hours").
		Group("month").
		Order("month").
		Scan(&volunteerMonthly)

	var dutyMonthly []MonthHours
	h.DB.Model(&models.DutyRecord{}).
		Where("status = ?", models.DutyStatusCompleted).
		Select("strftime('%Y-%m', checkin_time) as month, COALESCE(SUM(duration), 0) as hours").
		Group("month").
		Order("month").
		Scan(&dutyMonthly)

	response.OK(c, gin.H{
		"volunteer": volunteerMonthly,
		"duty":      dutyMonthly,
	})
}
