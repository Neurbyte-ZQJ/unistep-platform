# Tests

本目录提供与后端集成的脚本化测试。

## 单元测试（Go）

直接进入 backend 目录运行：

```powershell
cd backend
$env:GOPROXY = 'https://goproxy.cn,direct'
go test ./...
```

> SQLite 驱动 `gorm.io/driver/sqlite` 依赖 cgo，请确保本机安装了 GCC（Windows 推荐 [tdm-gcc](https://jmeubank.github.io/tdm-gcc/) 或 mingw-w64）。

涵盖用例：

- `internal/router/auth_test.go`：注册/登录/JWT/角色（Sprint2 既有用例）
- `internal/router/member_test.go`（Sprint3 新增）
  - 团员档案 CRUD（含重复学号校验）
  - 入团申请 → 积极分子 → 发展对象 → 政审备案 → 电子档案聚合
  - 附件上传到 MinIO（使用内存 Uploader 桩件）
  - Swagger JSON / UI 端点可访问

## HTTP 集成测试（Python，无第三方依赖）

```powershell
# 1. 启动后端
cd backend
go run ./main.go

# 2. 新开终端
python tests\member_api_test.py http://localhost:8080
```

脚本会自动完成注册/登录、档案 CRUD、各阶段流程录入、电子档案生成校验，并验证 Swagger JSON。
若设置了 `MINIO_ENDPOINT`，还会执行附件上传用例。

## 测试数据生成（Go 脚本）

为开发/演示场景一次性生成中等规模业务数据，覆盖团员发展、社团活动、社区队伍、勤工助学全部模块。

```powershell
cd backend
$env:PATH = "D:\msys64\mingw64\bin;" + $env:PATH   # CGO 依赖
go run ./cmd/seed                                  # 增量生成（已存在则跳过）
go run ./cmd/seed -reset                           # 先清空业务表再重新生成（保留内置账号）
go run ./cmd/seed -db ./data/unistep.db -reset     # 指定数据库路径
```

数据规模：50 用户 / 45 团员档案 / 20 活动 / 10 队伍 / 20 勤工助学岗位，并自动生成报名、签到、值班、志愿、考勤、薪资等关联记录。

默认账号：

- `admin / admin123`（系统管理员）
- `teacher_wang / teacher123`（教师）
- `student_li / student123`（学生干部）
- `student_zhang|liu / student123`（普通学生）
- `stu_0001 ~ stu_0045 / test123`（脚本批量生成，其中每 10 位为学生干部，每 15 位为教师）
