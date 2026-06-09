package router

import (
	"bytes"
	"encoding/json"
	"fmt"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/config"
	"unistep-platform/backend/internal/models"
)

func setupWorkStudyRouter(t *testing.T) (*gin.Engine, *memoryUploader) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.WorkStudyJob{},
		&models.JobApplication{},
		&models.WorkAttendance{},
		&models.SalaryRecord{},
		&models.WorkStudyFile{},
	); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	uploader := newMemoryUploader()
	r := New(config.Config{JWTSecret: "test-secret", FrontendURL: "http://localhost:5173"}, db, uploader)
	return r, uploader
}

// registerSecondUser 注册第二个用户并返回 token
func registerSecondUser(t *testing.T, r *gin.Engine) string {
	t.Helper()
	doJSONRequest(r, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "tester2",
		"password": "password123",
	}, "")
	w := doJSONRequest(r, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"username": "tester2",
		"password": "password123",
	}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("second user login failed: %s", w.Body.String())
	}
	var body struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return body.Data.Token
}

// ---------- 测试用例 ----------

func TestWorkStudyJobCRUD(t *testing.T) {
	r, _ := setupWorkStudyRouter(t)
	token := registerAndLogin(t, r)

	// 创建岗位
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/workstudy/jobs", map[string]any{
		"title":         "图书馆助理",
		"department":    "图书馆",
		"location":      "校图书馆一楼",
		"description":   "协助图书整理与借阅服务",
		"quota":         5,
		"salaryPerHour": 18.5,
		"startTime":     "2025-09-01T08:00:00Z",
		"endTime":       "2026-01-15T18:00:00Z",
		"contactPerson": "王老师",
		"contactPhone":  "13800001111",
	}, token)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create job expected 201, got %d: %s", createResp.Code, createResp.Body.String())
	}
	var job models.WorkStudyJob
	decodeData(t, createResp, &job)
	if job.ID == 0 {
		t.Fatal("expected job id, got 0")
	}
	if job.Status != models.JobDraft {
		t.Fatalf("expected draft status, got %s", job.Status)
	}

	// 列表查询
	listResp := doJSONRequest(r, http.MethodGet, "/api/v1/workstudy/jobs?page=1&size=10", nil, token)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d", listResp.Code)
	}

	// 详情
	getResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/workstudy/jobs/%d", job.ID), nil, token)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get expected 200, got %d", getResp.Code)
	}

	// 更新岗位
	updateResp := doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/workstudy/jobs/%d", job.ID), map[string]any{
		"title":         "图书馆助理(更新)",
		"department":    "图书馆",
		"location":      "校图书馆一楼",
		"description":   "协助图书整理与借阅服务",
		"quota":         8,
		"salaryPerHour": 20.0,
		"startTime":     "2025-09-01T08:00:00Z",
		"endTime":       "2026-01-15T18:00:00Z",
		"contactPerson": "王老师",
		"contactPhone":  "13800001111",
	}, token)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update expected 200, got %d: %s", updateResp.Code, updateResp.Body.String())
	}
	var updated models.WorkStudyJob
	decodeData(t, updateResp, &updated)
	if updated.Title != "图书馆助理(更新)" {
		t.Fatalf("expected updated title, got %s", updated.Title)
	}
	if updated.Quota != 8 {
		t.Fatalf("expected updated quota 8, got %d", updated.Quota)
	}

	// 删除草稿岗位
	delResp := doJSONRequest(r, http.MethodDelete, fmt.Sprintf("/api/v1/workstudy/jobs/%d", job.ID), nil, token)
	if delResp.Code != http.StatusOK {
		t.Fatalf("delete expected 200, got %d: %s", delResp.Code, delResp.Body.String())
	}

	// 验证已删除
	getDelResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/workstudy/jobs/%d", job.ID), nil, token)
	if getDelResp.Code != http.StatusBadRequest {
		t.Fatalf("get deleted expected 400, got %d", getDelResp.Code)
	}
}

func TestWorkStudyJobPublishAndClose(t *testing.T) {
	r, _ := setupWorkStudyRouter(t)
	token := registerAndLogin(t, r)

	// 创建岗位
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/workstudy/jobs", map[string]any{
		"title":         "实验室管理员",
		"department":    "计算机学院",
		"location":      "实验楼B201",
		"description":   "负责实验室设备维护",
		"quota":         3,
		"salaryPerHour": 15.0,
		"startTime":     "2025-09-01T08:00:00Z",
		"endTime":       "2026-06-30T18:00:00Z",
		"contactPerson": "李老师",
		"contactPhone":  "13900002222",
	}, token)
	var job models.WorkStudyJob
	decodeData(t, createResp, &job)

	// 发布岗位
	publishResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/publish", job.ID), nil, token)
	if publishResp.Code != http.StatusOK {
		t.Fatalf("publish expected 200, got %d: %s", publishResp.Code, publishResp.Body.String())
	}
	var published models.WorkStudyJob
	decodeData(t, publishResp, &published)
	if published.Status != models.JobPublished {
		t.Fatalf("expected published status, got %s", published.Status)
	}

	// 尝试编辑已发布岗位（应失败）
	editResp := doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/workstudy/jobs/%d", job.ID), map[string]any{
		"title":         "实验室管理员(修改)",
		"department":    "计算机学院",
		"location":      "实验楼B201",
		"description":   "负责实验室设备维护",
		"quota":         5,
		"salaryPerHour": 15.0,
		"startTime":     "2025-09-01T08:00:00Z",
		"endTime":       "2026-06-30T18:00:00Z",
		"contactPerson": "李老师",
		"contactPhone":  "13900002222",
	}, token)
	if editResp.Code != http.StatusBadRequest {
		t.Fatalf("edit published job expected 400, got %d", editResp.Code)
	}

	// 关闭岗位
	closeResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/close", job.ID), nil, token)
	if closeResp.Code != http.StatusOK {
		t.Fatalf("close expected 200, got %d: %s", closeResp.Code, closeResp.Body.String())
	}
	var closed models.WorkStudyJob
	decodeData(t, closeResp, &closed)
	if closed.Status != models.JobClosed {
		t.Fatalf("expected closed status, got %s", closed.Status)
	}
}

func TestWorkStudyApplicationFlow(t *testing.T) {
	r, _ := setupWorkStudyRouter(t)
	token := registerAndLogin(t, r)

	// 创建并发布岗位
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/workstudy/jobs", map[string]any{
		"title":         "行政助理",
		"department":    "学生处",
		"location":      "行政楼302",
		"description":   "协助处理学生事务",
		"quota":         2,
		"salaryPerHour": 16.0,
		"startTime":     "2025-09-01T08:00:00Z",
		"endTime":       "2026-06-30T18:00:00Z",
		"contactPerson": "赵老师",
		"contactPhone":  "13700003333",
	}, token)
	var job models.WorkStudyJob
	decodeData(t, createResp, &job)

	doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/publish", job.ID), nil, token)

	// 报名
	applyResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/apply", job.ID), nil, token)
	if applyResp.Code != http.StatusCreated {
		t.Fatalf("apply expected 201, got %d: %s", applyResp.Code, applyResp.Body.String())
	}
	var application models.JobApplication
	decodeData(t, applyResp, &application)
	if application.Status != models.AppApplied {
		t.Fatalf("expected applied status, got %s", application.Status)
	}

	// 重复报名应失败
	dupApplyResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/apply", job.ID), nil, token)
	if dupApplyResp.Code != http.StatusBadRequest {
		t.Fatalf("duplicate apply expected 400, got %d", dupApplyResp.Code)
	}

	// 查看报名列表
	listResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/workstudy/jobs/%d/applications", job.ID), nil, token)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list applications expected 200, got %d", listResp.Code)
	}

	// 录用报名
	acceptResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/applications/%d/accept", job.ID, application.ID), map[string]any{
		"remark": "符合岗位要求",
	}, token)
	if acceptResp.Code != http.StatusOK {
		t.Fatalf("accept expected 200, got %d: %s", acceptResp.Code, acceptResp.Body.String())
	}
	var accepted models.JobApplication
	decodeData(t, acceptResp, &accepted)
	if accepted.Status != models.AppAccepted {
		t.Fatalf("expected accepted status, got %s", accepted.Status)
	}

	// 创建第二个用户，报名，然后拒绝
	token2 := registerSecondUser(t, r)
	applyResp2 := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/apply", job.ID), nil, token2)
	if applyResp2.Code != http.StatusCreated {
		t.Fatalf("second apply expected 201, got %d: %s", applyResp2.Code, applyResp2.Body.String())
	}
	var application2 models.JobApplication
	decodeData(t, applyResp2, &application2)

	rejectResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/applications/%d/reject", job.ID, application2.ID), map[string]any{
		"remark": "名额已满",
	}, token)
	if rejectResp.Code != http.StatusOK {
		t.Fatalf("reject expected 200, got %d: %s", rejectResp.Code, rejectResp.Body.String())
	}
	var rejected models.JobApplication
	decodeData(t, rejectResp, &rejected)
	if rejected.Status != models.AppRejected {
		t.Fatalf("expected rejected status, got %s", rejected.Status)
	}

	// 创建第三个用户，报名后取消
	token3 := registerThirdUser(t, r)
	applyResp3 := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/apply", job.ID), nil, token3)
	if applyResp3.Code != http.StatusCreated {
		t.Fatalf("third apply expected 201, got %d: %s", applyResp3.Code, applyResp3.Body.String())
	}

	cancelResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/cancel-application", job.ID), nil, token3)
	if cancelResp.Code != http.StatusOK {
		t.Fatalf("cancel expected 200, got %d: %s", cancelResp.Code, cancelResp.Body.String())
	}
}

func registerThirdUser(t *testing.T, r *gin.Engine) string {
	t.Helper()
	doJSONRequest(r, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "tester3",
		"password": "password123",
	}, "")
	w := doJSONRequest(r, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"username": "tester3",
		"password": "password123",
	}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("third user login failed: %s", w.Body.String())
	}
	var body struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return body.Data.Token
}

func TestWorkStudyAttendance(t *testing.T) {
	r, _ := setupWorkStudyRouter(t)
	token := registerAndLogin(t, r)

	// 创建并发布岗位
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/workstudy/jobs", map[string]any{
		"title":         "机房值班员",
		"department":    "信息中心",
		"location":      "机房A区",
		"description":   "负责机房日常值班",
		"quota":         3,
		"salaryPerHour": 17.0,
		"startTime":     "2025-09-01T08:00:00Z",
		"endTime":       "2026-06-30T18:00:00Z",
		"contactPerson": "陈老师",
		"contactPhone":  "13600004444",
	}, token)
	var job models.WorkStudyJob
	decodeData(t, createResp, &job)

	doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/publish", job.ID), nil, token)

	// 报名并录用
	applyResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/apply", job.ID), nil, token)
	var application models.JobApplication
	decodeData(t, applyResp, &application)
	doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/applications/%d/accept", job.ID, application.ID), nil, token)

	// 创建考勤记录
	attResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/attendances", job.ID), map[string]any{
		"studentId": application.StudentID,
		"date":      "2025-10-15",
		"method":    "manual",
	}, token)
	if attResp.Code != http.StatusCreated {
		t.Fatalf("create attendance expected 201, got %d: %s", attResp.Code, attResp.Body.String())
	}
	var attendance models.WorkAttendance
	decodeData(t, attResp, &attendance)
	if attendance.ID == 0 {
		t.Fatal("expected attendance id, got 0")
	}

	// 查看考勤列表
	listResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/workstudy/jobs/%d/attendances", job.ID), nil, token)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list attendances expected 200, got %d", listResp.Code)
	}

	// 签退
	checkoutResp := doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/workstudy/attendances/%d/checkout", attendance.ID), nil, token)
	if checkoutResp.Code != http.StatusOK {
		t.Fatalf("checkout expected 200, got %d: %s", checkoutResp.Code, checkoutResp.Body.String())
	}
	var checked models.WorkAttendance
	decodeData(t, checkoutResp, &checked)
	if checked.Hours < 0 {
		t.Fatalf("expected non-negative hours after checkout, got %f", checked.Hours)
	}
	if checked.CheckoutTime == nil {
		t.Fatal("expected checkout time to be set")
	}

	// 同一天重复考勤应失败
	dupAttResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/attendances", job.ID), map[string]any{
		"studentId": application.StudentID,
		"date":      "2025-10-15",
		"method":    "manual",
	}, token)
	if dupAttResp.Code != http.StatusBadRequest {
		t.Fatalf("duplicate attendance expected 400, got %d", dupAttResp.Code)
	}
}

func TestWorkStudySalary(t *testing.T) {
	r, _ := setupWorkStudyRouter(t)
	token := registerAndLogin(t, r)

	// 创建并发布岗位（时薪20元）
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/workstudy/jobs", map[string]any{
		"title":         "教务助理",
		"department":    "教务处",
		"location":      "教务楼105",
		"description":   "协助教务日常事务",
		"quota":         2,
		"salaryPerHour": 20.0,
		"startTime":     "2025-09-01T08:00:00Z",
		"endTime":       "2026-06-30T18:00:00Z",
		"contactPerson": "刘老师",
		"contactPhone":  "13500005555",
	}, token)
	var job models.WorkStudyJob
	decodeData(t, createResp, &job)

	doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/publish", job.ID), nil, token)

	// 报名并录用
	applyResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/apply", job.ID), nil, token)
	var application models.JobApplication
	decodeData(t, applyResp, &application)
	doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/applications/%d/accept", job.ID, application.ID), nil, token)

	// 创建考勤并签退
	attResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/attendances", job.ID), map[string]any{
		"studentId": application.StudentID,
		"date":      "2025-10-10",
		"method":    "manual",
	}, token)
	var attendance models.WorkAttendance
	decodeData(t, attResp, &attendance)
	doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/workstudy/attendances/%d/checkout", attendance.ID), nil, token)

	// 计算薪资
	calcResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/salary/calculate", job.ID), map[string]any{
		"month": "2025-10",
	}, token)
	if calcResp.Code != http.StatusOK {
		t.Fatalf("calculate salary expected 200, got %d: %s", calcResp.Code, calcResp.Body.String())
	}
	var salaryRecords []models.SalaryRecord
	decodeData(t, calcResp, &salaryRecords)
	if len(salaryRecords) == 0 {
		t.Fatal("expected at least 1 salary record")
	}
	salary := salaryRecords[0]
	if salary.Status != models.SalaryPending {
		t.Fatalf("expected pending status, got %s", salary.Status)
	}
	if salary.Amount < 0 {
		t.Fatalf("expected non-negative amount, got %f", salary.Amount)
	}

	// 查看薪资列表
	listResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/workstudy/jobs/%d/salary", job.ID), nil, token)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list salaries expected 200, got %d", listResp.Code)
	}

	// 发放薪资
	payResp := doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/workstudy/salary/%d/pay", salary.ID), nil, token)
	if payResp.Code != http.StatusOK {
		t.Fatalf("pay salary expected 200, got %d: %s", payResp.Code, payResp.Body.String())
	}
	var paid models.SalaryRecord
	decodeData(t, payResp, &paid)
	if paid.Status != models.SalaryPaid {
		t.Fatalf("expected paid status, got %s", paid.Status)
	}
}

func TestWorkStudyStatistics(t *testing.T) {
	r, _ := setupWorkStudyRouter(t)
	token := registerAndLogin(t, r)

	// 创建几个岗位
	for i := 0; i < 3; i++ {
		doJSONRequest(r, http.MethodPost, "/api/v1/workstudy/jobs", map[string]any{
			"title":         fmt.Sprintf("岗位%d", i),
			"department":    fmt.Sprintf("部门%d", i),
			"location":      "教学楼",
			"description":   "测试岗位描述",
			"quota":         5,
			"salaryPerHour": 15.0,
			"startTime":     "2025-09-01T08:00:00Z",
			"endTime":       "2026-06-30T18:00:00Z",
			"contactPerson": "联系人",
			"contactPhone":  "13800001111",
		}, token)
	}

	// 获取统计
	statResp := doJSONRequest(r, http.MethodGet, "/api/v1/workstudy/statistics", nil, token)
	if statResp.Code != http.StatusOK {
		t.Fatalf("statistics expected 200, got %d: %s", statResp.Code, statResp.Body.String())
	}

	var statData struct {
		TotalJobs        int64 `json:"totalJobs"`
		StatusBreakdown  []struct {
			Status string `json:"status"`
			Count  int64  `json:"count"`
		} `json:"statusBreakdown"`
		TotalApplications int64   `json:"totalApplications"`
		TotalAccepted     int64   `json:"totalAccepted"`
		TotalSalaryPaid   float64 `json:"totalSalaryPaid"`
	}
	decodeData(t, statResp, &statData)
	if statData.TotalJobs != 3 {
		t.Fatalf("expected totalJobs=3, got %d", statData.TotalJobs)
	}
	if len(statData.StatusBreakdown) == 0 {
		t.Fatal("expected non-empty statusBreakdown")
	}
}

func TestWorkStudyFileUpload(t *testing.T) {
	r, uploader := setupWorkStudyRouter(t)
	token := registerAndLogin(t, r)

	// 创建岗位
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/workstudy/jobs", map[string]any{
		"title":         "后勤助理",
		"department":    "后勤处",
		"location":      "后勤楼",
		"description":   "协助后勤管理",
		"quota":         2,
		"salaryPerHour": 14.0,
		"startTime":     "2025-09-01T08:00:00Z",
		"endTime":       "2026-06-30T18:00:00Z",
		"contactPerson": "孙老师",
		"contactPhone":  "13400006666",
	}, token)
	var job models.WorkStudyJob
	decodeData(t, createResp, &job)

	// 上传文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("fileType", "document")
	part, _ := writer.CreateFormFile("file", "岗位说明.pdf")
	_, _ = part.Write([]byte("fake pdf data"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/workstudy/jobs/%d/files", job.ID), body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("upload expected 201, got %d: %s", w.Code, w.Body.String())
	}
	if len(uploader.Objects) != 1 {
		t.Fatalf("expected 1 uploaded object, got %d", len(uploader.Objects))
	}
}
