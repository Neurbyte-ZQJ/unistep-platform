package handler

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/models"
	"unistep-platform/backend/internal/response"
)

// ActivityHandler 社团活动模块处理器
type ActivityHandler struct {
	DB       *gorm.DB
	Uploader Uploader
}

// NewActivityHandler 创建社团活动处理器
func NewActivityHandler(db *gorm.DB, uploader Uploader) *ActivityHandler {
	return &ActivityHandler{DB: db, Uploader: uploader}
}

// ---------- 请求结构 ----------

type ActivityRequest struct {
	ClubName    string   `json:"clubName" binding:"required,max=128"`
	Title       string   `json:"title" binding:"required,max=255"`
	StartTime   string   `json:"startTime" binding:"required"`
	EndTime     string   `json:"endTime" binding:"required"`
	Location    string   `json:"location" binding:"required,max=255"`
	Capacity    int      `json:"capacity" binding:"required,min=1"`
	Description string   `json:"description" binding:"required"`
	Budget      *float64 `json:"budget"`
}

type ApprovalRequest struct {
	Opinion string `json:"opinion" binding:"required"`
	Approve bool   `json:"approve"`
}

type SummaryRequest struct {
	Summary string `json:"summary" binding:"required"`
}

type CheckinRequest struct {
	StudentID uint `json:"studentId" binding:"required"`
}

// ---------- 活动 CRUD ----------

// ListActivities 活动列表（分页+状态/社团过滤）
func (h *ActivityHandler) ListActivities(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	query := h.DB.Model(&models.ClubActivity{})
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if clubName := c.Query("clubName"); clubName != "" {
		query = query.Where("club_name LIKE ?", "%"+clubName+"%")
	}
	if title := c.Query("title"); title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}

	var total int64
	query.Count(&total)

	var items []models.ClubActivity
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

// CreateActivity 创建活动（草稿）
func (h *ActivityHandler) CreateActivity(c *gin.Context) {
	var req ActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", "开始时间格式错误，需ISO8601")
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", "结束时间格式错误，需ISO8601")
		return
	}
	if !endTime.After(startTime) {
		response.Fail(c, "INVALID_PARAMS", "结束时间必须晚于开始时间")
		return
	}

	userID, _ := c.Get("userId")
	uid, _ := userID.(float64)

	activity := models.ClubActivity{
		ClubName:    req.ClubName,
		Title:       req.Title,
		StartTime:   startTime,
		EndTime:     endTime,
		Location:    req.Location,
		Capacity:    req.Capacity,
		Description: req.Description,
		Budget:      req.Budget,
		Status:      models.ActivityDraft,
		CreatedBy:   uint(uid),
	}
	if err := h.DB.Create(&activity).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "创建活动失败")
		return
	}
	response.Created(c, activity)
}

// GetActivity 获取活动详情
func (h *ActivityHandler) GetActivity(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var activity models.ClubActivity
	if err := h.DB.
		Preload("Registrations").
		Preload("Checkins").
		Preload("Files").
		First(&activity, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "活动不存在")
		return
	}
	response.OK(c, activity)
}

// UpdateActivity 更新活动（仅草稿/已驳回状态可编辑）
func (h *ActivityHandler) UpdateActivity(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var activity models.ClubActivity
	if err := h.DB.First(&activity, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "活动不存在")
		return
	}
	if activity.Status != models.ActivityDraft && activity.Status != models.ActivityRejected {
		response.Fail(c, "INVALID_STATUS", "仅草稿或已驳回状态可编辑")
		return
	}

	var req ActivityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", "开始时间格式错误")
		return
	}
	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", "结束时间格式错误")
		return
	}
	if !endTime.After(startTime) {
		response.Fail(c, "INVALID_PARAMS", "结束时间必须晚于开始时间")
		return
	}

	updates := map[string]any{
		"club_name":    req.ClubName,
		"title":        req.Title,
		"start_time":   startTime,
		"end_time":     endTime,
		"location":     req.Location,
		"capacity":     req.Capacity,
		"description":  req.Description,
		"budget":       req.Budget,
	}
	if err := h.DB.Model(&activity).Updates(updates).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "更新失败")
		return
	}
	h.DB.First(&activity, id)
	response.OK(c, activity)
}

// DeleteActivity 删除活动（仅草稿状态可删除）
func (h *ActivityHandler) DeleteActivity(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var activity models.ClubActivity
	if err := h.DB.First(&activity, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "活动不存在")
		return
	}
	if activity.Status != models.ActivityDraft {
		response.Fail(c, "INVALID_STATUS", "仅草稿状态可删除")
		return
	}

	h.DB.Where("activity_id = ?", id).Delete(&models.ActivityFile{})
	h.DB.Where("activity_id = ?", id).Delete(&models.ActivityCheckin{})
	h.DB.Where("activity_id = ?", id).Delete(&models.ActivityRegistration{})
	h.DB.Delete(&activity)
	response.OK(c, gin.H{"id": id})
}

// ---------- 审批流程 ----------

// SubmitForSubmit 提交审批
func (h *ActivityHandler) SubmitForApproval(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var activity models.ClubActivity
	if err := h.DB.First(&activity, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "活动不存在")
		return
	}
	if activity.Status != models.ActivityDraft && activity.Status != models.ActivityRejected {
		response.Fail(c, "INVALID_STATUS", "仅草稿或已驳回状态可提交审批")
		return
	}

	h.DB.Model(&activity).Update("status", models.ActivityPending)
	response.OK(c, activity)
}

// ApproveActivity 审批活动
func (h *ActivityHandler) ApproveActivity(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var activity models.ClubActivity
	if err := h.DB.First(&activity, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "活动不存在")
		return
	}
	if activity.Status != models.ActivityPending {
		response.Fail(c, "INVALID_STATUS", "仅待审批状态可审批")
		return
	}

	var req ApprovalRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	userID, _ := c.Get("userId")
	uid, _ := userID.(float64)
	now := time.Now()

	newStatus := models.ActivityRegOpen
	if !req.Approve {
		newStatus = models.ActivityRejected
	}

	h.DB.Model(&activity).Updates(map[string]any{
		"status":           newStatus,
		"approval_opinion": req.Opinion,
		"approved_by":      uint(uid),
		"approved_at":      &now,
	})
	h.DB.First(&activity, id)
	response.OK(c, activity)
}

// ---------- 报名与签到 ----------

// RegisterActivity 活动报名
func (h *ActivityHandler) RegisterActivity(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var activity models.ClubActivity
	if err := h.DB.First(&activity, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "活动不存在")
		return
	}
	if activity.Status != models.ActivityRegOpen {
		response.Fail(c, "INVALID_STATUS", "活动未开放报名")
		return
	}

	// 检查容量
	var regCount int64
	h.DB.Model(&models.ActivityRegistration{}).Where("activity_id = ? AND status = ?", id, "registered").Count(&regCount)
	if int(regCount) >= activity.Capacity {
		response.Fail(c, "CAPACITY_FULL", "活动名额已满")
		return
	}

	userID, _ := c.Get("userId")
	uid, _ := userID.(float64)

	// 检查重复报名
	var existCount int64
	h.DB.Model(&models.ActivityRegistration{}).Where("activity_id = ? AND student_id = ? AND status = ?", id, uint(uid), "registered").Count(&existCount)
	if existCount > 0 {
		response.Fail(c, "ALREADY_REGISTERED", "已报名该活动")
		return
	}

	registration := models.ActivityRegistration{
		ActivityID:   id,
		StudentID:    uint(uid),
		Status:       "registered",
		RegisteredAt: time.Now(),
	}
	if err := h.DB.Create(&registration).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "报名失败")
		return
	}
	response.Created(c, registration)
}

// CancelRegistration 取消报名
func (h *ActivityHandler) CancelRegistration(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	userID, _ := c.Get("userId")
	uid, _ := userID.(float64)
	now := time.Now()

	result := h.DB.Model(&models.ActivityRegistration{}).
		Where("activity_id = ? AND student_id = ? AND status = ?", id, uint(uid), "registered").
		Updates(map[string]any{
			"status":       "cancelled",
			"cancelled_at": &now,
		})
	if result.RowsAffected == 0 {
		response.Fail(c, "NOT_FOUND", "未找到有效报名记录")
		return
	}
	response.OK(c, gin.H{"activityId": id})
}

// CheckinActivity 活动签到
func (h *ActivityHandler) CheckinActivity(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var req CheckinRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	var activity models.ClubActivity
	if err := h.DB.First(&activity, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "活动不存在")
		return
	}
	if activity.Status != models.ActivityInProgress && activity.Status != models.ActivityRegOpen && activity.Status != models.ActivityRegClosed {
		response.Fail(c, "INVALID_STATUS", "活动当前状态不允许签到")
		return
	}

	// 检查是否已签到
	var existCount int64
	h.DB.Model(&models.ActivityCheckin{}).Where("activity_id = ? AND student_id = ?", id, req.StudentID).Count(&existCount)
	if existCount > 0 {
		response.Fail(c, "ALREADY_CHECKED_IN", "该学生已签到")
		return
	}

	checkin := models.ActivityCheckin{
		ActivityID:    id,
		StudentID:     req.StudentID,
		CheckinTime:   time.Now(),
		CheckinMethod: "manual",
	}
	if err := h.DB.Create(&checkin).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "签到失败")
		return
	}
	response.Created(c, checkin)
}

// ---------- 图片上传 ----------

// UploadActivityImage 上传活动图片/文件
func (h *ActivityHandler) UploadActivityImage(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var activity models.ClubActivity
	if err := h.DB.First(&activity, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "活动不存在")
		return
	}

	if h.Uploader == nil || h.Uploader.Disabled() {
		response.Fail(c, "STORAGE_DISABLED", "对象存储未启用，请配置 MINIO_ENDPOINT")
		return
	}

	fileType := c.DefaultPostForm("fileType", "image")
	fileHeader, err := c.FormFile("file")
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", "缺少文件: "+err.Error())
		return
	}

	f, err := fileHeader.Open()
	if err != nil {
		response.Fail(c, "INTERNAL_ERROR", "无法读取上传文件")
		return
	}
	defer f.Close()

	objectKey, url, err := h.Uploader.Upload(c.Request.Context(), "activities", fileHeader.Filename, fileHeader.Size, f, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		response.Fail(c, "UPLOAD_FAILED", err.Error())
		return
	}

	uploadedBy, _ := c.Get("userId")
	uploaderID, _ := uploadedBy.(float64)

	activityFile := models.ActivityFile{
		ActivityID: id,
		FileName:   fileHeader.Filename,
		ObjectKey:  objectKey,
		URL:        url,
		FileType:   fileType,
		Size:       fileHeader.Size,
		UploadedBy: uint(uploaderID),
	}
	if err := h.DB.Create(&activityFile).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "保存文件记录失败")
		return
	}
	response.Created(c, activityFile)
}

// ---------- 总结归档 ----------

// SubmitSummary 提交活动总结
func (h *ActivityHandler) SubmitSummary(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var activity models.ClubActivity
	if err := h.DB.First(&activity, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "活动不存在")
		return
	}
	if activity.Status != models.ActivityCompleted {
		response.Fail(c, "INVALID_STATUS", "仅已完成状态可提交总结")
		return
	}

	var req SummaryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	h.DB.Model(&activity).Updates(map[string]any{
		"summary": req.Summary,
		"status":  models.ActivityArchived,
	})
	h.DB.First(&activity, id)
	response.OK(c, activity)
}

// UpdateStatus 更新活动状态（用于手动推进状态）
func (h *ActivityHandler) UpdateStatus(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var body struct {
		Status string `json:"status" binding:"required"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	validStatuses := map[string]bool{
		models.ActivityDraft:      true,
		models.ActivityPending:    true,
		models.ActivityRejected:   true,
		models.ActivityRegOpen:    true,
		models.ActivityRegClosed:  true,
		models.ActivityInProgress: true,
		models.ActivityCompleted:  true,
		models.ActivityArchived:   true,
	}
	if !validStatuses[body.Status] {
		response.Fail(c, "INVALID_PARAMS", "无效的状态值")
		return
	}

	var activity models.ClubActivity
	if err := h.DB.First(&activity, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "活动不存在")
		return
	}

	h.DB.Model(&activity).Update("status", body.Status)
	h.DB.First(&activity, id)
	response.OK(c, activity)
}

// ---------- 统计 ----------

// ActivityStatistics 活动统计
func (h *ActivityHandler) ActivityStatistics(c *gin.Context) {
	var totalCount int64
	h.DB.Model(&models.ClubActivity{}).Count(&totalCount)

	// 按状态统计
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var statusCounts []StatusCount
	h.DB.Model(&models.ClubActivity{}).Select("status, count(*) as count").Group("status").Scan(&statusCounts)

	// 总报名人数
	var totalRegistrations int64
	h.DB.Model(&models.ActivityRegistration{}).Where("status = ?", "registered").Count(&totalRegistrations)

	// 总签到人数
	var totalCheckins int64
	h.DB.Model(&models.ActivityCheckin{}).Count(&totalCheckins)

	// 近期活动（最近5个）
	var recentActivities []models.ClubActivity
	h.DB.Order("created_at DESC").Limit(5).Find(&recentActivities)

	response.OK(c, gin.H{
		"totalActivities":      totalCount,
		"statusBreakdown":      statusCounts,
		"totalRegistrations":   totalRegistrations,
		"totalCheckins":        totalCheckins,
		"recentActivities":     recentActivities,
	})
}
