# 09 - RBAC 权限模型设计

> 文档版本：v1.0
> 适用系统：学生"一站式"自主管理过程管理系统（unistep-platform）
> 编制目的：解决当前管理端、教师端、学生端功能完全一致的设计缺陷，基于高校业务场景重新设计 RBAC（基于角色的访问控制）权限模型。

---

## 1. 现状问题分析

### 1.1 当前实现的缺陷

通过对 [router.go](file:///d:/Teach/AI_Coding/unistep-platform/backend/internal/router/router.go)、[jwt.go](file:///d:/Teach/AI_Coding/unistep-platform/backend/internal/middleware/jwt.go)、[user.go](file:///d:/Teach/AI_Coding/unistep-platform/backend/internal/models/user.go) 与前端 [router/index.ts](file:///d:/Teach/AI_Coding/unistep-platform/frontend/src/router/index.ts)、[MainLayout.vue](file:///d:/Teach/AI_Coding/unistep-platform/frontend/src/layouts/MainLayout.vue) 的分析，发现：

| 维度 | 问题 |
|------|------|
| **角色模型** | `users.roles` 仅用逗号分隔字符串存储；新注册用户默认 `student`，但所有受保护接口都用 `JWTAuth` 统一拦截，没有按角色细分 |
| **API 权限** | 仅 `/admin/dashboard` 使用 `RequireRole("admin")`；增、删、改、审批接口全部仅校验登录态 |
| **菜单/页面** | 前端 `MainLayout.vue` 对所有登录用户展示相同 6 个菜单；`router/index.ts` 没有 `meta.roles` 控制 |
| **数据权限** | 完全未实现"我创建的""我所在队伍的""全部数据"的范围隔离，普通学生可直接调用 `GET /api/v1/members` 列出所有团员 |
| **JWT 载荷** | 仅含 `userId/username/roles`，缺少 `permissions`、`dataScope`、`orgId` 等细粒度授权信息 |

### 1.2 设计目标

1. 区分 4 类角色：**系统管理员、教师、学生干部、普通学生**
2. 引入 **角色 → 权限点（菜单 + API）** 二级映射
3. 引入 **数据权限作用域（DataScope）**：全部 / 本学院 / 本队伍 / 本人
4. JWT 携带最小必要授权信息，避免每次请求查库
5. 数据库新增 `roles / permissions / role_permissions / user_roles` 四张表，沿用 GORM AutoMigrate

---

## 2. 角色定义

| 角色编码 | 角色名称 | 业务定位 | 典型用户 | 数据权限默认值 |
|----------|----------|----------|----------|----------------|
| `admin` | 系统管理员 | 平台运维、用户与角色管理、全局参数配置；不参与具体业务审批 | 学工部/团委系统管理员 | `all`（全部） |
| `teacher` | 教师（辅导员/团委老师） | 团员发展审批、活动审批、勤工助学岗位发布与审批、统计分析 | 辅导员、团委老师、勤工助学指导教师 | `college`（本学院/本部门） |
| `student_cadre` | 学生干部 | 录入团员材料、组织活动、管理自治队伍、登记志愿服务时长 | 团支书、社团负责人、自治队伍队长 | `team`（本队伍/本社团） |
| `student` | 普通学生 | 查看个人档案、报名活动、申请岗位、查看本人考勤与薪资 | 在校生 | `self`（仅本人） |

> 角色可叠加：例如某教师同时是 `admin`，则取权限并集；数据权限取**最宽**作用域。

---

## 3. 权限矩阵（功能模块 × 角色）

> 图例：✅ 允许 / ❌ 拒绝 / 🔒 仅限自己创建/参与的数据

### 3.1 用户与系统管理

| 功能点 | admin | teacher | student_cadre | student |
|--------|:-----:|:-------:|:-------------:|:-------:|
| 用户列表/创建/重置密码 | ✅ | ❌ | ❌ | ❌ |
| 角色分配 | ✅ | ❌ | ❌ | ❌ |
| 系统参数配置 | ✅ | ❌ | ❌ | ❌ |
| 查看个人资料 | ✅ | ✅ | ✅ | ✅ |
| 修改本人密码 | ✅ | ✅ | ✅ | ✅ |

### 3.2 团员发展（members）

| 功能点 | admin | teacher | student_cadre | student |
|--------|:-----:|:-------:|:-------------:|:-------:|
| 列表查询全部团员 | ✅ | ✅(本学院) | ✅(本班级) | ❌ |
| 查看团员详情 | ✅ | ✅(本学院) | ✅(本班级) | 🔒(本人) |
| 新建团员档案 | ✅ | ✅ | ✅ | ❌ |
| 编辑团员档案 | ✅ | ✅(本学院) | ✅(本班级) | ❌ |
| 删除团员档案 | ✅ | ❌ | ❌ | ❌ |
| 录入申请/积极分子/发展对象/政审 | ✅ | ✅(本学院) | ✅(本班级) | ❌ |
| 上传档案附件 | ✅ | ✅ | ✅ | 🔒(本人申请书) |
| 生成档案 PDF | ✅ | ✅(本学院) | ❌ | 🔒(本人) |

### 3.3 社团活动（activities）

| 功能点 | admin | teacher | student_cadre | student |
|--------|:-----:|:-------:|:-------------:|:-------:|
| 列表/查看活动 | ✅ | ✅ | ✅ | ✅ |
| 新建/编辑活动 | ✅ | ✅ | ✅(本社团) | ❌ |
| 删除活动 | ✅ | ✅(本学院) | 🔒(草稿) | ❌ |
| 提交审批 | ✅ | ❌ | ✅(本社团) | ❌ |
| **审批/驳回活动** | ✅ | ✅ | ❌ | ❌ |
| 上传活动图片/总结 | ✅ | ✅ | ✅(本社团) | ❌ |
| 报名活动 | ❌ | ❌ | ✅ | ✅ |
| 取消报名 | ❌ | ❌ | 🔒(本人) | 🔒(本人) |
| 活动签到 | ✅ | ✅ | ✅ | ✅ |
| 活动统计 | ✅ | ✅ | ✅(本社团) | ❌ |

### 3.4 学生社区与自治队伍（community）

| 功能点 | admin | teacher | student_cadre | student |
|--------|:-----:|:-------:|:-------------:|:-------:|
| 列表队伍/查看队伍 | ✅ | ✅ | ✅ | ✅ |
| 新建/解散队伍 | ✅ | ✅ | ❌ | ❌ |
| 编辑队伍信息 | ✅ | ✅ | ✅(本队伍) | ❌ |
| 添加/移除队员 | ✅ | ✅ | ✅(本队伍) | ❌ |
| 排班（DutySchedule） | ✅ | ✅ | ✅(本队伍) | ❌ |
| 值班签到/签退 | ✅ | ✅ | 🔒(本人) | 🔒(本人) |
| 登记志愿服务时长 | ✅ | ✅ | ✅(本队伍) | 🔒(本人) |
| **核实志愿服务时长** | ✅ | ✅ | ❌ | ❌ |
| 服务画像/统计 | ✅ | ✅ | ✅(本队伍) | 🔒(本人) |

### 3.5 勤工助学（workstudy）

| 功能点 | admin | teacher | student_cadre | student |
|--------|:-----:|:-------:|:-------------:|:-------:|
| 列表/查看岗位 | ✅ | ✅ | ✅ | ✅(仅 published) |
| 新建/编辑岗位 | ✅ | ✅ | ❌ | ❌ |
| 删除/关闭岗位 | ✅ | ✅(本部门) | ❌ | ❌ |
| 发布岗位 | ✅ | ✅ | ❌ | ❌ |
| 申请岗位 | ❌ | ❌ | ✅ | ✅ |
| 取消申请 | ❌ | ❌ | 🔒(本人) | 🔒(本人) |
| 录用/拒绝申请 | ✅ | ✅(本部门) | ❌ | ❌ |
| 考勤打卡 | ❌ | ❌ | 🔒(本人) | 🔒(本人) |
| 考勤列表 | ✅ | ✅(本部门) | ❌ | 🔒(本人) |
| 计算/发放薪资 | ✅ | ✅(本部门) | ❌ | ❌ |
| 查看薪资 | ✅ | ✅(本部门) | 🔒(本人) | 🔒(本人) |

### 3.6 统计仪表盘（dashboard）

| 功能点 | admin | teacher | student_cadre | student |
|--------|:-----:|:-------:|:-------------:|:-------:|
| 平台总览 Overview | ✅(全部) | ✅(本学院) | ✅(本队伍) | 🔒(个人画像) |
| 团员/活动/服务趋势图 | ✅ | ✅(本学院) | ✅(本队伍) | ❌ |

---

## 4. 菜单矩阵

> 前端 `MainLayout.vue` 中根据当前用户角色按下表过滤渲染。Pinia 在登录后将 `menus` 计算好缓存，路由守卫读取。

| 菜单 Key | 标题 | 路径 | admin | teacher | student_cadre | student |
|----------|------|------|:-----:|:-------:|:-------------:|:-------:|
| `dashboard` | 工作台 | `/` | ✅ | ✅ | ✅ | ✅ |
| `services` | 服务入口 | `/services` | ✅ | ✅ | ✅ | ✅ |
| `members` | 团员发展 | `/members` | ✅ | ✅ | ✅ | ❌ |
| `member.my` | 我的团员档案 | `/members/me` | ❌ | ❌ | ✅ | ✅ |
| `activities` | 社团活动管理 | `/activities` | ✅ | ✅ | ✅ | ❌ |
| `activity.square` | 活动广场（报名） | `/activities/square` | ✅ | ✅ | ✅ | ✅ |
| `activity.approve` | 活动审批 | `/activities/approval` | ✅ | ✅ | ❌ | ❌ |
| `community` | 社区队伍管理 | `/community` | ✅ | ✅ | ✅ | ❌ |
| `community.duty` | 值班签到 | `/community/duty` | ✅ | ✅ | ✅ | ✅ |
| `community.profile` | 服务画像 | `/community/profile` | ✅ | ✅ | ✅ | ✅ |
| `workstudy` | 勤工助学管理 | `/workstudy` | ✅ | ✅ | ❌ | ❌ |
| `workstudy.square` | 岗位广场 | `/workstudy/square` | ✅ | ✅ | ✅ | ✅ |
| `workstudy.my` | 我的岗位与薪资 | `/workstudy/me` | ❌ | ❌ | ✅ | ✅ |
| `admin.users` | 用户管理 | `/admin/users` | ✅ | ❌ | ❌ | ❌ |
| `admin.roles` | 角色权限 | `/admin/roles` | ✅ | ❌ | ❌ | ❌ |
| `admin.settings` | 系统设置 | `/admin/settings` | ✅ | ❌ | ❌ | ❌ |

---

## 5. 数据权限矩阵（DataScope）

定义 4 种作用域，由后端 Handler 在 GORM 查询前注入 `WHERE` 条件，避免越权读取。

| 作用域 | 编码 | 含义 | SQL 条件示例 |
|--------|------|------|--------------|
| 全部 | `all` | 不附加条件 | `1=1` |
| 本学院/部门 | `college` | 仅本 college/department 数据 | `college = :userCollege` |
| 本队伍 | `team` | 仅本人所在队伍/社团数据 | `team_id IN (:userTeamIds)` |
| 本人 | `self` | 仅本人创建或参与的数据 | `created_by = :userId OR student_id = :userId` |

### 5.1 各模块数据权限映射

| 资源 | 过滤字段 | admin | teacher | student_cadre | student |
|------|----------|:-----:|:-------:|:-------------:|:-------:|
| `member_profiles` | `college` / `class_name` | all | college | team(本班级) | self |
| `club_activities` | `club_name` / `created_by` | all | college | team | self(报名/签到) |
| `activity_registrations` | `student_id` | all | college | team | self |
| `community_teams` | `id ∈ user_teams` | all | college | team | self(已加入) |
| `team_members` | `team_id` | all | college | team | self |
| `duty_records` | `user_id` | all | college | team | self |
| `volunteer_services` | `team_id` / `user_id` | all | college | team | self |
| `workstudy_jobs` | `department` / `created_by` | all | college | (publish only) | self(published) |
| `job_applications` | `student_id` | all | college | self | self |
| `work_attendances` | `student_id` | all | college | self | self |
| `salary_records` | `student_id` | all | college | self | self |

---

## 6. 页面访问权限

前端 `router/index.ts` 在每个 `route.meta` 上挂载 `roles: string[]`，并在 `beforeEach` 守卫中校验。

```ts
// 示例：在 meta 中声明角色白名单
{
  path: 'admin/users',
  name: 'AdminUsers',
  component: () => import('../views/admin/UserListView.vue'),
  meta: { roles: ['admin'] },
}

router.beforeEach((to) => {
  const auth = useAuthStore()
  if (!to.meta.public && !auth.isAuthenticated) {
    return { name: 'Login', query: { redirect: to.fullPath } }
  }
  // 角色校验
  const allow = to.meta.roles as string[] | undefined
  if (allow && !allow.some((r) => auth.hasRole(r))) {
    return { name: 'Forbidden' }
  }
  return true
})
```

### 6.1 页面 → 角色白名单

| 路由 name | 路径 | 允许角色 |
|-----------|------|----------|
| `Login` / `Register` | `/login`, `/register` | public |
| `Dashboard` | `/` | 全部 |
| `Services` | `/services` | 全部 |
| `MemberList` | `/members` | admin, teacher, student_cadre |
| `MemberCreate` / `MemberEdit` | `/members/new`, `/members/:id/edit` | admin, teacher, student_cadre |
| `MemberDetail` | `/members/:id` | admin, teacher, student_cadre, student(self) |
| `MyMember` | `/members/me` | student_cadre, student |
| `ActivityList` | `/activities` | admin, teacher, student_cadre |
| `ActivityCreate` / `ActivityEdit` | `/activities/new`, `/activities/:id/edit` | admin, teacher, student_cadre |
| `ActivityApproval` | `/activities/approval` | admin, teacher |
| `ActivitySquare` | `/activities/square` | 全部 |
| `CommunityList` | `/community` | admin, teacher, student_cadre |
| `CommunityDuty` | `/community/duty` | 全部 |
| `ServiceProfile` | `/community/profile` | 全部 |
| `JobList` | `/workstudy` | admin, teacher |
| `JobCreate` / `JobEdit` | `/workstudy/new`, `/workstudy/:id/edit` | admin, teacher |
| `JobSquare` | `/workstudy/square` | 全部 |
| `MyJob` | `/workstudy/me` | student_cadre, student |
| `AdminUsers` / `AdminRoles` / `AdminSettings` | `/admin/*` | admin |
| `Forbidden` | `/403` | public |

---

## 7. API 访问权限

### 7.1 设计思路

在 `middleware` 包新增 `RequirePermission(code string)` 与 `RequireAnyRole(roles ...string)`：

```go
// 推荐写法
admin.POST("/users", middleware.RequirePermission("user:create"), userHandler.CreateUser)
activities.POST("/:id/approve", middleware.RequirePermission("activity:approve"), activityHandler.ApproveActivity)
```

中间件流程：
1. 从 `c.GetString("permissions")` 读取已登录用户授权字符串（来自 JWT）
2. 比对所需权限码；不通过则返回 `403 FORBIDDEN`

### 7.2 完整 API 权限表

| HTTP & 路径 | 权限码 | admin | teacher | student_cadre | student |
|-------------|--------|:-----:|:-------:|:-------------:|:-------:|
| `POST /api/v1/auth/register` | (public) | ✅ | ✅ | ✅ | ✅ |
| `POST /api/v1/auth/login` | (public) | ✅ | ✅ | ✅ | ✅ |
| `GET  /api/v1/users/me` | `user:read:self` | ✅ | ✅ | ✅ | ✅ |
| `GET  /api/v1/admin/users` | `user:list` | ✅ | ❌ | ❌ | ❌ |
| `POST /api/v1/admin/users` | `user:create` | ✅ | ❌ | ❌ | ❌ |
| `PUT  /api/v1/admin/users/:id/roles` | `user:assignRole` | ✅ | ❌ | ❌ | ❌ |
| **团员发展** | | | | | |
| `GET  /api/v1/members` | `member:list` | ✅ | ✅ | ✅ | ❌ |
| `POST /api/v1/members` | `member:create` | ✅ | ✅ | ✅ | ❌ |
| `GET  /api/v1/members/:id` | `member:read` | ✅ | ✅ | ✅ | 🔒 |
| `PUT  /api/v1/members/:id` | `member:update` | ✅ | ✅ | ✅ | ❌ |
| `DELETE /api/v1/members/:id` | `member:delete` | ✅ | ❌ | ❌ | ❌ |
| `GET  /api/v1/members/:id/archive` | `member:export` | ✅ | ✅ | ❌ | 🔒 |
| `POST /api/v1/members/:id/applications` | `member:application` | ✅ | ✅ | ✅ | 🔒 |
| `POST /api/v1/members/:id/activists` | `member:activist` | ✅ | ✅ | ✅ | ❌ |
| `POST /api/v1/members/:id/develop-targets` | `member:develop` | ✅ | ✅ | ✅ | ❌ |
| `POST /api/v1/members/:id/political-reviews` | `member:political` | ✅ | ✅ | ❌ | ❌ |
| `POST /api/v1/members/:id/attachments` | `member:upload` | ✅ | ✅ | ✅ | 🔒 |
| **社团活动** | | | | | |
| `GET  /api/v1/activities` | `activity:list` | ✅ | ✅ | ✅ | ✅ |
| `POST /api/v1/activities` | `activity:create` | ✅ | ✅ | ✅ | ❌ |
| `GET  /api/v1/activities/:id` | `activity:read` | ✅ | ✅ | ✅ | ✅ |
| `PUT  /api/v1/activities/:id` | `activity:update` | ✅ | ✅ | 🔒 | ❌ |
| `DELETE /api/v1/activities/:id` | `activity:delete` | ✅ | ✅ | 🔒 | ❌ |
| `POST /api/v1/activities/:id/submit` | `activity:submit` | ✅ | ❌ | 🔒 | ❌ |
| `POST /api/v1/activities/:id/approve` | `activity:approve` | ✅ | ✅ | ❌ | ❌ |
| `POST /api/v1/activities/:id/register` | `activity:register` | ❌ | ❌ | ✅ | ✅ |
| `POST /api/v1/activities/:id/cancel-registration` | `activity:cancel` | ❌ | ❌ | 🔒 | 🔒 |
| `POST /api/v1/activities/:id/checkin` | `activity:checkin` | ✅ | ✅ | ✅ | ✅ |
| `POST /api/v1/activities/:id/files` | `activity:upload` | ✅ | ✅ | 🔒 | ❌ |
| `POST /api/v1/activities/:id/summary` | `activity:summary` | ✅ | ✅ | 🔒 | ❌ |
| `PUT  /api/v1/activities/:id/status` | `activity:status` | ✅ | ✅ | ❌ | ❌ |
| `GET  /api/v1/activities/statistics` | `activity:statistics` | ✅ | ✅ | 🔒 | ❌ |
| **社区队伍** | | | | | |
| `GET  /api/v1/community/teams` | `team:list` | ✅ | ✅ | ✅ | ✅ |
| `POST /api/v1/community/teams` | `team:create` | ✅ | ✅ | ❌ | ❌ |
| `PUT  /api/v1/community/teams/:id` | `team:update` | ✅ | ✅ | 🔒 | ❌ |
| `DELETE /api/v1/community/teams/:id` | `team:delete` | ✅ | ✅ | ❌ | ❌ |
| `POST /api/v1/community/teams/:id/members` | `team:addMember` | ✅ | ✅ | 🔒 | ❌ |
| `DELETE /api/v1/community/teams/:id/members/:memberId` | `team:removeMember` | ✅ | ✅ | 🔒 | ❌ |
| `POST /api/v1/community/teams/:id/duties` | `duty:create` | ✅ | ✅ | 🔒 | ❌ |
| `POST .../duties/:scheduleId/checkin` | `duty:checkin` | ✅ | ✅ | 🔒 | 🔒 |
| `POST .../duties/:scheduleId/checkout` | `duty:checkout` | ✅ | ✅ | 🔒 | 🔒 |
| `POST /api/v1/community/teams/:id/services` | `service:create` | ✅ | ✅ | 🔒 | 🔒 |
| `PUT  .../services/:serviceId/verify` | `service:verify` | ✅ | ✅ | ❌ | ❌ |
| **勤工助学** | | | | | |
| `GET  /api/v1/workstudy/jobs` | `job:list` | ✅ | ✅ | ✅ | ✅(published) |
| `POST /api/v1/workstudy/jobs` | `job:create` | ✅ | ✅ | ❌ | ❌ |
| `PUT  /api/v1/workstudy/jobs/:id` | `job:update` | ✅ | ✅ | ❌ | ❌ |
| `DELETE /api/v1/workstudy/jobs/:id` | `job:delete` | ✅ | ✅ | ❌ | ❌ |
| `POST /api/v1/workstudy/jobs/:id/publish` | `job:publish` | ✅ | ✅ | ❌ | ❌ |
| `POST /api/v1/workstudy/jobs/:id/close` | `job:close` | ✅ | ✅ | ❌ | ❌ |
| `POST /api/v1/workstudy/jobs/:id/apply` | `job:apply` | ❌ | ❌ | ✅ | ✅ |
| `POST .../cancel-application` | `job:cancel` | ❌ | ❌ | 🔒 | 🔒 |
| `GET  /api/v1/workstudy/jobs/:id/applications` | `job:listApp` | ✅ | ✅ | ❌ | ❌ |
| `POST .../applications/:appId/accept` | `job:accept` | ✅ | ✅ | ❌ | ❌ |
| `POST .../applications/:appId/reject` | `job:reject` | ✅ | ✅ | ❌ | ❌ |
| `POST /api/v1/workstudy/jobs/:id/attendances` | `attendance:create` | ❌ | ❌ | 🔒 | 🔒 |
| `GET  /api/v1/workstudy/jobs/:id/attendances` | `attendance:list` | ✅ | ✅ | 🔒 | 🔒 |
| `PUT  /api/v1/workstudy/attendances/:attId/checkout` | `attendance:checkout` | ❌ | ❌ | 🔒 | 🔒 |
| `POST /api/v1/workstudy/jobs/:id/salary/calculate` | `salary:calc` | ✅ | ✅ | ❌ | ❌ |
| `GET  /api/v1/workstudy/jobs/:id/salary` | `salary:list` | ✅ | ✅ | 🔒 | 🔒 |
| `PUT  /api/v1/workstudy/salary/:salaryId/pay` | `salary:pay` | ✅ | ✅ | ❌ | ❌ |
| **仪表盘** | | | | | |
| `GET  /api/v1/dashboard/overview` | `dashboard:overview` | ✅ | ✅ | ✅ | 🔒 |
| `GET  /api/v1/dashboard/member-trend` | `dashboard:member` | ✅ | ✅ | ❌ | ❌ |
| `GET  /api/v1/dashboard/activity-trend` | `dashboard:activity` | ✅ | ✅ | 🔒 | ❌ |
| `GET  /api/v1/dashboard/service-trend` | `dashboard:service` | ✅ | ✅ | 🔒 | 🔒 |
| `GET  /api/v1/admin/dashboard` | `admin:dashboard` | ✅ | ❌ | ❌ | ❌ |

---

## 8. JWT 权限设计

### 8.1 当前 Claims

[auth.go](file:///d:/Teach/AI_Coding/unistep-platform/backend/internal/handler/auth.go#L108-L113) 现有：

```go
claims := jwt.MapClaims{
  "userId":   user.ID,
  "username": user.Username,
  "roles":    user.Roles,
  "exp":      time.Now().Add(24 * time.Hour).Unix(),
}
```

### 8.2 优化后的 Claims

| 字段 | 类型 | 说明 |
|------|------|------|
| `sub` | string | 用户 ID（标准字段，替换 `userId`） |
| `username` | string | 用户名（展示用） |
| `roles` | []string | 角色编码数组，如 `["teacher","student_cadre"]` |
| `perms` | []string | 权限码集合（去重并集），如 `["activity:approve","member:list"]` |
| `scope` | string | 数据作用域：`all` / `college` / `team` / `self`（取最宽） |
| `college` | string | 所属学院/部门（用于 `college` scope 查询） |
| `teamIds` | []uint | 所在队伍 ID 列表（用于 `team` scope 查询） |
| `iat` | int64 | 签发时间 |
| `exp` | int64 | 过期时间（建议 2h；引入 refresh token 续签） |
| `jti` | string | JWT ID，配合 Redis 黑名单实现注销 |

### 8.3 生成示例

```go
perms, scope, college, teamIds := authz.Resolve(db, user) // 从 user_roles → role_permissions 聚合

claims := jwt.MapClaims{
    "sub":      user.ID,
    "username": user.Username,
    "roles":    strings.Split(user.Roles, ","),
    "perms":    perms,
    "scope":    scope,
    "college":  college,
    "teamIds":  teamIds,
    "iat":      time.Now().Unix(),
    "exp":      time.Now().Add(2 * time.Hour).Unix(),
    "jti":      uuid.NewString(),
}
```

### 8.4 体积控制

- 当 `perms` 数量较多（>50）时，将 `perms` 改为 `permsHash`（如 SHA256 前 16 位），由后端启动时把 `hash → []perm` 缓存到内存
- 学生用户通常 `perms` 不超过 15 条，可直接内嵌

### 8.5 刷新与注销

- 颁发短期 `access_token`（2h）+ 长期 `refresh_token`（7d，httpOnly Cookie）
- 注销时把 `jti` 写入 Redis 黑名单，中间件每次校验时查询；若不引入 Redis，可在 `users` 表加 `token_version` 字段，重置密码/角色变更时 +1，JWT 中带 `tv` 校验

---

## 9. 数据库表设计调整方案

### 9.1 现状

[user.go](file:///d:/Teach/AI_Coding/unistep-platform/backend/internal/models/user.go) 中 `users.roles` 用逗号分隔字符串，无法描述权限点、数据作用域、组织归属。

### 9.2 调整目标

1. `users` 表保留登录信息，新增组织字段
2. 角色（`roles`）、权限（`permissions`）、关系（`role_permissions`、`user_roles`）独立成表
3. 兼容已有 `users.roles` 字段：保留以快照 + 兜底，新逻辑以 `user_roles` 为准

### 9.3 表结构（GORM Model）

#### 9.3.1 用户表（修改）

```go
// User 用户模型
type User struct {
    ID         uint      `gorm:"primarykey" json:"id"`
    Username   string    `gorm:"uniqueIndex;size:64;not null" json:"username"`
    Password   string    `gorm:"size:255;not null" json:"-"`
    Email      string    `gorm:"size:128" json:"email"`
    Roles      string    `gorm:"size:255;default:student" json:"roles"`     // 兼容旧逻辑：保存角色编码快照
    RealName   string    `gorm:"size:64" json:"realName"`                    // 真实姓名
    College    string    `gorm:"size:64;index" json:"college"`               // 所属学院/部门，数据权限关键字段
    ClassName  string    `gorm:"size:64;index" json:"className"`             // 班级（学生）
    TokenVersion int     `gorm:"not null;default:1" json:"-"`                // 用于强制 token 失效
    Status     string    `gorm:"size:16;not null;default:active" json:"status"` // active/disabled
    CreatedAt  time.Time `json:"createdAt"`
    UpdatedAt  time.Time `json:"updatedAt"`
}
```

#### 9.3.2 新增：角色表

```go
// Role 角色
type Role struct {
    ID          uint      `gorm:"primarykey" json:"id"`
    Code        string    `gorm:"uniqueIndex;size:32;not null" json:"code"`   // admin/teacher/student_cadre/student
    Name        string    `gorm:"size:64;not null" json:"name"`               // 中文名
    DataScope   string    `gorm:"size:16;not null;default:self" json:"dataScope"` // all/college/team/self
    Description string    `gorm:"type:text" json:"description"`
    Builtin     bool      `gorm:"not null;default:false" json:"builtin"`      // 内置角色不可删除
    CreatedAt   time.Time `json:"createdAt"`
    UpdatedAt   time.Time `json:"updatedAt"`
}
```

#### 9.3.3 新增：权限表

```go
// Permission 权限点
type Permission struct {
    ID          uint      `gorm:"primarykey" json:"id"`
    Code        string    `gorm:"uniqueIndex;size:64;not null" json:"code"`   // 形如 activity:approve
    Name        string    `gorm:"size:128;not null" json:"name"`
    Module      string    `gorm:"size:32;not null;index" json:"module"`       // member/activity/community/workstudy
    Type        string    `gorm:"size:16;not null;default:api" json:"type"`   // api/menu/button
    Description string    `gorm:"type:text" json:"description"`
    CreatedAt   time.Time `json:"createdAt"`
}
```

#### 9.3.4 新增：角色-权限关联

```go
// RolePermission 角色-权限多对多
type RolePermission struct {
    RoleID       uint      `gorm:"primaryKey" json:"roleId"`
    PermissionID uint      `gorm:"primaryKey" json:"permissionId"`
    CreatedAt    time.Time `json:"createdAt"`
}
```

#### 9.3.5 新增：用户-角色关联

```go
// UserRole 用户-角色多对多
type UserRole struct {
    UserID    uint      `gorm:"primaryKey" json:"userId"`
    RoleID    uint      `gorm:"primaryKey" json:"roleId"`
    GrantedBy uint      `json:"grantedBy"`              // 授权人
    CreatedAt time.Time `json:"createdAt"`
}
```

### 9.4 SQL 初始化（init.sql 增量）

```sql
-- 角色表
CREATE TABLE IF NOT EXISTS roles (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    code        VARCHAR(32) NOT NULL UNIQUE,
    name        VARCHAR(64) NOT NULL,
    data_scope  VARCHAR(16) NOT NULL DEFAULT 'self',
    description TEXT,
    builtin     BOOLEAN NOT NULL DEFAULT 0,
    created_at  DATETIME,
    updated_at  DATETIME
);

-- 权限表
CREATE TABLE IF NOT EXISTS permissions (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    code        VARCHAR(64) NOT NULL UNIQUE,
    name        VARCHAR(128) NOT NULL,
    module      VARCHAR(32) NOT NULL,
    type        VARCHAR(16) NOT NULL DEFAULT 'api',
    description TEXT,
    created_at  DATETIME
);
CREATE INDEX idx_permissions_module ON permissions(module);

-- 角色-权限
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id       INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    created_at    DATETIME,
    PRIMARY KEY (role_id, permission_id)
);

-- 用户-角色
CREATE TABLE IF NOT EXISTS user_roles (
    user_id    INTEGER NOT NULL,
    role_id    INTEGER NOT NULL,
    granted_by INTEGER,
    created_at DATETIME,
    PRIMARY KEY (user_id, role_id)
);

-- users 表新增字段
ALTER TABLE users ADD COLUMN real_name      VARCHAR(64);
ALTER TABLE users ADD COLUMN college        VARCHAR(64);
ALTER TABLE users ADD COLUMN class_name     VARCHAR(64);
ALTER TABLE users ADD COLUMN token_version  INTEGER NOT NULL DEFAULT 1;
ALTER TABLE users ADD COLUMN status         VARCHAR(16) NOT NULL DEFAULT 'active';
CREATE INDEX IF NOT EXISTS idx_users_college    ON users(college);
CREATE INDEX IF NOT EXISTS idx_users_class_name ON users(class_name);

-- 内置角色
INSERT INTO roles (code, name, data_scope, builtin) VALUES
  ('admin',          '系统管理员', 'all',     1),
  ('teacher',        '教师',       'college', 1),
  ('student_cadre',  '学生干部',   'team',    1),
  ('student',        '普通学生',   'self',    1);
```

### 9.5 后端代码改造点

1. **AutoMigrate**：在 [database.go](file:///d:/Teach/AI_Coding/unistep-platform/backend/internal/database/database.go) 注册 `Role`、`Permission`、`RolePermission`、`UserRole`
2. **`internal/authz` 新包**：
   - `Resolve(db, user) (perms []string, scope string, ...)`：聚合用户所有角色 → 权限并集，作用域取最宽
   - `Scoped(db, c) *gorm.DB`：根据 `c.GetString("scope")` 自动注入 `WHERE` 条件，业务 handler 直接 `authz.Scoped(c.DB, c).Find(...)`
3. **`middleware/jwt.go` 扩展**：
   - 解析 `perms`、`scope`、`college`、`teamIds` 写入 Gin Context
   - 新增 `RequirePermission(code string)` 中间件
4. **`handler/auth.go` 登录改造**：从 `user_roles` + `role_permissions` 聚合后再签发 JWT
5. **`router/router.go` 接入**：所有写操作类接口替换 `RequireRole("admin")` 为 `RequirePermission("xxx:xxx")`

### 9.6 数据迁移注意事项

- 已有用户的 `users.roles` 字符串需通过一次性迁移脚本拆分后写入 `user_roles`
- 学生用户默认绑定 `student` 角色；学号前缀或特定用户名前缀可识别为 `student_cadre`（建议人工核对）
- 数据库为 SQLite + CGO，`ALTER TABLE` 仅支持新增列；如需修改字段需走"建新表 → 复制 → 改名"流程

---

## 10. 实施路线

| 阶段 | 目标 | 关键产出 |
|------|------|----------|
| **P1：建模** | 落地角色/权限四张表、AutoMigrate、内置角色 + 权限种子 | `models/rbac.go`、`sql/init.sql` |
| **P2：JWT** | 重写登录签发、解析；新增 `RequirePermission` 中间件 | `handler/auth.go`、`middleware/jwt.go`、`internal/authz` |
| **P3：API 落点** | 按第 7 章给所有接口挂权限码 | `router/router.go` 全量更新 |
| **P4：数据权限** | `authz.Scoped` 注入 GORM 查询 | 改造 `handler/member.go` 等所有 List/Get |
| **P5：前端** | `auth.ts` 增加 `permissions/menus/dataScope`；菜单与路由按角色过滤；新增 `/403` 页面、`/admin/*` 角色管理页 | `stores/auth.ts`、`router/index.ts`、`MainLayout.vue`、`views/admin/*` |
| **P6：测试** | 单元测试覆盖每个角色的 4 类 CRUD 是否符合矩阵 | `router/*_test.go` 新增角色矩阵用例 |

---

## 附录 A：权限码命名规范

`<模块>:<操作>[:<范围>]`

- 模块：`user/member/activity/team/duty/service/job/attendance/salary/dashboard/admin`
- 操作：`list/read/create/update/delete/approve/submit/publish/close/apply/cancel/accept/reject/checkin/checkout/upload/export/statistics/verify/calc/pay/assignRole`
- 范围（可选）：`self`（仅本人）

约 70 个权限码可覆盖现有所有接口，详见第 7 章。
