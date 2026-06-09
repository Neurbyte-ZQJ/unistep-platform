package router

import (
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/config"
	"unistep-platform/backend/internal/models"
)

func setupDashboardRouter(t *testing.T) *gin.Engine {
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
		&models.ClubActivity{},
		&models.ActivityRegistration{},
		&models.ActivityCheckin{},
		&models.ActivityFile{},
		&models.CommunityTeam{},
		&models.TeamMember{},
		&models.DutySchedule{},
		&models.DutyRecord{},
		&models.VolunteerService{},
		&models.WorkStudyJob{},
		&models.JobApplication{},
		&models.WorkAttendance{},
		&models.SalaryRecord{},
		&models.WorkStudyFile{},
	); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}

	return New(config.Config{JWTSecret: "test-secret", FrontendURL: "http://localhost:5173"}, db, nil)
}

// ---------- 测试用例 ----------

func TestDashboardOverviewEmpty(t *testing.T) {
	r := setupDashboardRouter(t)
	token := registerAndLogin(t, r)

	resp := doJSONRequest(r, http.MethodGet, "/api/v1/dashboard/overview", nil, token)
	if resp.Code != http.StatusOK {
		t.Fatalf("overview expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var data struct {
		Code string `json:"code"`
		Data struct {
			Members   struct{ Total int64 } `json:"members"`
			Activities struct{ Total int64 } `json:"activities"`
			Services  struct{ TotalHours float64 } `json:"services"`
			Workstudy struct{ TotalJobs int64 } `json:"workstudy"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &data); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if data.Data.Members.Total != 0 {
		t.Fatalf("expected 0 members, got %d", data.Data.Members.Total)
	}
	if data.Data.Activities.Total != 0 {
		t.Fatalf("expected 0 activities, got %d", data.Data.Activities.Total)
	}
	if data.Data.Workstudy.TotalJobs != 0 {
		t.Fatalf("expected 0 jobs, got %d", data.Data.Workstudy.TotalJobs)
	}
}

func TestDashboardOverviewWithData(t *testing.T) {
	r := setupDashboardRouter(t)
	token := registerAndLogin(t, r)

	// 创建团员档案
	for i := 0; i < 5; i++ {
		stage := "applicant"
		if i >= 3 {
			stage = "activist"
		}
		if i >= 4 {
			stage = "league_member"
		}
		doJSONRequest(r, http.MethodPost, "/api/v1/members", map[string]any{
			"userId":    1,
			"name":      fmt.Sprintf("团员%d", i),
			"studentNo": fmt.Sprintf("STU%03d", i),
			"stage":     stage,
		}, token)
	}

	// 创建活动
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

	// 创建勤工助学岗位
	doJSONRequest(r, http.MethodPost, "/api/v1/workstudy/jobs", map[string]any{
		"title":         "图书馆助理",
		"department":    "图书馆",
		"location":      "图书馆一楼",
		"description":   "协助图书整理",
		"quota":         2,
		"salaryPerHour": 15.0,
		"startTime":     "2025-09-01T08:00:00Z",
		"endTime":       "2026-01-31T18:00:00Z",
		"contactPerson": "张老师",
		"contactPhone":  "13800138000",
	}, token)

	// 获取概览
	resp := doJSONRequest(r, http.MethodGet, "/api/v1/dashboard/overview", nil, token)
	if resp.Code != http.StatusOK {
		t.Fatalf("overview expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var data struct {
		Code string `json:"code"`
		Data struct {
			Members struct {
				Total          int64  `json:"total"`
				StageBreakdown []struct {
					Stage string `json:"stage"`
					Count int64  `json:"count"`
				} `json:"stageBreakdown"`
			} `json:"members"`
			Activities struct {
				Total           int64 `json:"total"`
				TotalRegistrations int64 `json:"totalRegistrations"`
				TotalCheckins   int64 `json:"totalCheckins"`
			} `json:"activities"`
			Services struct {
				TotalServiceHours float64 `json:"totalServiceHours"`
				TotalDutyHours    float64 `json:"totalDutyHours"`
				TotalHours        float64 `json:"totalHours"`
			} `json:"services"`
			Workstudy struct {
				TotalJobs         int64   `json:"totalJobs"`
				TotalApplications int64   `json:"totalApplications"`
				TotalAccepted     int64   `json:"totalAccepted"`
				TotalSalaryPaid   float64 `json:"totalSalaryPaid"`
				TotalWorkHours    float64 `json:"totalWorkHours"`
			} `json:"workstudy"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &data); err != nil {
		t.Fatalf("decode failed: %v", err)
	}

	if data.Data.Members.Total != 5 {
		t.Fatalf("expected 5 members, got %d", data.Data.Members.Total)
	}
	if data.Data.Activities.Total != 3 {
		t.Fatalf("expected 3 activities, got %d", data.Data.Activities.Total)
	}
	if data.Data.Workstudy.TotalJobs != 1 {
		t.Fatalf("expected 1 job, got %d", data.Data.Workstudy.TotalJobs)
	}

	// 验证阶段分布
	stageMap := map[string]int64{}
	for _, s := range data.Data.Members.StageBreakdown {
		stageMap[s.Stage] = s.Count
	}
	if stageMap["applicant"] != 3 {
		t.Fatalf("expected 3 applicants, got %d", stageMap["applicant"])
	}
	if stageMap["activist"] != 1 {
		t.Fatalf("expected 1 activist, got %d", stageMap["activist"])
	}
	if stageMap["league_member"] != 1 {
		t.Fatalf("expected 1 league_member, got %d", stageMap["league_member"])
	}
}

func TestDashboardMemberTrend(t *testing.T) {
	r := setupDashboardRouter(t)
	token := registerAndLogin(t, r)

	// 创建团员
	for i := 0; i < 3; i++ {
		doJSONRequest(r, http.MethodPost, "/api/v1/members", map[string]any{
			"userId":    1,
			"name":      fmt.Sprintf("团员%d", i),
			"studentNo": fmt.Sprintf("TREND%03d", i),
			"stage":     "applicant",
		}, token)
	}

	resp := doJSONRequest(r, http.MethodGet, "/api/v1/dashboard/member-trend", nil, token)
	if resp.Code != http.StatusOK {
		t.Fatalf("member-trend expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var data struct {
		Code string `json:"code"`
		Data struct {
			Trend []struct {
				Month string `json:"month"`
				Count int64  `json:"count"`
			} `json:"trend"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &data); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(data.Data.Trend) == 0 {
		t.Fatal("expected at least 1 trend entry")
	}
}

func TestDashboardActivityTrend(t *testing.T) {
	r := setupDashboardRouter(t)
	token := registerAndLogin(t, r)

	// 创建活动
	for i := 0; i < 2; i++ {
		doJSONRequest(r, http.MethodPost, "/api/v1/activities", map[string]any{
			"clubName":    fmt.Sprintf("社团%d", i),
			"title":       fmt.Sprintf("趋势活动%d", i),
			"startTime":   "2025-09-01T09:00:00Z",
			"endTime":     "2025-09-01T18:00:00Z",
			"location":    "教室",
			"capacity":    50,
			"description": "测试活动",
		}, token)
	}

	resp := doJSONRequest(r, http.MethodGet, "/api/v1/dashboard/activity-trend", nil, token)
	if resp.Code != http.StatusOK {
		t.Fatalf("activity-trend expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var data struct {
		Code string `json:"code"`
		Data struct {
			Trend []struct {
				Month string `json:"month"`
				Count int64  `json:"count"`
			} `json:"trend"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &data); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	if len(data.Data.Trend) == 0 {
		t.Fatal("expected at least 1 trend entry")
	}
}

func TestDashboardServiceTrend(t *testing.T) {
	r := setupDashboardRouter(t)
	token := registerAndLogin(t, r)

	resp := doJSONRequest(r, http.MethodGet, "/api/v1/dashboard/service-trend", nil, token)
	if resp.Code != http.StatusOK {
		t.Fatalf("service-trend expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var data struct {
		Code string `json:"code"`
		Data struct {
			Volunteer []struct {
				Month string  `json:"month"`
				Hours float64 `json:"hours"`
			} `json:"volunteer"`
			Duty []struct {
				Month string  `json:"month"`
				Hours float64 `json:"hours"`
			} `json:"duty"`
		} `json:"data"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &data); err != nil {
		t.Fatalf("decode failed: %v", err)
	}
	// 空数据时也应正常返回空数组
}

func TestDashboardRequiresAuth(t *testing.T) {
	r := setupDashboardRouter(t)

	// 未登录访问应返回 401
	resp := doJSONRequest(r, http.MethodGet, "/api/v1/dashboard/overview", nil, "")
	if resp.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", resp.Code)
	}
}
