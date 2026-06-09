package router

import (
	"bytes"
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

func setupActivityRouter(t *testing.T) (*gin.Engine, *memoryUploader) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.ClubActivity{},
		&models.ActivityRegistration{},
		&models.ActivityCheckin{},
		&models.ActivityFile{},
	); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	uploader := newMemoryUploader()
	r := New(config.Config{JWTSecret: "test-secret", FrontendURL: "http://localhost:5173"}, db, uploader)
	return r, uploader
}

// ---------- 测试用例 ----------

func TestActivityCRUD(t *testing.T) {
	r, _ := setupActivityRouter(t)
	token := registerAndLogin(t, r)

	// 创建活动
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/activities", map[string]any{
		"clubName":    "计算机协会",
		"title":       "编程马拉松",
		"startTime":   "2025-09-01T09:00:00Z",
		"endTime":     "2025-09-01T18:00:00Z",
		"location":    "教学楼A101",
		"capacity":    50,
		"description": "24小时编程挑战赛",
		"budget":      2000.00,
	}, token)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create activity expected 201, got %d: %s", createResp.Code, createResp.Body.String())
	}
	var activity models.ClubActivity
	decodeData(t, createResp, &activity)
	if activity.ID == 0 {
		t.Fatal("expected activity id, got 0")
	}
	if activity.Status != models.ActivityDraft {
		t.Fatalf("expected draft status, got %s", activity.Status)
	}

	// 列表查询
	listResp := doJSONRequest(r, http.MethodGet, "/api/v1/activities?page=1&size=10", nil, token)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d", listResp.Code)
	}

	// 详情
	getResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/activities/%d", activity.ID), nil, token)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get expected 200, got %d", getResp.Code)
	}

	// 更新
	updateResp := doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/activities/%d", activity.ID), map[string]any{
		"clubName":    "计算机协会",
		"title":       "编程马拉松(更新)",
		"startTime":   "2025-09-01T09:00:00Z",
		"endTime":     "2025-09-01T18:00:00Z",
		"location":    "教学楼A101",
		"capacity":    60,
		"description": "24小时编程挑战赛-升级版",
	}, token)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update expected 200, got %d: %s", updateResp.Code, updateResp.Body.String())
	}
}

func TestActivityApprovalFlow(t *testing.T) {
	r, _ := setupActivityRouter(t)
	token := registerAndLogin(t, r)

	// 创建活动
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/activities", map[string]any{
		"clubName":    "文学社",
		"title":       "读书分享会",
		"startTime":   "2025-10-01T14:00:00Z",
		"endTime":     "2025-10-01T17:00:00Z",
		"location":    "图书馆报告厅",
		"capacity":    30,
		"description": "好书推荐与分享",
	}, token)
	var activity models.ClubActivity
	decodeData(t, createResp, &activity)

	// 提交审批
	submitResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/submit", activity.ID), nil, token)
	if submitResp.Code != http.StatusOK {
		t.Fatalf("submit expected 200, got %d: %s", submitResp.Code, submitResp.Body.String())
	}

	// 审批通过
	approveResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/approve", activity.ID), map[string]any{
		"opinion": "活动方案合理，同意开展",
		"approve": true,
	}, token)
	if approveResp.Code != http.StatusOK {
		t.Fatalf("approve expected 200, got %d: %s", approveResp.Code, approveResp.Body.String())
	}
	var approved models.ClubActivity
	decodeData(t, approveResp, &approved)
	if approved.Status != models.ActivityRegOpen {
		t.Fatalf("expected reg_open status, got %s", approved.Status)
	}

	// 创建第二个活动测试驳回
	createResp2 := doJSONRequest(r, http.MethodPost, "/api/v1/activities", map[string]any{
		"clubName":    "文学社",
		"title":       "读书分享会2",
		"startTime":   "2025-11-01T14:00:00Z",
		"endTime":     "2025-11-01T17:00:00Z",
		"location":    "图书馆报告厅",
		"capacity":    30,
		"description": "好书推荐与分享2",
	}, token)
	var activity2 models.ClubActivity
	decodeData(t, createResp2, &activity2)

	doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/submit", activity2.ID), nil, token)
	rejectResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/approve", activity2.ID), map[string]any{
		"opinion": "预算不合理",
		"approve": false,
	}, token)
	if rejectResp.Code != http.StatusOK {
		t.Fatalf("reject expected 200, got %d", rejectResp.Code)
	}
	var rejected models.ClubActivity
	decodeData(t, rejectResp, &rejected)
	if rejected.Status != models.ActivityRejected {
		t.Fatalf("expected rejected status, got %s", rejected.Status)
	}
}

func TestActivityRegistrationAndCheckin(t *testing.T) {
	r, _ := setupActivityRouter(t)
	token := registerAndLogin(t, r)

	// 创建并审批活动
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/activities", map[string]any{
		"clubName":    "摄影社",
		"title":       "校园摄影大赛",
		"startTime":   "2025-09-15T08:00:00Z",
		"endTime":     "2025-09-15T20:00:00Z",
		"location":    "校园各处",
		"capacity":    100,
		"description": "记录校园之美",
	}, token)
	var activity models.ClubActivity
	decodeData(t, createResp, &activity)

	// 提交审批并通过
	doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/submit", activity.ID), nil, token)
	doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/approve", activity.ID), map[string]any{
		"opinion": "同意",
		"approve": true,
	}, token)

	// 报名
	regResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/register", activity.ID), nil, token)
	if regResp.Code != http.StatusCreated {
		t.Fatalf("register expected 201, got %d: %s", regResp.Code, regResp.Body.String())
	}

	// 重复报名应失败
	dupRegResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/register", activity.ID), nil, token)
	if dupRegResp.Code != http.StatusBadRequest {
		t.Fatalf("duplicate register expected 400, got %d", dupRegResp.Code)
	}

	// 推进到进行中状态
	doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/activities/%d/status", activity.ID), map[string]any{
		"status": "in_progress",
	}, token)

	// 签到
	checkinResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/checkin", activity.ID), map[string]any{
		"studentId": 1,
	}, token)
	if checkinResp.Code != http.StatusCreated {
		t.Fatalf("checkin expected 201, got %d: %s", checkinResp.Code, checkinResp.Body.String())
	}

	// 重复签到应失败
	dupCheckinResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/checkin", activity.ID), map[string]any{
		"studentId": 1,
	}, token)
	if dupCheckinResp.Code != http.StatusBadRequest {
		t.Fatalf("duplicate checkin expected 400, got %d", dupCheckinResp.Code)
	}
}

func TestActivityFileUpload(t *testing.T) {
	r, uploader := setupActivityRouter(t)
	token := registerAndLogin(t, r)

	// 创建活动
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/activities", map[string]any{
		"clubName":    "音乐社",
		"title":       "校园歌手大赛",
		"startTime":   "2025-12-01T18:00:00Z",
		"endTime":     "2025-12-01T21:00:00Z",
		"location":    "大礼堂",
		"capacity":    200,
		"description": "展示音乐才华",
	}, token)
	var activity models.ClubActivity
	decodeData(t, createResp, &activity)

	// 上传文件
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("fileType", "image")
	part, _ := writer.CreateFormFile("file", "活动海报.png")
	_, _ = part.Write([]byte("fake png data"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/files", activity.ID), body)
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

func TestActivitySummaryAndArchive(t *testing.T) {
	r, _ := setupActivityRouter(t)
	token := registerAndLogin(t, r)

	// 创建活动并推进到已完成
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/activities", map[string]any{
		"clubName":    "篮球社",
		"title":       "3v3篮球赛",
		"startTime":   "2025-11-10T09:00:00Z",
		"endTime":     "2025-11-10T17:00:00Z",
		"location":    "体育馆",
		"capacity":    64,
		"description": "三对三篮球对抗赛",
	}, token)
	var activity models.ClubActivity
	decodeData(t, createResp, &activity)

	// 推进到已完成
	doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/activities/%d/status", activity.ID), map[string]any{
		"status": "completed",
	}, token)

	// 提交总结
	summaryResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/activities/%d/summary", activity.ID), map[string]any{
		"summary": "本次活动共32支队伍参赛，现场气氛热烈，达到了预期效果。",
	}, token)
	if summaryResp.Code != http.StatusOK {
		t.Fatalf("summary expected 200, got %d: %s", summaryResp.Code, summaryResp.Body.String())
	}
	var archived models.ClubActivity
	decodeData(t, summaryResp, &archived)
	if archived.Status != models.ActivityArchived {
		t.Fatalf("expected archived status, got %s", archived.Status)
	}
	if archived.Summary == "" {
		t.Fatal("expected non-empty summary")
	}
}

func TestActivityStatistics(t *testing.T) {
	r, _ := setupActivityRouter(t)
	token := registerAndLogin(t, r)

	// 创建几个活动
	for i := 0; i < 3; i++ {
		doJSONRequest(r, http.MethodPost, "/api/v1/activities", map[string]any{
			"clubName":    fmt.Sprintf("社团%d", i),
			"title":       fmt.Sprintf("活动%d", i),
			"startTime":   "2025-09-01T09:00:00Z",
			"endTime":     "2025-09-01T18:00:00Z",
			"location":    "教室",
			"capacity":    50,
			"description": "测试活动",
		}, token)
	}

	// 获取统计
	statResp := doJSONRequest(r, http.MethodGet, "/api/v1/activities/statistics", nil, token)
	if statResp.Code != http.StatusOK {
		t.Fatalf("statistics expected 200, got %d: %s", statResp.Code, statResp.Body.String())
	}
}

func TestActivityDelete(t *testing.T) {
	r, _ := setupActivityRouter(t)
	token := registerAndLogin(t, r)

	// 创建活动
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/activities", map[string]any{
		"clubName":    "测试社团",
		"title":       "待删除活动",
		"startTime":   "2025-09-01T09:00:00Z",
		"endTime":     "2025-09-01T18:00:00Z",
		"location":    "教室",
		"capacity":    10,
		"description": "将被删除",
	}, token)
	var activity models.ClubActivity
	decodeData(t, createResp, &activity)

	// 删除草稿
	delResp := doJSONRequest(r, http.MethodDelete, fmt.Sprintf("/api/v1/activities/%d", activity.ID), nil, token)
	if delResp.Code != http.StatusOK {
		t.Fatalf("delete expected 200, got %d: %s", delResp.Code, delResp.Body.String())
	}

	// 验证已删除
	getResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/activities/%d", activity.ID), nil, token)
	if getResp.Code != http.StatusBadRequest {
		t.Fatalf("get deleted expected 400, got %d", getResp.Code)
	}
}
