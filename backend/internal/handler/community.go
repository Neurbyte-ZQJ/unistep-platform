package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/models"
	"unistep-platform/backend/internal/response"
)

// CommunityHandler 学生社区与自治队伍模块处理器
type CommunityHandler struct {
	DB *gorm.DB
}

// NewCommunityHandler 创建社区队伍处理器
func NewCommunityHandler(db *gorm.DB) *CommunityHandler {
	return &CommunityHandler{DB: db}
}

// ---------- 请求结构 ----------

type TeamRequest struct {
	Name        string `json:"name" binding:"required,max=128"`
	TeamType    string `json:"teamType" binding:"required,oneof=autonomy volunteer duty"`
	Description string `json:"description"`
	Quota       int    `json:"quota"`
	Location    string `json:"location,max=255"`
	ContactInfo string `json:"contactInfo,max=255"`
}

type TeamMemberRequest struct {
	UserID    uint   `json:"userId" binding:"required"`
	Name      string `json:"name" binding:"required,max=64"`
	StudentNo string `json:"studentNo" binding:"required,max=32"`
	Role      string `json:"role" binding:"required,oneof=leader vice member trainee"`
	JoinDate  string `json:"joinDate" binding:"required"`
	TermStart string `json:"termStart"`
	TermEnd   string `json:"termEnd"`
	Remark    string `json:"remark"`
}

type DutyScheduleRequest struct {
	Date      string `json:"date" binding:"required"`
	StartTime string `json:"startTime" binding:"required"`
	EndTime   string `json:"endTime" binding:"required"`
	Location  string `json:"location,max=255"`
	MemberIDs []uint `json:"memberIds" binding:"required,min=1"`
}

type DutyCheckinRequest struct {
	UserID uint `json:"userId" binding:"required"`
}

type DutyCheckoutRequest struct {
	UserID uint `json:"userId" binding:"required"`
}

type VolunteerServiceRequest struct {
	UserID      uint    `json:"userId" binding:"required"`
	Name        string  `json:"name" binding:"required,max=64"`
	StudentNo   string  `json:"studentNo" binding:"required,max=32"`
	Title       string  `json:"title" binding:"required,max=255"`
	Date        string  `json:"date" binding:"required"`
	Hours       float64 `json:"hours" binding:"required,gt=0"`
	Category    string  `json:"category" binding:"required,max=32"`
	Description string  `json:"description"`
}

type VerifyServiceRequest struct {
	Verified bool `json:"verified"`
}

// ---------- 队伍 CRUD ----------

// ListTeams 队伍列表
func (h *CommunityHandler) ListTeams(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	query := h.DB.Model(&models.CommunityTeam{})
	if teamType := c.Query("teamType"); teamType != "" {
		query = query.Where("team_type = ?", teamType)
	}
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var items []models.CommunityTeam
	if err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "查询失败")
		return
	}

	response.OK(c, gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// CreateTeam 创建队伍
func (h *CommunityHandler) CreateTeam(c *gin.Context) {
	var req TeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userId")
	uid, _ := userID.(float64)

	team := models.CommunityTeam{
		Name:        req.Name,
		TeamType:    req.TeamType,
		Description: req.Description,
		Quota:       req.Quota,
		Location:    req.Location,
		ContactInfo: req.ContactInfo,
		Status:      "active",
		CreatedBy:   uint(uid),
	}
	if err := h.DB.Create(&team).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "创建队伍失败")
		return
	}
	response.Created(c, team)
}

// GetTeam 获取队伍详情
func (h *CommunityHandler) GetTeam(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var team models.CommunityTeam
	if err := h.DB.Preload("Members", func(db *gorm.DB) *gorm.DB {
		return db.Where("status != ?", models.MemberStatusLeft).Order("role ASC, id ASC")
	}).First(&team, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "队伍不存在")
		return
	}
	response.OK(c, team)
}

// UpdateTeam 更新队伍
func (h *CommunityHandler) UpdateTeam(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var team models.CommunityTeam
	if err := h.DB.First(&team, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "队伍不存在")
		return
	}

	var req TeamRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	updates := map[string]any{
		"name":         req.Name,
		"team_type":    req.TeamType,
		"description":  req.Description,
		"quota":        req.Quota,
		"location":     req.Location,
		"contact_info": req.ContactInfo,
	}
	if err := h.DB.Model(&team).Updates(updates).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "更新失败")
		return
	}
	h.DB.First(&team, id)
	response.OK(c, team)
}

// DeleteTeam 解散队伍
func (h *CommunityHandler) DeleteTeam(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var team models.CommunityTeam
	if err := h.DB.First(&team, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "队伍不存在")
		return
	}

	// 软删除：标记为 disbanded
	h.DB.Model(&team).Update("status", "disbanded")
	response.OK(c, gin.H{"id": id})
}

// ---------- 成员管理（纳新换届）----------

// ListTeamMembers 队伍成员列表
func (h *CommunityHandler) ListTeamMembers(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var members []models.TeamMember
	query := h.DB.Where("team_id = ?", id)
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if role := c.Query("role"); role != "" {
		query = query.Where("role = ?", role)
	}
	if err := query.Order("role ASC, id ASC").Find(&members).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "查询成员失败")
		return
	}
	response.OK(c, members)
}

// AddTeamMember 添加成员（纳新）
func (h *CommunityHandler) AddTeamMember(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var team models.CommunityTeam
	if err := h.DB.First(&team, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "队伍不存在")
		return
	}

	var req TeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	// 检查是否已在队伍中
	var existCount int64
	h.DB.Model(&models.TeamMember{}).Where("team_id = ? AND user_id = ? AND status != ?", id, req.UserID, models.MemberStatusLeft).Count(&existCount)
	if existCount > 0 {
		response.Fail(c, "ALREADY_MEMBER", "该成员已在此队伍中")
		return
	}

	// 检查编制
	if team.Quota > 0 {
		var activeCount int64
		h.DB.Model(&models.TeamMember{}).Where("team_id = ? AND status = ?", id, models.MemberStatusActive).Count(&activeCount)
		if int(activeCount) >= team.Quota {
			response.Fail(c, "QUOTA_FULL", "队伍编制已满")
			return
		}
	}

	member := models.TeamMember{
		TeamID:    id,
		UserID:    req.UserID,
		Name:      req.Name,
		StudentNo: req.StudentNo,
		Role:      req.Role,
		Status:    models.MemberStatusActive,
		JoinDate:  req.JoinDate,
		TermStart: req.TermStart,
		TermEnd:   req.TermEnd,
		Remark:    req.Remark,
	}
	if err := h.DB.Create(&member).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "添加成员失败")
		return
	}
	response.Created(c, member)
}

// UpdateTeamMember 更新成员信息（换届等）
func (h *CommunityHandler) UpdateTeamMember(c *gin.Context) {
	memberID, err := parseID(c.Param("memberId"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var member models.TeamMember
	if err := h.DB.First(&member, memberID).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "成员不存在")
		return
	}

	var req TeamMemberRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	updates := map[string]any{
		"role":       req.Role,
		"term_start": req.TermStart,
		"term_end":   req.TermEnd,
		"remark":     req.Remark,
	}
	if err := h.DB.Model(&member).Updates(updates).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "更新成员失败")
		return
	}
	h.DB.First(&member, memberID)
	response.OK(c, member)
}

// RemoveTeamMember 移除成员（换届退出）
func (h *CommunityHandler) RemoveTeamMember(c *gin.Context) {
	memberID, err := parseID(c.Param("memberId"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var member models.TeamMember
	if err := h.DB.First(&member, memberID).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "成员不存在")
		return
	}

	now := time.Now().Format("2006-01-02")
	h.DB.Model(&member).Updates(map[string]any{
		"status":     models.MemberStatusLeft,
		"leave_date": now,
	})
	response.OK(c, gin.H{"id": memberID})
}

// ---------- 值班管理 ----------

// ListDutySchedules 值班安排列表
func (h *CommunityHandler) ListDutySchedules(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	query := h.DB.Model(&models.DutySchedule{}).Where("team_id = ?", id)
	if date := c.Query("date"); date != "" {
		query = query.Where("date = ?", date)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var total int64
	query.Count(&total)

	var items []models.DutySchedule
	if err := query.Preload("Records").Order("date DESC, start_time ASC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "查询值班安排失败")
		return
	}

	response.OK(c, gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// CreateDutySchedule 创建值班安排
func (h *CommunityHandler) CreateDutySchedule(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var team models.CommunityTeam
	if err := h.DB.First(&team, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "队伍不存在")
		return
	}

	var req DutyScheduleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userId")
	uid, _ := userID.(float64)

	schedule := models.DutySchedule{
		TeamID:    id,
		Date:      req.Date,
		StartTime: req.StartTime,
		EndTime:   req.EndTime,
		Location:  req.Location,
		Status:    models.DutyStatusScheduled,
		CreatedBy: uint(uid),
	}
	if err := h.DB.Create(&schedule).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "创建值班安排失败")
		return
	}

	// 为每个成员创建值班记录
	for _, memberID := range req.MemberIDs {
		var member models.TeamMember
		if err := h.DB.First(&member, memberID).Error; err != nil {
			continue
		}
		record := models.DutyRecord{
			ScheduleID: schedule.ID,
			TeamID:     id,
			UserID:     member.UserID,
			Name:       member.Name,
			Status:     models.DutyStatusScheduled,
		}
		h.DB.Create(&record)
	}

	h.DB.Preload("Records").First(&schedule, schedule.ID)
	response.Created(c, schedule)
}

// DutyCheckin 值班签到
func (h *CommunityHandler) DutyCheckin(c *gin.Context) {
	scheduleID, err := parseID(c.Param("scheduleId"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var req DutyCheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	var record models.DutyRecord
	if err := h.DB.Where("schedule_id = ? AND user_id = ?", scheduleID, req.UserID).First(&record).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "未找到值班记录")
		return
	}

	if record.CheckinTime != nil {
		response.Fail(c, "ALREADY_CHECKED_IN", "已签到")
		return
	}

	now := time.Now()
	h.DB.Model(&record).Updates(map[string]any{
		"checkin_time": &now,
		"status":       models.DutyStatusActive,
	})
	h.DB.First(&record, record.ID)
	response.OK(c, record)
}

// DutyCheckout 值班签退（自动计算时长）
func (h *CommunityHandler) DutyCheckout(c *gin.Context) {
	scheduleID, err := parseID(c.Param("scheduleId"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var req DutyCheckoutRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	var record models.DutyRecord
	if err := h.DB.Where("schedule_id = ? AND user_id = ?", scheduleID, req.UserID).First(&record).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "未找到值班记录")
		return
	}

	if record.CheckinTime == nil {
		response.Fail(c, "NOT_CHECKED_IN", "尚未签到")
		return
	}

	if record.CheckoutTime != nil {
		response.Fail(c, "ALREADY_CHECKED_OUT", "已签退")
		return
	}

	now := time.Now()
	// 自动计算时长（小时）
	duration := now.Sub(*record.CheckinTime).Hours()

	h.DB.Model(&record).Updates(map[string]any{
		"checkout_time": &now,
		"duration":      duration,
		"status":        models.DutyStatusCompleted,
	})
	h.DB.First(&record, record.ID)
	response.OK(c, record)
}

// ---------- 志愿服务 ----------

// ListVolunteerServices 志愿服务列表
func (h *CommunityHandler) ListVolunteerServices(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	query := h.DB.Model(&models.VolunteerService{}).Where("team_id = ?", id)
	if category := c.Query("category"); category != "" {
		query = query.Where("category = ?", category)
	}
	if verified := c.Query("verified"); verified != "" {
		query = query.Where("verified = ?", verified == "true")
	}

	var total int64
	query.Count(&total)

	var items []models.VolunteerService
	if err := query.Order("id DESC").Offset((page - 1) * size).Limit(size).Find(&items).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "查询志愿服务失败")
		return
	}

	response.OK(c, gin.H{
		"items": items,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// CreateVolunteerService 记录志愿服务
func (h *CommunityHandler) CreateVolunteerService(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var team models.CommunityTeam
	if err := h.DB.First(&team, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "队伍不存在")
		return
	}

	var req VolunteerServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userId")
	uid, _ := userID.(float64)

	service := models.VolunteerService{
		TeamID:      id,
		UserID:      req.UserID,
		Name:        req.Name,
		StudentNo:   req.StudentNo,
		Title:       req.Title,
		Date:        req.Date,
		Hours:       req.Hours,
		Category:    req.Category,
		Description: req.Description,
		Verified:    false,
		CreatedBy:   uint(uid),
	}
	if err := h.DB.Create(&service).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "记录志愿服务失败")
		return
	}
	response.Created(c, service)
}

// VerifyVolunteerService 核实志愿服务
func (h *CommunityHandler) VerifyVolunteerService(c *gin.Context) {
	serviceID, err := parseID(c.Param("serviceId"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var service models.VolunteerService
	if err := h.DB.First(&service, serviceID).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "服务记录不存在")
		return
	}

	var req VerifyServiceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userId")
	uid, _ := userID.(float64)
	now := time.Now()

	h.DB.Model(&service).Updates(map[string]any{
		"verified":    req.Verified,
		"verified_by": uint(uid),
		"verified_at": &now,
	})
	h.DB.First(&service, serviceID)
	response.OK(c, service)
}

// ---------- 统计 ----------

// TeamStatistics 队伍统计
func (h *CommunityHandler) TeamStatistics(c *gin.Context) {
	var totalTeams int64
	h.DB.Model(&models.CommunityTeam{}).Where("status = ?", "active").Count(&totalTeams)

	// 按类型统计
	type TypeCount struct {
		TeamType string `json:"teamType"`
		Count    int64  `json:"count"`
	}
	var typeCounts []TypeCount
	h.DB.Model(&models.CommunityTeam{}).Select("team_type, count(*) as count").Where("status = ?", "active").Group("team_type").Scan(&typeCounts)

	// 总成员数
	var totalMembers int64
	h.DB.Model(&models.TeamMember{}).Where("status = ?", models.MemberStatusActive).Count(&totalMembers)

	// 总志愿服务时长
	var totalServiceHours float64
	h.DB.Model(&models.VolunteerService{}).Where("verified = ?", true).Select("COALESCE(SUM(hours), 0)").Scan(&totalServiceHours)

	// 总值班时长
	var totalDutyHours float64
	h.DB.Model(&models.DutyRecord{}).Where("status = ?", models.DutyStatusCompleted).Select("COALESCE(SUM(duration), 0)").Scan(&totalDutyHours)

	response.OK(c, gin.H{
		"totalTeams":        totalTeams,
		"typeBreakdown":     typeCounts,
		"totalMembers":      totalMembers,
		"totalServiceHours": totalServiceHours,
		"totalDutyHours":    totalDutyHours,
	})
}

// ServiceProfile 服务时长个人档案
func (h *CommunityHandler) ServiceProfile(c *gin.Context) {
	userID, _ := c.Get("userId")
	uid, _ := userID.(float64)
	targetUserID := uint(uid)

	// 支持查看他人档案
	if idStr := c.Query("userId"); idStr != "" {
		if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
			targetUserID = uint(id)
		}
	}

	// 志愿服务总时长（已核实）
	var serviceHours float64
	h.DB.Model(&models.VolunteerService{}).
		Where("user_id = ? AND verified = ?", targetUserID, true).
		Select("COALESCE(SUM(hours), 0)").Scan(&serviceHours)

	// 值班总时长
	var dutyHours float64
	h.DB.Model(&models.DutyRecord{}).
		Where("user_id = ? AND status = ?", targetUserID, models.DutyStatusCompleted).
		Select("COALESCE(SUM(duration), 0)").Scan(&dutyHours)

	// 志愿服务明细
	var services []models.VolunteerService
	h.DB.Where("user_id = ?", targetUserID).Order("date DESC").Find(&services)

	// 值班记录明细
	var dutyRecords []models.DutyRecord
	h.DB.Where("user_id = ? AND status = ?", targetUserID, models.DutyStatusCompleted).Order("id DESC").Find(&dutyRecords)

	// 所在队伍
	var memberships []models.TeamMember
	h.DB.Where("user_id = ? AND status = ?", targetUserID, models.MemberStatusActive).Find(&memberships)

	response.OK(c, gin.H{
		"userId":              targetUserID,
		"totalServiceHours":   serviceHours,
		"totalDutyHours":      dutyHours,
		"totalHours":          serviceHours + dutyHours,
		"services":            services,
		"dutyRecords":         dutyRecords,
		"teamMemberships":     memberships,
	})
}
