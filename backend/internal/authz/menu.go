package authz

// MenuItem 描述一个前端菜单项；用于权限种子数据 + /users/me 接口返回
type MenuItem struct {
	Key      string `json:"key"`            // 菜单 key，唯一
	Title    string `json:"title"`          // 显示标题
	Path     string `json:"path"`           // 前端路由 path
	Icon     string `json:"icon,omitempty"` // 可选图标名
	ParentID string `json:"parentId,omitempty"`
}

// BuiltinMenus 全平台菜单清单
// 与前端 router/index.ts、MainLayout.vue 保持一致，由权限码控制可见性
var BuiltinMenus = []MenuItem{
	{Key: "dashboard", Title: "工作台", Path: "/"},
	{Key: "services", Title: "服务入口", Path: "/services"},
	{Key: "members", Title: "团员发展", Path: "/members"},
	{Key: "activities", Title: "社团活动", Path: "/activities"},
	{Key: "community", Title: "社区队伍", Path: "/community"},
	{Key: "community.duty", Title: "值班签到", Path: "/community/duty"},
	{Key: "community.profile", Title: "服务画像", Path: "/community/profile"},
	{Key: "workstudy", Title: "勤工助学", Path: "/workstudy"},
	{Key: "admin.users", Title: "用户管理", Path: "/admin/users"},
	{Key: "admin.roles", Title: "角色权限", Path: "/admin/roles"},
}

// menuPermissionCode 把菜单 key 转换为权限码
func MenuPermissionCode(key string) string { return "menu:" + key }

// RoleMenuMatrix 角色 → 拥有的菜单 key 列表
// 与 docs/09-rbac-design.md 第 4 章菜单矩阵保持一致
var RoleMenuMatrix = map[string][]string{
	"admin": {
		"dashboard", "services",
		"members", "activities", "community", "community.duty", "community.profile", "workstudy",
		"admin.users", "admin.roles",
	},
	"teacher": {
		"dashboard", "services",
		"members", "activities", "community", "community.duty", "community.profile", "workstudy",
	},
	"student_cadre": {
		"dashboard", "services",
		"members", "activities", "community", "community.duty", "community.profile", "workstudy",
	},
	"student": {
		"dashboard", "services",
		"community.duty", "community.profile", "workstudy",
	},
}

// BuiltinRoles 内置角色定义
var BuiltinRoles = []struct {
	Code        string
	Name        string
	DataScope   string
	Description string
}{
	{"admin", "系统管理员", "all", "平台运维、用户与角色管理"},
	{"teacher", "教师", "college", "辅导员/团委老师，负责审批与统计"},
	{"student_cadre", "学生干部", "team", "团支书、社团负责人、自治队伍队长"},
	{"student", "普通学生", "self", "在校生，仅查看本人相关信息"},
}
