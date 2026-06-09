package router

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/config"
	"unistep-platform/backend/internal/models"
)

func setupCommunityRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test db failed: %v", err)
	}
	if err := db.AutoMigrate(
		&models.User{},
		&models.CommunityTeam{},
		&models.TeamMember{},
		&models.DutySchedule{},
		&models.DutyRecord{},
		&models.VolunteerService{},
	); err != nil {
		t.Fatalf("migrate failed: %v", err)
	}
	r := New(config.Config{JWTSecret: "test-secret", FrontendURL: "http://localhost:5173"}, db, nil)
	return r
}

// ---------- 队伍 CRUD 测试 ----------

func TestTeamCRUD(t *testing.T) {
	r := setupCommunityRouter(t)
	token := registerAndLogin(t, r)

	// 创建队伍
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/community/teams", map[string]any{
		"name":        "学生自律委员会",
		"teamType":    "autonomy",
		"description": "负责学生社区日常管理",
		"quota":       30,
		"location":    "学生社区服务中心",
		"contactInfo": "community@example.com",
	}, token)
	if createResp.Code != http.StatusCreated {
		t.Fatalf("create team expected 201, got %d: %s", createResp.Code, createResp.Body.String())
	}
	var team models.CommunityTeam
	decodeData(t, createResp, &team)
	if team.ID == 0 {
		t.Fatal("expected team id, got 0")
	}
	if team.Status != "active" {
		t.Fatalf("expected active status, got %s", team.Status)
	}

	// 列表查询
	listResp := doJSONRequest(r, http.MethodGet, "/api/v1/community/teams?page=1&size=10", nil, token)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d", listResp.Code)
	}

	// 按类型过滤
	listByType := doJSONRequest(r, http.MethodGet, "/api/v1/community/teams?teamType=autonomy", nil, token)
	if listByType.Code != http.StatusOK {
		t.Fatalf("list by type expected 200, got %d", listByType.Code)
	}

	// 详情
	getResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/community/teams/%d", team.ID), nil, token)
	if getResp.Code != http.StatusOK {
		t.Fatalf("get expected 200, got %d", getResp.Code)
	}

	// 更新
	updateResp := doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/community/teams/%d", team.ID), map[string]any{
		"name":        "学生自律委员会(更新)",
		"teamType":    "autonomy",
		"description": "负责学生社区日常管理与服务",
		"quota":       35,
		"location":    "学生社区服务中心2楼",
		"contactInfo": "community2@example.com",
	}, token)
	if updateResp.Code != http.StatusOK {
		t.Fatalf("update expected 200, got %d: %s", updateResp.Code, updateResp.Body.String())
	}

	// 解散（软删除）
	delResp := doJSONRequest(r, http.MethodDelete, fmt.Sprintf("/api/v1/community/teams/%d", team.ID), nil, token)
	if delResp.Code != http.StatusOK {
		t.Fatalf("delete expected 200, got %d: %s", delResp.Code, delResp.Body.String())
	}
}

// ---------- 成员管理（纳新换届）测试 ----------

func TestTeamMemberManagement(t *testing.T) {
	r := setupCommunityRouter(t)
	token := registerAndLogin(t, r)

	// 创建队伍
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/community/teams", map[string]any{
		"name":     "志愿服务队",
		"teamType": "volunteer",
		"quota":    2,
	}, token)
	var team models.CommunityTeam
	decodeData(t, createResp, &team)

	// 纳新：添加成员1
	addResp1 := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/members", team.ID), map[string]any{
		"userId":    1,
		"name":      "张三",
		"studentNo": "2024001",
		"role":      "leader",
		"joinDate":  "2025-09-01",
		"termStart": "2025-09",
		"termEnd":   "2026-06",
	}, token)
	if addResp1.Code != http.StatusCreated {
		t.Fatalf("add member expected 201, got %d: %s", addResp1.Code, addResp1.Body.String())
	}

	// 纳新：添加成员2
	addResp2 := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/members", team.ID), map[string]any{
		"userId":    2,
		"name":      "李四",
		"studentNo": "2024002",
		"role":      "member",
		"joinDate":  "2025-09-01",
	}, token)
	if addResp2.Code != http.StatusCreated {
		t.Fatalf("add member2 expected 201, got %d: %s", addResp2.Code, addResp2.Body.String())
	}

	// 重复添加应失败
	dupResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/members", team.ID), map[string]any{
		"userId":    1,
		"name":      "张三",
		"studentNo": "2024001",
		"role":      "member",
		"joinDate":  "2025-09-01",
	}, token)
	if dupResp.Code != http.StatusBadRequest {
		t.Fatalf("duplicate member expected 400, got %d", dupResp.Code)
	}

	// 编制已满应失败
	addResp3 := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/members", team.ID), map[string]any{
		"userId":    3,
		"name":      "王五",
		"studentNo": "2024003",
		"role":      "member",
		"joinDate":  "2025-09-01",
	}, token)
	if addResp3.Code != http.StatusBadRequest {
		t.Fatalf("quota full expected 400, got %d", addResp3.Code)
	}

	// 查看成员列表
	listResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/community/teams/%d/members", team.ID), nil, token)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list members expected 200, got %d", listResp.Code)
	}

	// 换届：更新成员角色
	var member1 models.TeamMember
	decodeData(t, addResp1, &member1)
	updateMemberResp := doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/community/teams/%d/members/%d", team.ID, member1.ID), map[string]any{
		"userId":    1,
		"name":      "张三",
		"studentNo": "2024001",
		"role":      "vice",
		"joinDate":  "2025-09-01",
		"termStart": "2025-09",
		"termEnd":   "2026-06",
	}, token)
	if updateMemberResp.Code != http.StatusOK {
		t.Fatalf("update member expected 200, got %d: %s", updateMemberResp.Code, updateMemberResp.Body.String())
	}

	// 换届退出
	removeResp := doJSONRequest(r, http.MethodDelete, fmt.Sprintf("/api/v1/community/teams/%d/members/%d", team.ID, member1.ID), nil, token)
	if removeResp.Code != http.StatusOK {
		t.Fatalf("remove member expected 200, got %d: %s", removeResp.Code, removeResp.Body.String())
	}
}

// ---------- 值班管理测试 ----------

func TestDutyManagement(t *testing.T) {
	r := setupCommunityRouter(t)
	token := registerAndLogin(t, r)

	// 创建队伍
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/community/teams", map[string]any{
		"name":     "值班队伍",
		"teamType": "duty",
		"quota":    10,
	}, token)
	var team models.CommunityTeam
	decodeData(t, createResp, &team)

	// 添加成员
	addResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/members", team.ID), map[string]any{
		"userId":    1,
		"name":      "值班员A",
		"studentNo": "2024001",
		"role":      "member",
		"joinDate":  "2025-09-01",
	}, token)
	var member models.TeamMember
	decodeData(t, addResp, &member)

	// 创建值班安排
	scheduleResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/duties", team.ID), map[string]any{
		"date":      "2025-10-15",
		"startTime": "08:00",
		"endTime":   "12:00",
		"location":  "社区服务中心",
		"memberIds": []uint{member.ID},
	}, token)
	if scheduleResp.Code != http.StatusCreated {
		t.Fatalf("create duty schedule expected 201, got %d: %s", scheduleResp.Code, scheduleResp.Body.String())
	}
	var schedule models.DutySchedule
	decodeData(t, scheduleResp, &schedule)
	if len(schedule.Records) != 1 {
		t.Fatalf("expected 1 duty record, got %d", len(schedule.Records))
	}

	// 查看值班安排列表
	listResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/community/teams/%d/duties", team.ID), nil, token)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list duties expected 200, got %d", listResp.Code)
	}

	// 签到
	checkinResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/duties/%d/checkin", team.ID, schedule.ID), map[string]any{
		"userId": 1,
	}, token)
	if checkinResp.Code != http.StatusOK {
		t.Fatalf("checkin expected 200, got %d: %s", checkinResp.Code, checkinResp.Body.String())
	}

	// 重复签到应失败
	dupCheckinResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/duties/%d/checkin", team.ID, schedule.ID), map[string]any{
		"userId": 1,
	}, token)
	if dupCheckinResp.Code != http.StatusBadRequest {
		t.Fatalf("duplicate checkin expected 400, got %d", dupCheckinResp.Code)
	}

	// 签退（自动计算时长）
	checkoutResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/duties/%d/checkout", team.ID, schedule.ID), map[string]any{
		"userId": 1,
	}, token)
	if checkoutResp.Code != http.StatusOK {
		t.Fatalf("checkout expected 200, got %d: %s", checkoutResp.Code, checkoutResp.Body.String())
	}

	// 重复签退应失败
	dupCheckoutResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/duties/%d/checkout", team.ID, schedule.ID), map[string]any{
		"userId": 1,
	}, token)
	if dupCheckoutResp.Code != http.StatusBadRequest {
		t.Fatalf("duplicate checkout expected 400, got %d", dupCheckoutResp.Code)
	}
}

// ---------- 志愿服务测试 ----------

func TestVolunteerService(t *testing.T) {
	r := setupCommunityRouter(t)
	token := registerAndLogin(t, r)

	// 创建队伍
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/community/teams", map[string]any{
		"name":     "社区服务队",
		"teamType": "volunteer",
		"quota":    20,
	}, token)
	var team models.CommunityTeam
	decodeData(t, createResp, &team)

	// 记录志愿服务
	serviceResp := doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/services", team.ID), map[string]any{
		"userId":      1,
		"name":        "志愿者A",
		"studentNo":   "2024001",
		"title":       "社区清洁志愿活动",
		"date":        "2025-10-20",
		"hours":       3.5,
		"category":    "community",
		"description": "参与社区环境清洁",
	}, token)
	if serviceResp.Code != http.StatusCreated {
		t.Fatalf("create service expected 201, got %d: %s", serviceResp.Code, serviceResp.Body.String())
	}
	var service models.VolunteerService
	decodeData(t, serviceResp, &service)
	if service.Verified {
		t.Fatal("expected verified=false for new service")
	}

	// 查看志愿服务列表
	listResp := doJSONRequest(r, http.MethodGet, fmt.Sprintf("/api/v1/community/teams/%d/services", team.ID), nil, token)
	if listResp.Code != http.StatusOK {
		t.Fatalf("list services expected 200, got %d", listResp.Code)
	}

	// 核实志愿服务
	verifyResp := doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/community/teams/%d/services/%d/verify", team.ID, service.ID), map[string]any{
		"verified": true,
	}, token)
	if verifyResp.Code != http.StatusOK {
		t.Fatalf("verify service expected 200, got %d: %s", verifyResp.Code, verifyResp.Body.String())
	}
	var verified models.VolunteerService
	decodeData(t, verifyResp, &verified)
	if !verified.Verified {
		t.Fatal("expected verified=true after verification")
	}
}

// ---------- 统计与个人档案测试 ----------

func TestStatisticsAndProfile(t *testing.T) {
	r := setupCommunityRouter(t)
	token := registerAndLogin(t, r)

	// 创建队伍
	createResp := doJSONRequest(r, http.MethodPost, "/api/v1/community/teams", map[string]any{
		"name":     "统计测试队",
		"teamType": "autonomy",
		"quota":    10,
	}, token)
	var team models.CommunityTeam
	decodeData(t, createResp, &team)

	// 添加成员
	doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/members", team.ID), map[string]any{
		"userId":    1,
		"name":      "统计员",
		"studentNo": "2024001",
		"role":      "member",
		"joinDate":  "2025-09-01",
	}, token)

	// 记录志愿服务
	doJSONRequest(r, http.MethodPost, fmt.Sprintf("/api/v1/community/teams/%d/services", team.ID), map[string]any{
		"userId":    1,
		"name":      "统计员",
		"studentNo": "2024001",
		"title":     "社区服务",
		"date":      "2025-10-01",
		"hours":     4.0,
		"category":  "community",
	}, token)

	// 核实服务
	doJSONRequest(r, http.MethodPut, fmt.Sprintf("/api/v1/community/teams/%d/services/1/verify", team.ID), map[string]any{
		"verified": true,
	}, token)

	// 队伍统计
	statResp := doJSONRequest(r, http.MethodGet, "/api/v1/community/teams/statistics", nil, token)
	if statResp.Code != http.StatusOK {
		t.Fatalf("statistics expected 200, got %d: %s", statResp.Code, statResp.Body.String())
	}

	// 个人档案
	profileResp := doJSONRequest(r, http.MethodGet, "/api/v1/community/service-profile", nil, token)
	if profileResp.Code != http.StatusOK {
		t.Fatalf("profile expected 200, got %d: %s", profileResp.Code, profileResp.Body.String())
	}
}
