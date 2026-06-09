# UniStep Platform API 规范（Sprint5：社区队伍）

> 文档与 `backend/internal/handler/swagger.go` 中暴露的 OpenAPI JSON 保持一致。
> 服务启动后可通过以下入口访问可视化文档：
> - OpenAPI JSON：`GET /swagger.json`
> - Swagger UI：`GET /swagger`

## 全局约定

- 鉴权方式：`Authorization: Bearer <JWT>`（除 `/api/v1/auth/*` 外，所有 `/api/v1` 接口均需登录）
- 统一响应体：

```json
{ "code": "OK", "message": "success", "data": {} }
```

| HTTP | code 示例 | 含义 |
| ---- | --------- | ---- |
| 200/201 | `OK` / `CREATED` | 成功 |
| 400 | `INVALID_PARAMS` / `PROFILE_EXISTS` / `STORAGE_DISABLED` / `UPLOAD_FAILED` / `NOT_FOUND` / `INVALID_STATUS` / `CAPACITY_FULL` / `ALREADY_REGISTERED` / `ALREADY_CHECKED_IN` / `ALREADY_MEMBER` / `QUOTA_FULL` / `NOT_CHECKED_IN` / `ALREADY_CHECKED_OUT` | 业务错误 |
| 401 | `UNAUTHORIZED` | 未登录或 token 无效 |
| 403 | `FORBIDDEN` | 角色不足 |

## 团员发展接口

### 1. 列表查询

`GET /api/v1/members?page=1&size=10&stage=activist&name=张`

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| page | query | 否 | 页码，默认 1 |
| size | query | 否 | 每页数量，1~100，默认 10 |
| stage | query | 否 | 阶段过滤：`applicant`/`activist`/`develop_target`/`political_review`/`league_member` |
| name | query | 否 | 姓名模糊搜索 |

响应 `data` 示例：

```json
{
  "items": [
    { "id": 1, "name": "张三", "studentNo": "2024001", "stage": "activist", ... }
  ],
  "total": 1,
  "page": 1,
  "size": 10
}
```

### 2. 创建档案

`POST /api/v1/members`

请求体（`MemberProfileRequest`）：

```json
{
  "name": "张三",
  "studentNo": "2024001",
  "gender": "男",
  "college": "计算机学院",
  "major": "软件工程",
  "className": "软工2班",
  "stage": "applicant"
}
```

- `name`、`studentNo` 必填；`studentNo` 在系统内唯一。
- `stage` 可选，默认 `applicant`。

### 3. 获取档案详情

`GET /api/v1/members/{id}`

返回完整档案，包含 `applications`、`activistRecords`、`developRecords`、`politicalRecords`、`attachments` 五个关联集合。

### 4. 更新档案

`PUT /api/v1/members/{id}`，请求体同 `MemberProfileRequest`。

### 5. 删除档案

`DELETE /api/v1/members/{id}`

### 6. 入团申请

`POST /api/v1/members/{id}/applications`

```json
{ "applyDate": "2024-09-01", "motivation": "...", "introducer": "王老师" }
```

提交后档案 `stage` 若为空将自动推进为 `applicant`。

### 7. 积极分子培养

`POST /api/v1/members/{id}/activists`

```json
{ "startDate": "2024-10-01", "trainer": "李辅导员", "trainPlan": "...", "score": 85.5 }
```

档案阶段自动更新为 `activist`。

### 8. 发展对象

`POST /api/v1/members/{id}/develop-targets`

```json
{ "confirmedDate": "2025-03-01", "mentor": "赵书记", "conclusion": "同意公示" }
```

档案阶段自动更新为 `develop_target`。

### 9. 政审备案

`POST /api/v1/members/{id}/political-reviews`

```json
{ "reviewDate": "2025-04-01", "reviewer": "校团委", "conclusion": "符合发展条件" }
```

档案阶段自动更新为 `political_review`。

### 10. 团员电子档案

`GET /api/v1/members/{id}/archive`

聚合档案、附件、各阶段记录与汇总信息：

```json
{
  "profile": { ... },
  "timeline": [
    { "date": "2024-09-01", "stage": "入团申请", "text": "提交入团申请，状态=pending" }
  ],
  "summary": {
    "stage": "political_review",
    "applicationCount": 1,
    "activistCount": 1,
    "developRecordCount": 1,
    "politicalRecordCount": 1,
    "attachmentCount": 2
  }
}
```

### 11. 附件上传到 MinIO

`POST /api/v1/members/{id}/attachments`

- `Content-Type: multipart/form-data`
- 表单字段：
  - `category`：`application`/`activist`/`develop`/`political`/`other`
  - `file`：二进制文件

服务端通过 MinIO Go SDK 将文件写入 `MINIO_BUCKET`（默认 `unistep`），并在 `member_attachments` 表中保存元数据。
未配置 `MINIO_ENDPOINT` 时接口将返回 `STORAGE_DISABLED`。

## 社团活动接口

### 1. 活动列表

`GET /api/v1/activities?page=1&size=10&status=draft&clubName=计算机&title=编程`

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| page | query | 否 | 页码，默认 1 |
| size | query | 否 | 每页数量，1~100，默认 10 |
| status | query | 否 | 状态过滤：`draft`/`pending`/`rejected`/`reg_open`/`reg_closed`/`in_progress`/`completed`/`archived` |
| clubName | query | 否 | 社团名称模糊搜索 |
| title | query | 否 | 活动名称模糊搜索 |

响应 `data` 示例：

```json
{
  "items": [
    { "id": 1, "clubName": "计算机协会", "title": "编程马拉松", "status": "draft", ... }
  ],
  "total": 1,
  "page": 1,
  "size": 10
}
```

### 2. 创建活动

`POST /api/v1/activities`

请求体（`ActivityRequest`）：

```json
{
  "clubName": "计算机协会",
  "title": "编程马拉松",
  "startTime": "2025-09-01T09:00:00Z",
  "endTime": "2025-09-01T18:00:00Z",
  "location": "教学楼A101",
  "capacity": 50,
  "description": "24小时编程挑战赛",
  "budget": 2000.00
}
```

- `clubName`、`title`、`startTime`、`endTime`、`location`、`capacity`、`description` 必填。
- `budget` 可选。
- 创建后状态默认为 `draft`。

### 3. 获取活动详情

`GET /api/v1/activities/{id}`

返回完整活动信息，包含 `registrations`、`checkins`、`files` 三个关联集合。

### 4. 更新活动

`PUT /api/v1/activities/{id}`，请求体同 `ActivityRequest`。

仅草稿（`draft`）或已驳回（`rejected`）状态可编辑。

### 5. 删除活动

`DELETE /api/v1/activities/{id}`

仅草稿状态可删除，会级联删除关联的报名、签到和文件记录。

### 6. 提交审批

`POST /api/v1/activities/{id}/submit`

将活动从 `draft`/`rejected` 状态推进为 `pending`。

### 7. 审批活动

`POST /api/v1/activities/{id}/approve`

```json
{ "opinion": "活动方案合理，同意开展", "approve": true }
```

- `approve: true` → 状态变为 `reg_open`（报名开放）
- `approve: false` → 状态变为 `rejected`（已驳回）

### 8. 活动报名

`POST /api/v1/activities/{id}/register`

仅 `reg_open` 状态可报名。自动检查容量和重复报名。

### 9. 取消报名

`POST /api/v1/activities/{id}/cancel-registration`

### 10. 活动签到

`POST /api/v1/activities/{id}/checkin`

```json
{ "studentId": 1 }
```

活动处于 `reg_open`/`reg_closed`/`in_progress` 状态时允许签到。

### 11. 上传活动图片/文件

`POST /api/v1/activities/{id}/files`

- `Content-Type: multipart/form-data`
- 表单字段：
  - `fileType`：`image`/`document`/`summary`
  - `file`：二进制文件

### 12. 提交活动总结

`POST /api/v1/activities/{id}/summary`

```json
{ "summary": "本次活动共32支队伍参赛，现场气氛热烈。" }
```

仅 `completed` 状态可提交总结，提交后状态自动变为 `archived`。

### 13. 变更活动状态

`PUT /api/v1/activities/{id}/status`

```json
{ "status": "in_progress" }
```

手动推进活动状态，合法值：`draft`/`pending`/`rejected`/`reg_open`/`reg_closed`/`in_progress`/`completed`/`archived`。

### 14. 活动统计

`GET /api/v1/activities/statistics`

响应 `data` 示例：

```json
{
  "totalActivities": 10,
  "statusBreakdown": [
    { "status": "draft", "count": 3 },
    { "status": "reg_open", "count": 2 }
  ],
  "totalRegistrations": 150,
  "totalCheckins": 120,
  "recentActivities": [...]
}
```

## 数据库模型概览

| 表 | 关键字段 |
| --- | --- |
| `member_profiles` | id, user_id, name, student_no(唯一), gender, stage 等 |
| `league_applications` | id, profile_id, apply_date, motivation, status |
| `activist_records` | id, profile_id, start_date, trainer, train_plan, score |
| `develop_target_records` | id, profile_id, confirmed_date, mentor, conclusion |
| `political_reviews` | id, profile_id, review_date, reviewer, conclusion, status |
| `member_attachments` | id, profile_id, category, file_name, object_key, url, size |
| `club_activities` | id, club_name, title, start_time, end_time, location, capacity, description, budget, status, summary |
| `activity_registrations` | id, activity_id, student_id, status(registered/cancelled) |
| `activity_checkins` | id, activity_id, student_id, checkin_time, checkin_method |
| `activity_files` | id, activity_id, file_name, object_key, url, file_type, size |

所有表通过 `gorm.AutoMigrate` 在服务首次启动时自动创建。

## 学生社区与自治队伍接口

### 1. 队伍列表

`GET /api/v1/community/teams?page=1&size=10&teamType=autonomy&name=自律&status=active`

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| page | query | 否 | 页码，默认 1 |
| size | query | 否 | 每页数量，1~100，默认 10 |
| teamType | query | 否 | 类型过滤：`autonomy`/`volunteer`/`duty` |
| name | query | 否 | 队伍名称模糊搜索 |
| status | query | 否 | 状态过滤：`active`/`disbanded` |

响应 `data` 示例：

```json
{
  "items": [
    { "id": 1, "name": "学生自律委员会", "teamType": "autonomy", "quota": 30, "status": "active", ... }
  ],
  "total": 1,
  "page": 1,
  "size": 10
}
```

### 2. 创建队伍

`POST /api/v1/community/teams`

请求体（`TeamRequest`）：

```json
{
  "name": "学生自律委员会",
  "teamType": "autonomy",
  "description": "负责学生社区日常管理",
  "quota": 30,
  "location": "学生社区服务中心",
  "contactInfo": "community@example.com"
}
```

- `name`、`teamType` 必填；`teamType` 取值：`autonomy`/`volunteer`/`duty`。
- 创建后状态默认为 `active`。

### 3. 获取队伍详情

`GET /api/v1/community/teams/{id}`

返回完整队伍信息，包含 `members` 关联集合（仅在职成员）。

### 4. 更新队伍

`PUT /api/v1/community/teams/{id}`，请求体同 `TeamRequest`。

### 5. 解散队伍

`DELETE /api/v1/community/teams/{id}`

软删除：将状态标记为 `disbanded`。

### 6. 队伍成员列表

`GET /api/v1/community/teams/{id}/members?status=active&role=leader`

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| status | query | 否 | 状态过滤：`active`/`pending`/`left` |
| role | query | 否 | 角色过滤：`leader`/`vice`/`member`/`trainee` |

### 7. 添加成员（纳新）

`POST /api/v1/community/teams/{id}/members`

请求体（`TeamMemberRequest`）：

```json
{
  "userId": 1,
  "name": "张三",
  "studentNo": "2024001",
  "role": "member",
  "joinDate": "2025-09-01",
  "termStart": "2025-09",
  "termEnd": "2026-06"
}
```

- 自动检查重复加入和编制上限。

### 8. 更新成员信息（换届）

`PUT /api/v1/community/teams/{id}/members/{memberId}`

请求体同 `TeamMemberRequest`，用于更新角色、届次等信息。

### 9. 移除成员

`DELETE /api/v1/community/teams/{id}/members/{memberId}`

将成员状态标记为 `left`，并记录离队日期。

### 10. 值班安排列表

`GET /api/v1/community/teams/{id}/duties?page=1&size=10&date=2025-10-15&status=scheduled`

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| page | query | 否 | 页码，默认 1 |
| size | query | 否 | 每页数量，1~100，默认 10 |
| date | query | 否 | 日期过滤（yyyy-MM-dd） |
| status | query | 否 | 状态过滤：`scheduled`/`active`/`completed`/`absent` |

返回结果包含 `records` 关联集合。

### 11. 创建值班安排

`POST /api/v1/community/teams/{id}/duties`

请求体（`DutyScheduleRequest`）：

```json
{
  "date": "2025-10-15",
  "startTime": "08:00",
  "endTime": "12:00",
  "location": "社区服务中心",
  "memberIds": [1, 2]
}
```

- 自动为每个成员创建值班记录。

### 12. 值班签到

`POST /api/v1/community/teams/{id}/duties/{scheduleId}/checkin`

```json
{ "userId": 1 }
```

### 13. 值班签退（自动计算时长）

`POST /api/v1/community/teams/{id}/duties/{scheduleId}/checkout`

```json
{ "userId": 1 }
```

- 签退时自动计算值班时长（小时），基于签到时间和签退时间差。
- 记录状态自动变为 `completed`。

### 14. 志愿服务列表

`GET /api/v1/community/teams/{id}/services?page=1&size=10&category=community&verified=true`

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| page | query | 否 | 页码，默认 1 |
| size | query | 否 | 每页数量，1~100，默认 10 |
| category | query | 否 | 类别过滤 |
| verified | query | 否 | 核实状态过滤：`true`/`false` |

### 15. 记录志愿服务

`POST /api/v1/community/teams/{id}/services`

请求体（`VolunteerServiceRequest`）：

```json
{
  "userId": 1,
  "name": "志愿者A",
  "studentNo": "2024001",
  "title": "社区清洁志愿活动",
  "date": "2025-10-20",
  "hours": 3.5,
  "category": "community",
  "description": "参与社区环境清洁"
}
```

- 创建后 `verified` 默认为 `false`，需管理员核实。

### 16. 核实志愿服务

`PUT /api/v1/community/teams/{id}/services/{serviceId}/verify`

```json
{ "verified": true }
```

### 17. 队伍统计

`GET /api/v1/community/teams/statistics`

响应 `data` 示例：

```json
{
  "totalTeams": 5,
  "typeBreakdown": [
    { "teamType": "autonomy", "count": 2 },
    { "teamType": "volunteer", "count": 3 }
  ],
  "totalMembers": 80,
  "totalServiceHours": 256.5,
  "totalDutyHours": 120.0
}
```

### 18. 服务时长个人档案

`GET /api/v1/community/service-profile?userId=1`

| 参数 | 位置 | 必填 | 说明 |
| --- | --- | --- | --- |
| userId | query | 否 | 目标用户ID，不传则查看自己 |

响应 `data` 示例：

```json
{
  "userId": 1,
  "totalServiceHours": 12.5,
  "totalDutyHours": 8.0,
  "totalHours": 20.5,
  "services": [...],
  "dutyRecords": [...],
  "teamMemberships": [...]
}
```

## 数据库模型概览（Sprint5 新增）

| 表 | 关键字段 |
| --- | --- |
| `community_teams` | id, name, team_type, description, quota, location, contact_info, status, created_by |
| `team_members` | id, team_id, user_id, name, student_no, role, status, join_date, leave_date, term_start, term_end |
| `duty_schedules` | id, team_id, date, start_time, end_time, location, status, created_by |
| `duty_records` | id, schedule_id, team_id, user_id, name, checkin_time, checkout_time, duration, status |
| `volunteer_services` | id, team_id, user_id, name, student_no, title, date, hours, category, description, verified, verified_by |

## 环境变量（新增）

| 变量 | 默认值 | 说明 |
| --- | --- | --- |
| `MINIO_ENDPOINT` | _空_ | 形如 `localhost:9000`；为空时附件功能禁用 |
| `MINIO_ACCESS_KEY` | `minioadmin` | MinIO Access Key |
| `MINIO_SECRET_KEY` | `minioadmin` | MinIO Secret Key |
| `MINIO_BUCKET` | `unistep` | 存储桶名称，缺失时自动创建 |
| `MINIO_USE_SSL` | `false` | 是否使用 HTTPS |
| `MINIO_PUBLIC_URL` | _空_ | 可选，对外可访问地址前缀，用于拼接附件 URL |
