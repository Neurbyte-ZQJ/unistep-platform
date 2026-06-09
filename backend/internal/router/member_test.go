package router

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/config"
	"unistep-platform/backend/internal/models"
)

// memoryUploader 是用于测试的 Uploader 实现，记录上传的对象到内存中
type memoryUploader struct {
	mu      sync.Mutex
	Objects map[string][]byte
}

func newMemoryUploader() *memoryUploader {
	return &memoryUploader{Objects: map[string][]byte{}}
}

func (m *memoryUploader) Disabled() bool { return false }

func (m *memoryUploader) Upload(_ context.Context, category, fileName string, _ int64, r io.Reader, _ string) (string, string, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return "", "", err
	}
	key := fmt.Sprintf("members/%s/%s", category, fileName)
	m.mu.Lock()
	m.Objects[key] = data
	m.mu.Unlock()
	return key, "memory://" + key, nil
}

func setupMemberRouter(t *testing.T) (*gin.Engine, *memoryUploader) {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.MemberProfile{},
		&models.LeagueApplication{},
		&models.ActivistRecord{},
		&models.DevelopTargetRecord{},
		&models.PoliticalReview{},
		&models.MemberAttachment{},
	); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	uploader := newMemoryUploader()
	r := New(config.Config{JWTSecret: "test-secret", FrontendURL: "http://localhost:5173"}, db, uploader)
	return r, uploader
}

// registerAndLogin 用于复用注册 + 登录获取 token 的流程
func registerAndLogin(t *testing.T, r *gin.Engine) string {
	t.Helper()
	doJSONRequest(r, http.MethodPost, "/api/v1/auth/register", map[string]string{
		"username": "tester",
		"password": "password123",
	}, "")
	w := doJSONRequest(r, http.MethodPost, "/api/v1/auth/login", map[string]string{
		"username": "tester",
		"password": "password123",
	}, "")
	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %s", w.Body.String())
	}
	var body struct {
		Data struct {
			Token string `json:"token"`
		} `json:"data"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &body)
	return body.Data.Token
}

func decodeData(t *testing.T, w *httptest.ResponseRecorder, out any) {
	t.Helper()
	var resp struct {
		Code    string          `json:"code"`
		Message string          `json:"message"`
		Data    json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response failed: %v body=%s", err, w.Body.String())
	}
	if out != nil {
		if err := json.Unmarshal(resp.Data, out); err != nil {
			t.Fatalf("decode data failed: %v", err)
		}
	}
}

// ---------- 测试用例 ----------

func TestMemberProfileCRUD(t *testing.T) {
	r, _ := setupMemberRouter(t)
	token := registerAndLogin(t, r)

	// 创建档案
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/members", map[string]any{
		"name":      "张三",
		"studentNo": "2024001",
		"gender":    "男",
		"college":   "计算机学院",
		"major":     "软件工程",
		"className": "软工2班",
	}, token)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create profile expected 201, got %d: %s", createResp.Code, createResp.Body.String())
	}
	var profile models.MemberProfile
	decodeData(t, createResp, &profile)
	if profile.ID == 0 {
		t.Fatalf("expected profile id, got 0")
	}

	// 重复学号 -> 失败
	dup := doJSONRequest(r, http.MethodPost, "/api/v1/members", map[string]any{
		"name":      "李四",
		"studentNo": "2024001",
	}, token)
	if dup.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 on duplicate studentNo, got %d", dup.Code)
	}

	// 列表查询
	listResp := doJSONRequest(r, http.MethodGet, "/api/v1/members?page=1&size=10", nil, token)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d", listResp.Code)
	}

	// 更新档案
	updateResp := doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/members/%d", profile.ID), map[string]any{
		"name":      "张三",
		"studentNo": "2024001",
		"phone":     "13800000000",
		"stage":     "activist",
	}, token)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update expected 200, got %d: %s", updateResp.Code, updateResp.Body.String())
	}

	// 详情
	getResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/members/%d", profile.ID), nil, token)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get expected 200, got %d", getResp.Code)
	}
}

func TestMemberDevelopmentFlow(t *testing.T) {
	r, _ := setupMemberRouter(t)
	token := registerAndLogin(t, r)

	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/members", map[string]any{
		"name":      "李四",
		"studentNo": "2024002",
	}, token)
	var profile models.MemberProfile
	decodeData(t, createResp, &profile)

	pid := profile.ID

	// 入团申请
	appResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/members/%d/applications", pid), map[string]any{
		"applyDate":  "2024-09-01",
		"motivation": "向团组织靠拢",
		"introducer": "王老师",
	}, token)
	if appResp.Code != http.StatusCreated {
		t.Fatalf("create application expected 201, got %d: %s", appResp.Code, appResp.Body.String())
	}

	// 积极分子
	actResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/members/%d/activists", pid), map[string]any{
		"startDate": "2024-10-01",
		"trainer":   "李辅导员",
		"trainPlan": "参加团课与社会实践",
		"score":     85.5,
	}, token)
	if actResp.Code != http.StatusCreated {
		t.Fatalf("create activist expected 201, got %d", actResp.Code)
	}

	// 发展对象
	devResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/members/%d/develop-targets", pid), map[string]any{
		"confirmedDate": "2025-03-01",
		"mentor":        "赵书记",
		"conclusion":    "同意公示",
	}, token)
	if devResp.Code != http.StatusCreated {
		t.Fatalf("create develop target expected 201, got %d", devResp.Code)
	}

	// 政审备案
	polResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/members/%d/political-reviews", pid), map[string]any{
		"reviewDate": "2025-04-01",
		"reviewer":   "校团委",
		"conclusion": "符合发展条件",
	}, token)
	if polResp.Code != http.StatusCreated {
		t.Fatalf("create political review expected 201, got %d", polResp.Code)
	}

	// 生成电子档案
	archiveResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/members/%d/archive", pid), nil, token)
	if archiveResp.Code != http.StatusOK {
		t.Fatalf("archive expected 200, got %d", archiveResp.Code)
	}
	var archive struct {
		Summary struct {
			ApplicationCount     int `json:"applicationCount"`
			ActivistCount        int `json:"activistCount"`
			DevelopRecordCount   int `json:"developRecordCount"`
			PoliticalRecordCount int `json:"politicalRecordCount"`
		} `json:"summary"`
		Timeline []map[string]string `json:"timeline"`
	}
	decodeData(t, archiveResp, &archive)
	if archive.Summary.ApplicationCount != 1 || archive.Summary.ActivistCount != 1 ||
		archive.Summary.DevelopRecordCount != 1 || archive.Summary.PoliticalRecordCount != 1 {
		t.Fatalf("archive summary unexpected: %+v", archive.Summary)
	}
	if len(archive.Timeline) != 4 {
		t.Fatalf("expected 4 timeline events, got %d", len(archive.Timeline))
	}
}

func TestMemberAttachmentUpload(t *testing.T) {
	r, uploader := setupMemberRouter(t)
	token := registerAndLogin(t, r)

	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/members", map[string]any{
		"name":      "王五",
		"studentNo": "2024003",
	}, token)
	var profile models.MemberProfile
	decodeData(t, createResp, &profile)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	_ = writer.WriteField("category", "application")
	part, _ := writer.CreateFormFile("file", "申请表.txt")
	_, _ = part.Write([]byte("hello minio"))
	_ = writer.Close()

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/v1/members/%d/attachments", profile.ID), body)
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

func TestSwaggerEndpoints(t *testing.T) {
	r, _ := setupMemberRouter(t)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/swagger.json", nil))
	if w.Code != http.StatusOK {
		t.Fatalf("swagger.json expected 200, got %d", w.Code)
	}
	if w.Header().Get("Content-Type") == "" {
		t.Fatal("missing content-type for swagger.json")
	}

	w2 := httptest.NewRecorder()
	r.ServeHTTP(w2, httptest.NewRequest(http.MethodGet, "/swagger", nil))
	if w2.Code != http.StatusOK {
		t.Fatalf("swagger ui expected 200, got %d", w2.Code)
	}
}
