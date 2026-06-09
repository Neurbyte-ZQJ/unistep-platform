package handler

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/models"
	"unistep-platform/backend/internal/response"
)

// Uploader 定义文件上传抽象，便于在测试中使用本地实现替换 MinIO
type Uploader interface {
	Upload(ctx context.Context, category, fileName string, size int64, reader io.Reader, contentType string) (objectKey, url string, err error)
	Disabled() bool
}

// MemberHandler 团员发展模块处理器
type MemberHandler struct {
	DB       *gorm.DB
	Uploader Uploader
}

// NewMemberHandler 创建团员发展处理器
func NewMemberHandler(db *gorm.DB, uploader Uploader) *MemberHandler {
	return &MemberHandler{DB: db, Uploader: uploader}
}

// ---------- 请求/响应结构 ----------

// MemberProfileRequest 团员档案录入/更新参数
type MemberProfileRequest struct {
	UserID    uint   `json:"userId"`
	Name      string `json:"name" binding:"required,max=64"`
	StudentNo string `json:"studentNo" binding:"required,max=32"`
	Gender    string `json:"gender" binding:"omitempty,oneof=男 女"`
	Birthday  string `json:"birthday"`
	IDCard    string `json:"idCard"`
	Nation    string `json:"nation"`
	Phone     string `json:"phone"`
	College   string `json:"college"`
	Major     string `json:"major"`
	ClassName string `json:"className"`
	Stage     string `json:"stage" binding:"omitempty,oneof=applicant activist develop_target political_review league_member"`
	JoinDate  string `json:"joinDate"`
	Remark    string `json:"remark"`
}

// ---------- 团员档案 CRUD ----------

// ListProfiles 团员列表（支持分页与阶段过滤）
// @Summary  团员列表
// @Tags     团员发展
// @Param    page   query  int     false  "页码，从 1 开始"
// @Param    size   query  int     false  "每页数量"
// @Param    stage  query  string  false  "阶段过滤"
// @Param    name   query  string  false  "姓名模糊搜索"
// @Success  200    {object}  response.Body
// @Router   /api/v1/members [get]
func (h *MemberHandler) ListProfiles(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "10"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 10
	}

	query := h.DB.Model(&models.MemberProfile{})
	if stage := c.Query("stage"); stage != "" {
		query = query.Where("stage = ?", stage)
	}
	if name := c.Query("name"); name != "" {
		query = query.Where("name LIKE ?", "%"+name+"%")
	}

	var total int64
	query.Count(&total)

	var items []models.MemberProfile
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

// CreateProfile 创建团员档案
// @Summary  创建团员档案
// @Tags     团员发展
// @Accept   json
// @Param    body  body   MemberProfileRequest  true  "档案信息"
// @Success  201   {object}  response.Body
// @Router   /api/v1/members [post]
func (h *MemberHandler) CreateProfile(c *gin.Context) {
	var req MemberProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	// 学号唯一性校验
	var count int64
	h.DB.Model(&models.MemberProfile{}).Where("student_no = ?", req.StudentNo).Count(&count)
	if count > 0 {
		response.Fail(c, "PROFILE_EXISTS", "该学号档案已存在")
		return
	}

	profile := models.MemberProfile{
		UserID:    req.UserID,
		Name:      req.Name,
		StudentNo: req.StudentNo,
		Gender:    req.Gender,
		Birthday:  req.Birthday,
		IDCard:    req.IDCard,
		Nation:    req.Nation,
		Phone:     req.Phone,
		College:   req.College,
		Major:     req.Major,
		ClassName: req.ClassName,
		Stage:     defaultString(req.Stage, models.StageApplicant),
		JoinDate:  req.JoinDate,
		Remark:    req.Remark,
	}
	if err := h.DB.Create(&profile).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "创建档案失败")
		return
	}
	response.Created(c, profile)
}

// GetProfile 获取档案详情（含子记录）
// @Summary  获取团员档案
// @Tags     团员发展
// @Param    id   path  int  true  "档案 ID"
// @Success  200  {object}  response.Body
// @Router   /api/v1/members/{id} [get]
func (h *MemberHandler) GetProfile(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var profile models.MemberProfile
	if err := h.DB.
		Preload("Applications").
		Preload("ActivistRecords").
		Preload("DevelopRecords").
		Preload("PoliticalRecords").
		Preload("Attachments").
		First(&profile, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "档案不存在")
		return
	}
	response.OK(c, profile)
}

// UpdateProfile 更新团员档案
// @Summary  更新团员档案
// @Tags     团员发展
// @Accept   json
// @Param    id    path  int                   true  "档案 ID"
// @Param    body  body  MemberProfileRequest  true  "档案信息"
// @Success  200   {object}  response.Body
// @Router   /api/v1/members/{id} [put]
func (h *MemberHandler) UpdateProfile(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var req MemberProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}

	var profile models.MemberProfile
	if err := h.DB.First(&profile, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "档案不存在")
		return
	}

	updates := map[string]any{
		"user_id":    req.UserID,
		"name":       req.Name,
		"student_no": req.StudentNo,
		"gender":     req.Gender,
		"birthday":   req.Birthday,
		"id_card":    req.IDCard,
		"nation":     req.Nation,
		"phone":      req.Phone,
		"college":    req.College,
		"major":      req.Major,
		"class_name": req.ClassName,
		"join_date":  req.JoinDate,
		"remark":     req.Remark,
	}
	if req.Stage != "" {
		updates["stage"] = req.Stage
	}
	if err := h.DB.Model(&profile).Updates(updates).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "更新失败")
		return
	}
	response.OK(c, profile)
}

// DeleteProfile 删除档案
// @Summary  删除团员档案
// @Tags     团员发展
// @Param    id   path  int  true  "档案 ID"
// @Success  200  {object}  response.Body
// @Router   /api/v1/members/{id} [delete]
func (h *MemberHandler) DeleteProfile(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}
	if err := h.DB.Delete(&models.MemberProfile{}, id).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "删除失败")
		return
	}
	response.OK(c, gin.H{"id": id})
}

// ---------- 入团申请 ----------

// CreateApplication 提交入团申请
// @Summary  提交入团申请
// @Tags     团员发展
// @Accept   json
// @Param    id    path  int                       true  "档案 ID"
// @Param    body  body  models.LeagueApplication  true  "申请信息"
// @Success  201   {object}  response.Body
// @Router   /api/v1/members/{id}/applications [post]
func (h *MemberHandler) CreateApplication(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}
	var record models.LeagueApplication
	if err := c.ShouldBindJSON(&record); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}
	if record.ApplyDate == "" {
		response.Fail(c, "INVALID_PARAMS", "申请日期必填")
		return
	}
	record.ID = 0
	record.ProfileID = id
	if record.Status == "" {
		record.Status = "pending"
	}
	if err := h.DB.Create(&record).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "创建失败")
		return
	}
	// 同步推进档案阶段
	h.DB.Model(&models.MemberProfile{}).Where("id = ? AND stage = ?", id, "").Update("stage", models.StageApplicant)
	response.Created(c, record)
}

// CreateActivistRecord 录入积极分子培养记录
// @Summary  录入积极分子培养记录
// @Tags     团员发展
// @Accept   json
// @Param    id    path  int                       true  "档案 ID"
// @Param    body  body  models.ActivistRecord     true  "培养记录"
// @Success  201   {object}  response.Body
// @Router   /api/v1/members/{id}/activists [post]
func (h *MemberHandler) CreateActivistRecord(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}
	var record models.ActivistRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}
	if record.StartDate == "" {
		response.Fail(c, "INVALID_PARAMS", "培养开始日期必填")
		return
	}
	record.ID = 0
	record.ProfileID = id
	if err := h.DB.Create(&record).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "创建失败")
		return
	}
	h.DB.Model(&models.MemberProfile{}).Where("id = ?", id).Update("stage", models.StageActivist)
	response.Created(c, record)
}

// CreateDevelopRecord 录入发展对象记录
// @Summary  录入发展对象记录
// @Tags     团员发展
// @Accept   json
// @Param    id    path  int                          true  "档案 ID"
// @Param    body  body  models.DevelopTargetRecord   true  "发展对象信息"
// @Success  201   {object}  response.Body
// @Router   /api/v1/members/{id}/develop-targets [post]
func (h *MemberHandler) CreateDevelopRecord(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}
	var record models.DevelopTargetRecord
	if err := c.ShouldBindJSON(&record); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}
	if record.ConfirmedDate == "" {
		response.Fail(c, "INVALID_PARAMS", "确定日期必填")
		return
	}
	record.ID = 0
	record.ProfileID = id
	if err := h.DB.Create(&record).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "创建失败")
		return
	}
	h.DB.Model(&models.MemberProfile{}).Where("id = ?", id).Update("stage", models.StageDevelopTarget)
	response.Created(c, record)
}

// CreatePoliticalReview 政审备案
// @Summary  政审备案
// @Tags     团员发展
// @Accept   json
// @Param    id    path  int                       true  "档案 ID"
// @Param    body  body  models.PoliticalReview    true  "政审信息"
// @Success  201   {object}  response.Body
// @Router   /api/v1/members/{id}/political-reviews [post]
func (h *MemberHandler) CreatePoliticalReview(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}
	var record models.PoliticalReview
	if err := c.ShouldBindJSON(&record); err != nil {
		response.Fail(c, "INVALID_PARAMS", "参数错误: "+err.Error())
		return
	}
	if record.ReviewDate == "" {
		response.Fail(c, "INVALID_PARAMS", "政审日期必填")
		return
	}
	record.ID = 0
	record.ProfileID = id
	if record.Status == "" {
		record.Status = "filed"
	}
	if err := h.DB.Create(&record).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "创建失败")
		return
	}
	h.DB.Model(&models.MemberProfile{}).Where("id = ?", id).Update("stage", models.StagePoliticalReview)
	response.Created(c, record)
}

// GenerateArchive 生成团员电子档案（聚合视图）
// @Summary  生成团员电子档案
// @Tags     团员发展
// @Param    id   path  int  true  "档案 ID"
// @Success  200  {object}  response.Body
// @Router   /api/v1/members/{id}/archive [get]
func (h *MemberHandler) GenerateArchive(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}

	var profile models.MemberProfile
	if err := h.DB.
		Preload("Applications").
		Preload("ActivistRecords").
		Preload("DevelopRecords").
		Preload("PoliticalRecords").
		Preload("Attachments").
		First(&profile, id).Error; err != nil {
		response.Fail(c, "NOT_FOUND", "档案不存在")
		return
	}

	timeline := buildTimeline(&profile)
	response.OK(c, gin.H{
		"profile":  profile,
		"timeline": timeline,
		"summary": gin.H{
			"stage":               profile.Stage,
			"applicationCount":    len(profile.Applications),
			"activistCount":       len(profile.ActivistRecords),
			"developRecordCount":  len(profile.DevelopRecords),
			"politicalRecordCount": len(profile.PoliticalRecords),
			"attachmentCount":     len(profile.Attachments),
		},
	})
}

// UploadAttachment 上传档案附件至 MinIO
// @Summary  上传档案附件
// @Tags     团员发展
// @Accept   multipart/form-data
// @Param    id        path  int    true  "档案 ID"
// @Param    category  formData  string  true  "附件类别"
// @Param    file      formData  file    true  "文件内容"
// @Success  201       {object}  response.Body
// @Router   /api/v1/members/{id}/attachments [post]
func (h *MemberHandler) UploadAttachment(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		response.Fail(c, "INVALID_PARAMS", err.Error())
		return
	}
	if h.Uploader == nil || h.Uploader.Disabled() {
		response.Fail(c, "STORAGE_DISABLED", "对象存储未启用，请配置 MINIO_ENDPOINT")
		return
	}
	category := c.DefaultPostForm("category", "other")

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

	objectKey, url, err := h.Uploader.Upload(c.Request.Context(), category, fileHeader.Filename, fileHeader.Size, f, fileHeader.Header.Get("Content-Type"))
	if err != nil {
		response.Fail(c, "UPLOAD_FAILED", err.Error())
		return
	}

	uploadedBy, _ := c.Get("userId")
	uploaderID, _ := uploadedBy.(float64) // JWT claims 数字默认是 float64
	attachment := models.MemberAttachment{
		ProfileID:  id,
		Category:   category,
		FileName:   fileHeader.Filename,
		ObjectKey:  objectKey,
		URL:        url,
		Size:       fileHeader.Size,
		UploadedBy: uint(uploaderID),
	}
	if err := h.DB.Create(&attachment).Error; err != nil {
		response.Fail(c, "INTERNAL_ERROR", "保存附件记录失败")
		return
	}
	response.Created(c, attachment)
}

// ---------- helpers ----------

func parseID(raw string) (uint, error) {
	id, err := strconv.ParseUint(raw, 10, 64)
	if err != nil || id == 0 {
		return 0, errors.New("invalid id")
	}
	return uint(id), nil
}

func defaultString(value, fallback string) string {
	if value == "" {
		return fallback
	}
	return value
}

// buildTimeline 将各阶段记录按时间汇总为时间轴
func buildTimeline(p *models.MemberProfile) []map[string]string {
	timeline := make([]map[string]string, 0)
	for _, a := range p.Applications {
		timeline = append(timeline, map[string]string{
			"date":  a.ApplyDate,
			"stage": "入团申请",
			"text":  fmt.Sprintf("提交入团申请，状态=%s", a.Status),
		})
	}
	for _, a := range p.ActivistRecords {
		timeline = append(timeline, map[string]string{
			"date":  a.StartDate,
			"stage": "积极分子培养",
			"text":  fmt.Sprintf("培养联系人=%s", a.Trainer),
		})
	}
	for _, d := range p.DevelopRecords {
		timeline = append(timeline, map[string]string{
			"date":  d.ConfirmedDate,
			"stage": "发展对象",
			"text":  fmt.Sprintf("结论=%s", d.Conclusion),
		})
	}
	for _, r := range p.PoliticalRecords {
		timeline = append(timeline, map[string]string{
			"date":  r.ReviewDate,
			"stage": "政审备案",
			"text":  fmt.Sprintf("审核人=%s 结论=%s", r.Reviewer, r.Conclusion),
		})
	}
	return timeline
}

// 防止 gorm.ErrRecordNotFound 在导入但未使用时报错（GenerateArchive 中可能用到）
var _ = gorm.ErrRecordNotFound
