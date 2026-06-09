package handler

import (
	"math"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/models"
	"unistep-platform/backend/internal/response"
)

// WorkStudyHandler 勤工助学模块处理器
type WorkStudyHandler struct {
	DB       *gorm.DB
	Uploader Uploader
}

// NewWorkStudyHandler 创建勤工助学处理器
func NewWorkStudyHandler(db *gorm.DB, uploader Uploader) *WorkStudyHandler {
	return &WorkStudyHandler{DB: db, Uploader: uploader}
}

// ---------- 请求结构 ----------

type JobRequest struct {
	Title         string  `json:"title" binding:"required,max=255"`
	Department    string  `json:"department" binding:"required,max=128"`
	Location      string  `json:"location" binding:"required,max=255"`
	Description   string  `json:"description" binding:"required"`
	Quota         int     `json:"quota" binding:"required,min=1"`
	SalaryPerHour float64 `json:"salaryPerHour" binding:"required,gt=0"`
	StartTime     string  `json:"startTime" binding:"required"`
	EndTime       string  `json:"endTime" binding:"required"`
	ContactPerson string  `json:"contactPerson" binding:"required,max=64"`
	ContactPhone  string  `json:"contactPhone" binding:"required,max=20"`
}

type AttendanceRequest struct {
	StudentID uint   `json:"studentId" binding:"required"`
	Date      string `json:"date" binding:"required"`
	Method    string `json:"method"`
}

type CheckoutRequest struct {
	// empty - uses attendance ID from URL
}

type SalaryCalculateRequest struct {
	Month string `json:"month" binding:"required"` // YYYY-MM
}

type ApplicationActionRequest struct {
	Remark string `json:"remark"`
}

// ---------- 岗位 CRUD ----------

// ListJobs 岗位列表（分页+状态/部门/标题过滤）
func (h *WorkStudyHandler) ListJobs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	query := h.DB.Model(&models.WorkStudyJob{})
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if department := c.Query("department"); department != "" {
		query = query.Where("department LIKE ?", "%"+department+"%")
	}
	if title := c.Query("title"); title != "" {
		query = query.Where("title LIKE ?", "%"+title+"%")
	}

	var total int64
	query.Count(&total)

	var items []models.WorkStudyJob
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

// CreateJob 创建岗位（草稿）
func (h *WorkStudyHandler) CreateJob(c *gin.Context) {
	var req JobRequest
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

	job := models.WorkStudyJob{
		Title:         req.Title,
		Department:    req.Department,
		Location:      req.Location,
		Description:   req.Description,
		Quota:         req.Quota,
		SalaryPerHour: req.SalaryPerHour,
		StartTime:     startTime,
		EndTime:       endTime,
		ContactPerson: req.ContactPerson,
		ContactPhone:  req.ContactPhone,
		Status:        models.JobDraft,
		CreatedBy:     uint(uid),
	}
	if err := h.DB.Create(&job).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "创建岗位失败")
		return
	}
	response.Created(c, job)
}

// GetJob 获取岗位详情
func (h *WorkStudyHandler) GetJob(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var job models.WorkStudyJob
	if err := h.DB.
		Preload("Applications").
		Preload("Attendances").
		Preload("Salaries").
		Preload("Files").
		First(&job, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "岗位不存在")
		return
	}
	response.OK(c, job)
}

// UpdateJob 更新岗位（仅草稿状态可编辑）
func (h *WorkStudyHandler) UpdateJob(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var job models.WorkStudyJob
	if err := h.DB.First(&job, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "岗位不存在")
		return
	}
	if job.Status != models.JobDraft {
		response.Fail(c, "INVALID_STATUS", "仅草稿状态可编辑")
		return
	}

	var req JobRequest
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
		"title":          req.Title,
		"department":     req.Department,
		"location":       req.Location,
		"description":    req.Description,
		"quota":          req.Quota,
		"salary_per_hour": req.SalaryPerHour,
		"start_time":     startTime,
		"end_time":       endTime,
		"contact_person": req.ContactPerson,
		"contact_phone":  req.ContactPhone,
	}
	if err := h.DB.Model(&job).Updates(updates).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "更新失败")
		return
	}
	h.DB.First(&job, id)
	response.OK(c, job)
}

// DeleteJob 删除岗位（仅草稿状态可删除）
func (h *WorkStudyHandler) DeleteJob(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var job models.WorkStudyJob
	if err := h.DB.First(&job, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "岗位不存在")
		return
	}
	if job.Status != models.JobDraft {
		response.Fail(c, "INVALID_STATUS", "仅草稿状态可删除")
		return
	}

	h.DB.Where("job_id = ?", id).Delete(&models.WorkStudyFile{})
	h.DB.Where("job_id = ?", id).Delete(&models.SalaryRecord{})
	h.DB.Where("job_id = ?", id).Delete(&models.WorkAttendance{})
	h.DB.Where("job_id = ?", id).Delete(&models.JobApplication{})
	h.DB.Delete(&job)
	response.OK(c, gin.H{"id": id})
}

// ---------- 岗位状态流转 ----------

// PublishJob 发布岗位（草稿 -> 已发布）
func (h *WorkStudyHandler) PublishJob(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var job models.WorkStudyJob
	if err := h.DB.First(&job, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "岗位不存在")
		return
	}
	if job.Status != models.JobDraft {
		response.Fail(c, "INVALID_STATUS", "仅草稿状态可发布")
		return
	}

	h.DB.Model(&job).Update("status", models.JobPublished)
	h.DB.First(&job, id)
	response.OK(c, job)
}

// CloseJob 关闭岗位（已发布 -> 已关闭）
func (h *WorkStudyHandler) CloseJob(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var job models.WorkStudyJob
	if err := h.DB.First(&job, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "岗位不存在")
		return
	}
	if job.Status != models.JobPublished {
		response.Fail(c, "INVALID_STATUS", "仅已发布状态可关闭")
		return
	}

	h.DB.Model(&job).Update("status", models.JobClosed)
	h.DB.First(&job, id)
	response.OK(c, job)
}

// ---------- 报名管理 ----------

// ApplyJob 学生报名岗位
func (h *WorkStudyHandler) ApplyJob(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var job models.WorkStudyJob
	if err := h.DB.First(&job, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "岗位不存在")
		return
	}
	if job.Status != models.JobPublished {
		response.Fail(c, "INVALID_STATUS", "岗位未开放报名")
		return
	}

	// 检查名额
	var acceptedCount int64
	h.DB.Model(&models.JobApplication{}).Where("job_id = ? AND status = ?", id, models.AppAccepted).Count(&acceptedCount)
	if int(acceptedCount) >= job.Quota {
		response.Fail(c, "QUOTA_FULL", "岗位名额已满")
		return
	}

	userID, _ := c.Get("userId")
	uid, _ := userID.(float64)

	// 检查重复报名
	var existCount int64
	h.DB.Model(&models.JobApplication{}).Where("job_id = ? AND student_id = ? AND status = ?", id, uint(uid), models.AppApplied).Count(&existCount)
	if existCount > 0 {
		response.Fail(c, "ALREADY_APPLIED", "已报名该岗位")
		return
	}

	application := models.JobApplication{
		JobID:     id,
		StudentID: uint(uid),
		Status:    models.AppApplied,
		AppliedAt: time.Now(),
	}
	if err := h.DB.Create(&application).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "报名失败")
		return
	}
	response.Created(c, application)
}

// CancelApplication 学生取消报名
func (h *WorkStudyHandler) CancelApplication(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	userID, _ := c.Get("userId")
	uid, _ := userID.(float64)
	now := time.Now()

	result := h.DB.Model(&models.JobApplication{}).
		Where("job_id = ? AND student_id = ? AND status = ?", id, uint(uid), models.AppApplied).
		Updates(map[string]any{
			"status":       models.AppCancelled,
			"cancelled_at": &now,
		})
	if result.RowsAffected == 0 {
		response.Fail(c, "NOT_FOUND", "未找到有效报名记录")
		return
	}
	response.OK(c, gin.H{"jobId": id})
}

// ListApplications 岗位报名列表
func (h *WorkStudyHandler) ListApplications(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	query := h.DB.Model(&models.JobApplication{}).Where("job_id = ?", id)
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	var applications []models.JobApplication
	if err := query.Order("id DESC").Find(&applications).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "查询失败")
		return
	}
	response.OK(c, applications)
}

// AcceptApplication 录用报名
func (h *WorkStudyHandler) AcceptApplication(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}
	appID, err := parseID(c.Param("appId"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var job models.WorkStudyJob
	if err := h.DB.First(&job, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "岗位不存在")
		return
	}

	// 检查名额
	var acceptedCount int64
	h.DB.Model(&models.JobApplication{}).Where("job_id = ? AND status = ?", id, models.AppAccepted).Count(&acceptedCount)
	if int(acceptedCount) >= job.Quota {
		response.Fail(c, "QUOTA_FULL", "岗位名额已满")
		return
	}

	var application models.JobApplication
	if err := h.DB.First(&application, appID).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "报名记录不存在")
		return
	}
	if application.Status != models.AppApplied {
		response.Fail(c, "INVALID_STATUS", "仅已报名状态可录用")
		return
	}

	var req ApplicationActionRequest
	c.ShouldBindJSON(&req)

	now := time.Now()
	h.DB.Model(&application).Updates(map[string]any{
		"status":      models.AppAccepted,
		"accepted_at": &now,
		"remark":      req.Remark,
	})
	h.DB.First(&application, appID)
	response.OK(c, application)
}

// RejectApplication 拒绝报名
func (h *WorkStudyHandler) RejectApplication(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}
	appID, err := parseID(c.Param("appId"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var application models.JobApplication
	if err := h.DB.Where("id = ? AND job_id = ?", appID, id).First(&application).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "报名记录不存在")
		return
	}
	if application.Status != models.AppApplied {
		response.Fail(c, "INVALID_STATUS", "仅已报名状态可拒绝")
		return
	}

	var req ApplicationActionRequest
	c.ShouldBindJSON(&req)

	now := time.Now()
	h.DB.Model(&application).Updates(map[string]any{
		"status":      models.AppRejected,
		"rejected_at": &now,
		"remark":      req.Remark,
	})
	h.DB.First(&application, appID)
	response.OK(c, application)
}

// ---------- 考勤管理 ----------

// CreateAttendance 创建考勤记录
func (h *WorkStudyHandler) CreateAttendance(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var req AttendanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	var job models.WorkStudyJob
	if err := h.DB.First(&job, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "岗位不存在")
		return
	}
	if job.Status != models.JobPublished && job.Status != models.JobClosed {
		response.Fail(c, "INVALID_STATUS", "岗位当前状态不允许考勤")
		return
	}

	// 检查重复考勤
	var existCount int64
	h.DB.Model(&models.WorkAttendance{}).Where("job_id = ? AND student_id = ? AND date = ?", id, req.StudentID, req.Date).Count(&existCount)
	if existCount > 0 {
		response.Fail(c, "ALREADY_CHECKED_IN", "该学生当日已有考勤记录")
		return
	}

	method := req.Method
	if method == "" {
		method = "manual"
	}

	attendance := models.WorkAttendance{
		JobID:       id,
		StudentID:   req.StudentID,
		Date:        req.Date,
		CheckinTime: time.Now(),
		Method:      method,
	}
	if err := h.DB.Create(&attendance).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "创建考勤记录失败")
		return
	}
	response.Created(c, attendance)
}

// ListAttendances 考勤记录列表
func (h *WorkStudyHandler) ListAttendances(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	query := h.DB.Model(&models.WorkAttendance{}).Where("job_id = ?", id)
	if studentID := c.Query("studentId"); studentID != "" {
		query = query.Where("student_id = ?", studentID)
	}
	if date := c.Query("date"); date != "" {
		query = query.Where("date = ?", date)
	}

	var attendances []models.WorkAttendance
	if err := query.Order("id DESC").Find(&attendances).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "查询失败")
		return
	}
	response.OK(c, attendances)
}

// CheckoutAttendance 签退
func (h *WorkStudyHandler) CheckoutAttendance(c *gin.Context) {
	attID, err := parseID(c.Param("attId"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var _ CheckoutRequest // 确保结构体被引用

	var attendance models.WorkAttendance
	if err := h.DB.First(&attendance, attID).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "考勤记录不存在")
		return
	}
	if attendance.CheckoutTime != nil {
		response.Fail(c, "ALREADY_CHECKED_OUT", "已签退")
		return
	}

	now := time.Now()
	duration := now.Sub(attendance.CheckinTime).Hours()
	hours := math.Round(duration*10) / 10 // 保留1位小数

	h.DB.Model(&attendance).Updates(map[string]any{
		"checkout_time": &now,
		"hours":         hours,
	})
	h.DB.First(&attendance, attID)
	response.OK(c, attendance)
}

// ---------- 薪资管理 ----------

// CalculateSalary 计算月薪
func (h *WorkStudyHandler) CalculateSalary(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var req SalaryCalculateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	var job models.WorkStudyJob
	if err := h.DB.First(&job, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "岗位不存在")
		return
	}

	// 查询该月所有考勤记录
	var attendances []models.WorkAttendance
	if err := h.DB.Where("job_id = ? AND date LIKE ?", id, req.Month+"%").Find(&attendances).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "查询考勤记录失败")
		return
	}

	// 按学生汇总工时
	studentHours := make(map[uint]float64)
	for _, att := range attendances {
		studentHours[att.StudentID] += att.Hours
	}

	var salaryRecords []models.SalaryRecord
	for studentID, hours := range studentHours {
		hours = math.Round(hours*10) / 10 // 保留1位小数
		amount := math.Round(hours*job.SalaryPerHour*100) / 100

		// 查找是否已有该月记录
		var record models.SalaryRecord
		result := h.DB.Where("job_id = ? AND student_id = ? AND month = ?", id, studentID, req.Month).First(&record)

		if result.Error == gorm.ErrRecordNotFound {
			// 新建
			record = models.SalaryRecord{
				JobID:     id,
				StudentID: studentID,
				Month:     req.Month,
				Hours:     hours,
				Amount:    amount,
				Status:    models.SalaryPending,
			}
			if err := h.DB.Create(&record).Error; err != nil {
				response.Fail(c, "INTERNAL_ERROR", "创建薪资记录失败")
				return
			}
		} else {
			// 更新
			h.DB.Model(&record).Updates(map[string]any{
				"hours":  hours,
				"amount": amount,
			})
			h.DB.First(&record, record.ID)
		}
		salaryRecords = append(salaryRecords, record)
	}

	response.OK(c, salaryRecords)
}

// ListSalaries 薪资记录列表
func (h *WorkStudyHandler) ListSalaries(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	query := h.DB.Model(&models.SalaryRecord{}).Where("job_id = ?", id)
	if month := c.Query("month"); month != "" {
		query = query.Where("month = ?", month)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}
	if studentID := c.Query("studentId"); studentID != "" {
		query = query.Where("student_id = ?", studentID)
	}

	var salaries []models.SalaryRecord
	if err := query.Order("id DESC").Find(&salaries).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "查询失败")
		return
	}
	response.OK(c, salaries)
}

// PaySalary 标记薪资已发放
func (h *WorkStudyHandler) PaySalary(c *gin.Context) {
	salaryID, err := parseID(c.Param("salaryId"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var salary models.SalaryRecord
	if err := h.DB.First(&salary, salaryID).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "薪资记录不存在")
		return
	}
	if salary.Status != models.SalaryPending {
		response.Fail(c, "INVALID_STATUS", "仅待发放状态可标记发放")
		return
	}

	now := time.Now()
	h.DB.Model(&salary).Updates(map[string]any{
		"status":  models.SalaryPaid,
		"paid_at": &now,
	})
	h.DB.First(&salary, salaryID)
	response.OK(c, salary)
}

// ---------- 文件上传 ----------

// UploadJobFile 上传岗位附件
func (h *WorkStudyHandler) UploadJobFile(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var job models.WorkStudyJob
	if err := h.DB.First(&job, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "岗位不存在")
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

	objectKey, url, err := h.Uploader.Upload(c.Request.Context(), "workstudy", fileHeader.Filename, fileHeader.Size, f, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		response.Fail(c, "UPLOAD_FAILED", err.Error())
		return
	}

	uploadedBy, _ := c.Get("userId")
	uploaderID, _ := uploadedBy.(float64)

	jobFile := models.WorkStudyFile{
		JobID:      id,
		FileName:   fileHeader.Filename,
		ObjectKey:  objectKey,
		URL:        url,
		FileType:   fileType,
		Size:       fileHeader.Size,
		UploadedBy: uint(uploaderID),
	}
	if err := h.DB.Create(&jobFile).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "保存文件记录失败")
		return
	}
	response.Created(c, jobFile)
}

// ---------- 统计 ----------

// WorkStudyStatistics 勤工助学统计
func (h *WorkStudyHandler) WorkStudyStatistics(c *gin.Context) {
	var totalJobs int64
	h.DB.Model(&models.WorkStudyJob{}).Count(&totalJobs)

	// 按状态统计
	type StatusCount struct {
		Status string `json:"status"`
		Count  int64  `json:"count"`
	}
	var statusCounts []StatusCount
	h.DB.Model(&models.WorkStudyJob{}).Select("status, count(*) as count").Group("status").Scan(&statusCounts)

	// 总报名人数
	var totalApplications int64
	h.DB.Model(&models.JobApplication{}).Where("status = ?", models.AppApplied).Count(&totalApplications)

	// 总录用人数
	var totalAccepted int64
	h.DB.Model(&models.JobApplication{}).Where("status = ?", models.AppAccepted).Count(&totalAccepted)

	// 已发放薪资总额
	var totalSalaryPaid float64
	h.DB.Model(&models.SalaryRecord{}).Where("status = ?", models.SalaryPaid).Select("COALESCE(SUM(amount), 0)").Scan(&totalSalaryPaid)

	// 近期岗位（最近5个）
	var recentJobs []models.WorkStudyJob
	h.DB.Order("created_at DESC").Limit(5).Find(&recentJobs)

	response.OK(c, gin.H{
		"totalJobs":        totalJobs,
		"statusBreakdown":  statusCounts,
		"totalApplications": totalApplications,
		"totalAccepted":    totalAccepted,
		"totalSalaryPaid":  totalSalaryPaid,
		"recentJobs":       recentJobs,
	})
}
