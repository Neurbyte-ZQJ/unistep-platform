# 学生“一站式”自主管理过程管理系统（UniStep Platform）

> 面向高校团委、学生工作处与学生社区的数字化学生事务管理平台。
> 通过 Vibe Coding（AI 辅助开发）方式，覆盖团员发展、社团活动、社区自治、勤工助学四大业务领域，打造从需求 → 流程 → 数据 → 智能的“一站式”自主管理闭环。

---

## 目录

- [1. 工程信息](#1-工程信息)
- [2. 项目背景](#2-项目背景)
- [3. 功能介绍](#3-功能介绍)
- [4. 技术栈](#4-技术栈)
- [5. 目录结构](#5-目录结构)
- [6. 运行步骤](#6-运行步骤)
- [7. 环境变量](#7-环境变量)
- [8. 在线接口与文档](#8-在线接口与文档)
- [9. 迭代记录](#9-迭代记录)
- [10. 文档索引](#10-文档索引)

---

## 1. 工程信息

| 项目 | 描述 |
|------|------|
| 项目名称 | 学生“一站式”自主管理过程管理系统（UniStep Platform） |
| 项目代号 | unistep-platform |
| 项目类型 | 高校学生事务过程管理（B 端 + 学生 C 端） |
| 适用对象 | 高校团委、学生工作处、社区辅导员、社团负责人、自治队伍负责人、全体在校学生 |
| 架构形态 | 前后端分离 + 模块化单体 + 对象存储 |
| 部署方式 | Docker Compose 一键部署（Backend + Frontend + MinIO） |
| 主要文档 | 见 [`docs/`](./docs) 目录（PRD、BRD、SRD、架构、数据库、UI、Sprint、ADR） |
| 开发模式 | Vibe Coding（AI 辅助：需求分析 → 架构 → 数据库 → UI → 编码 → 迭代） |

---

## 2. 项目背景

高校学生事务管理普遍存在以下痛点：

- **流程碎片化**：入团申请走纸质、审批走微信、档案存 Excel，数据孤岛严重。
- **考勤不精准**：社团签到靠手写、勤工考勤靠打卡机，易代签漏签。
- **统计滞后**：月报/季报靠人工汇总，数据延迟 1-2 周。
- **档案缺失**：学生毕业时缺乏完整过程性记录，评优推荐缺乏数据支撑。
- **协作低效**：多部门业务交叉，依赖线下会议和群消息。

本项目目标是构建一套覆盖团员发展、社团活动、社区自治、勤工助学四大核心业务的数字化平台，实现：

1. **流程线上化**：4 大业务 100% 线上闭环；
2. **数据资产化**：过程数据可追溯、可分析；
3. **管理智能化**：AI 辅助档案生成、活动总结、评优推荐；
4. **体验一体化**：学生“一站式”入口完成所有事务。

详细背景与角色分析见 [`docs/01-project-analysis.md`](./docs/01-project-analysis.md)。

---

## 3. 功能介绍

系统围绕四大业务域提供端到端能力，并通过统一认证、文件存储、统计分析与 AI 能力进行支撑。

### 3.1 用户认证与权限

- 账号注册、登录、JWT 鉴权
- 基于角色（admin / 学生 / 老师 / 社团负责人 / 自治队伍负责人）的接口权限控制
- 个人信息查询（`/users/me`）
- 健康检查（`/health`）

### 3.2 团员发展模块

- 团员档案的增删改查、电子档案生成
- 入团申请提交、积极分子登记、发展对象记录、政审备案
- 思想汇报、证明材料等附件上传（MinIO）

### 3.3 社团活动模块

- 活动策划创建、编辑、删除、列表与详情
- 活动审批流（提交、通过/驳回、状态流转）
- 学生报名、取消报名、二维码/接口签到
- 活动图片上传、活动总结提交
- 活动数据统计

### 3.4 社区自治模块

- 自治队伍 CRUD、成员管理（增删改）
- 值班排班创建、值班签到/签退
- 志愿服务记录、服务时长审核
- 个人服务画像与队伍统计

### 3.5 勤工助学模块

- 岗位发布、编辑、上下架
- 学生报名、取消报名、录用/拒绝审核
- 在岗考勤打卡、签退、工时统计
- 工资计算与支付记录
- 岗位相关附件上传

### 3.6 统计分析与仪表盘

- 总览指标卡片（团员/活动/服务/勤工）
- 团员发展、社团活动、志愿服务等趋势图（基于 ECharts）
- 管理员专属仪表盘

### 3.7 AI 能力（规划中）

- AI 活动总结草稿
- AI 团员培养档案草稿
- AI 评优推荐理由
- 学生个人成长档案聚合

> 详细产品需求见 [`docs/02-prd.md`](./docs/02-prd.md)，业务需求见 [`docs/03-brd.md`](./docs/03-brd.md)，系统需求见 [`docs/04-srd.md`](./docs/04-srd.md)。

---

## 4. 技术栈

### 4.1 后端

| 组件 | 选型 | 版本 |
|------|------|------|
| 语言 | Go | 1.25 |
| Web 框架 | Gin | v1.10 |
| ORM | GORM | v1.25 |
| 数据库 | SQLite3（CGO） | v1.14 driver |
| 认证 | JWT（golang-jwt/v5） | v5.2 |
| 跨域 | gin-contrib/cors | v1.7 |
| 对象存储 | MinIO SDK | v7.2 |
| 密码加密 | golang.org/x/crypto/bcrypt | - |

### 4.2 前端

| 组件 | 选型 | 版本 |
|------|------|------|
| 框架 | Vue 3（Composition API + `<script setup lang="ts">`） | ^3.5 |
| 构建工具 | Vite | ^6.3 |
| UI 组件库 | Element Plus | ^2.9 |
| 图标 | @element-plus/icons-vue | ^2.3 |
| 状态管理 | Pinia | ^2.3 |
| 路由 | Vue Router | ^4.5 |
| 图表 | ECharts | ^6.1 |
| 语言 | TypeScript | ^5.8 |

### 4.3 部署 & 基础设施

- Docker / Docker Compose
- Nginx（前端静态托管 + 反向代理，配置见 [`docker/nginx.conf`](./docker/nginx.conf)）
- MinIO（对象存储，存储桶默认 `unistep`）

### 4.4 开发环境前置要求

- Go 1.25+，且因 SQLite 依赖 CGO，需要本机存在 C 编译器：
  - **Windows（项目规则）**：MSYS2 + MinGW-w64，路径 `D:\msys64\mingw64\bin`
  - 运行 Go 测试或编译前先设置 PATH：
    ```powershell
    $env:PATH = "D:\msys64\mingw64\bin;" + $env:PATH
    ```
- Node.js 18+ 与 npm
- （可选）Docker Desktop 24+

---

## 5. 目录结构

```text
unistep-platform/
├── backend/                         # Go 后端服务
│   ├── internal/
│   │   ├── config/                  # 配置加载（环境变量）
│   │   ├── database/                # SQLite + GORM 初始化与迁移
│   │   ├── handler/                 # Gin 业务处理（auth/member/activity/community/workstudy/dashboard/...）
│   │   ├── middleware/              # JWT、权限中间件
│   │   ├── models/                  # GORM 模型
│   │   ├── response/                # 统一响应封装 OK / Created / Fail
│   │   ├── router/                  # 路由注册与接口级测试
│   │   └── storage/                 # MinIO 上传器封装
│   ├── Dockerfile
│   ├── go.mod / go.sum
│   └── main.go                      # 服务入口
├── frontend/                        # Vue 3 前端
│   ├── src/
│   │   ├── api/                     # 各模块 API 封装
│   │   ├── layouts/                 # 主布局
│   │   ├── router/                  # 路由配置
│   │   ├── stores/                  # Pinia 状态管理（auth）
│   │   ├── styles/                  # 全局样式
│   │   └── views/                   # 业务页面（activities/community/members/workstudy/...）
│   ├── Dockerfile
│   ├── vite.config.ts
│   └── package.json
├── docker/
│   └── nginx.conf                   # 前端 Nginx 配置
├── docs/                            # 项目文档（PRD/BRD/SRD/架构/数据库/UI/Sprint）
│   ├── 01-project-analysis.md
│   ├── 02-prd.md
│   ├── 03-brd.md
│   ├── 04-srd.md
│   ├── 05-architecture.md
│   ├── 06-database-design.md
│   ├── 07-ui-design.md
│   ├── 08-sprint-plan.md
│   ├── api-spec.md
│   └── adr/                         # 架构决策记录（ADR-001~005）
├── sql/
│   └── init.sql                     # 初始化 SQL 脚本
├── tests/                           # 外部集成测试（Python）
│   └── member_api_test.py
├── docker-compose.yml               # 一键编排：backend + frontend + minio
└── README.md
```

---

## 6. 运行步骤

### 6.1 本地开发模式

#### 后端

```powershell
# Windows PowerShell：先注入 GCC（CGO 编译 SQLite 必需）
$env:PATH = "D:\msys64\mingw64\bin;" + $env:PATH

cd backend
go mod tidy
go run ./main.go
```

后端默认运行在 `http://localhost:8080`，健康检查：

```powershell
curl http://localhost:8080/health
# => {"status":"ok"}
```

#### 前端

```powershell
cd frontend
npm install
npm run dev
```

前端默认运行在 `http://localhost:5173`，开发模式下通过 Vite 直连后端 `http://localhost:8080`。

#### 运行后端测试

```powershell
$env:PATH = "D:\msys64\mingw64\bin;" + $env:PATH
cd backend
go test ./...
```

> 后端测试集中在 [`backend/internal/router/`](./backend/internal/router) 下，覆盖 auth / member / activity / community / workstudy / dashboard 等接口。

#### 运行前端类型检查与构建

```powershell
cd frontend
npm run build   # 等价于 vue-tsc -b && vite build
```

### 6.2 Docker Compose 一键部署

```powershell
docker compose up --build
```

启动后访问：

| 服务 | 地址 |
|------|------|
| 前端（Nginx） | http://localhost:8088 |
| 后端 API | http://localhost:8080/api/v1 |
| 后端健康检查 | http://localhost:8080/health |
| MinIO 控制台 | http://localhost:9001 （账号 `minioadmin` / `minioadmin`） |
| MinIO S3 接口 | http://localhost:9000 |

后端数据通过 `backend-data` 卷持久化（SQLite 文件位于 `/app/data/unistep.db`），MinIO 数据通过 `minio-data` 卷持久化。

---

## 7. 环境变量

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `PORT` | `8080` | 后端 HTTP 监听端口 |
| `DATABASE_PATH` | `data/unistep.db` | SQLite 数据库文件路径 |
| `JWT_SECRET` | `change-me-in-production` | JWT 签名密钥（生产务必替换） |
| `FRONTEND_URL` | `http://localhost:5173` | CORS 允许的前端来源 |
| `MINIO_ENDPOINT` | `minio:9000` | MinIO 服务地址（compose 内为服务名） |
| `MINIO_ACCESS_KEY` | `minioadmin` | MinIO Access Key |
| `MINIO_SECRET_KEY` | `minioadmin` | MinIO Secret Key |
| `MINIO_BUCKET` | `unistep` | 默认对象存储桶 |
| `MINIO_USE_SSL` | `false` | 是否启用 HTTPS 访问 MinIO |
| `MINIO_PUBLIC_URL` | `http://localhost:8088/minio` | 文件外链前缀（用于前端预览/下载） |

> 如果未配置 MinIO，附件相关接口会返回 `STORAGE_DISABLED`，但其他业务接口仍可正常运行。

---

## 8. 在线接口与文档

后端内置 Swagger 文档入口：

- Swagger UI：`http://localhost:8080/swagger`
- Swagger JSON：`http://localhost:8080/swagger.json`

主要接口分组（基于 [`backend/internal/router/router.go`](./backend/internal/router/router.go)）：

| 模块 | 前缀 | 描述 |
|------|------|------|
| 认证 | `/api/v1/auth` | 注册、登录 |
| 用户 | `/api/v1/users/me`、`/api/v1/me` | 个人信息 |
| 团员 | `/api/v1/members` | 档案、申请、积极分子、发展对象、政审、附件、电子档案 |
| 活动 | `/api/v1/activities` | 创建、审批、报名、签到、图片、总结、统计 |
| 社区 | `/api/v1/community/teams` | 队伍、成员、值班排班、志愿服务、统计、服务画像 |
| 勤工 | `/api/v1/workstudy` | 岗位、申请、审核、考勤、工资 |
| 仪表盘 | `/api/v1/dashboard` | 总览、各业务趋势 |
| 管理员 | `/api/v1/admin` | 需要 `admin` 角色 |

更完整的接口规范见 [`docs/api-spec.md`](./docs/api-spec.md)。

---

## 9. 迭代记录

本项目基于 Vibe Coding 流程演进，依据 [`docs/08-sprint-plan.md`](./docs/08-sprint-plan.md) 划分为 8 个 Sprint，每个 Sprint 默认 2 周。

| Sprint | 主题 | 关键交付 | 状态 |
|--------|------|---------|------|
| Sprint 1 | 项目初始化 | 前后端工程骨架、SQLite 初始化、统一响应封装、Docker Compose 基础、健康检查 `/health` | ✅ 已完成 |
| Sprint 2 | 用户权限模块 | 用户模型、注册/登录、JWT 中间件、`admin` 角色保护、个人信息接口 | ✅ 已完成 |
| Sprint 3 | 团员发展模块 | 团员档案 CRUD、入团申请、积极分子、发展对象、政审、附件上传、电子档案 | ✅ 已完成 |
| Sprint 4 | 社团活动模块 | 活动创建/审批/状态流转、报名/取消、签到、图片上传、活动总结、活动统计 | ✅ 已完成 |
| Sprint 5 | 自治队伍模块 | 队伍 CRUD、成员管理、值班排班与签到签退、志愿服务记录与审核、服务画像 | ✅ 已完成 |
| Sprint 6 | 勤工助学模块 | 岗位发布/上下架、报名/审核、考勤打卡与签退、工资计算与支付、岗位附件 | ✅ 已完成 |
| Sprint 7 | 统计分析模块 | Dashboard 总览、团员/活动/服务趋势、ECharts 可视化 | 🟢 进行中（基础接口与图表已上线） |
| Sprint 8 | AI 能力模块 | AI 活动总结、AI 培养档案、AI 评优推荐、成长档案 | 📋 规划中 |

### 9.1 里程碑

| 里程碑 | 对应 Sprint | 交付物 | 状态 |
|--------|------------|--------|------|
| M1：基础平台可运行 | S1 – S2 | 工程骨架、认证权限、基础部署 | ✅ |
| M2：核心业务 MVP | S3 – S4 | 团员发展、社团活动核心流程 | ✅ |
| M3：四大业务闭环 | S5 – S6 | 自治队伍、勤工助学模块 | ✅ |
| M4：数据化管理 | S7 | Dashboard、统计分析、导出 | 🟢 |
| M5：智能辅助上线 | S8 | AI 总结、AI 档案、AI 推荐 | 📋 |

### 9.2 关键架构决策（ADR）

| 编号 | 决策 | 文档 |
|------|------|------|
| ADR-001 | 选用 Go + Gin 作为后端技术栈 | [adr/ADR-001-why-go-gin.md](./docs/adr/ADR-001-why-go-gin.md) |
| ADR-002 | 选用 SQLite 作为 MVP 数据库 | [adr/ADR-002-why-sqlite.md](./docs/adr/ADR-002-why-sqlite.md) |
| ADR-003 | 选用 Vue 3 + Element Plus 作为前端方案 | [adr/ADR-003-why-vue3-element-plus.md](./docs/adr/ADR-003-why-vue3-element-plus.md) |
| ADR-004 | 选用 JWT 进行无状态认证 | [adr/ADR-004-why-jwt-auth.md](./docs/adr/ADR-004-why-jwt-auth.md) |
| ADR-005 | 采用前后端分离架构 | [adr/ADR-005-why-frontend-backend-separation.md](./docs/adr/ADR-005-why-frontend-backend-separation.md) |

### 9.3 开发约定（节选自 [`/.trae/rules/project_rules.md`](./.trae/rules/project_rules.md)）

- 后端 Handler 直接操作 GORM，未引入 Service 层（MVP 简化）
- 统一响应格式：`response.OK()` / `response.Created()` / `response.Fail()`
- 文件上传通过 `Uploader` 接口抽象，便于测试时替换为内存实现
- 前端统一使用 Composition API + `<script setup lang="ts">`
- 前端 API 调用统一封装在 `frontend/src/api/` 目录

---

## 10. 文档索引

| 类型 | 文档 |
|------|------|
| 项目分析 | [`docs/01-project-analysis.md`](./docs/01-project-analysis.md) |
| 产品需求 PRD | [`docs/02-prd.md`](./docs/02-prd.md) |
| 业务需求 BRD | [`docs/03-brd.md`](./docs/03-brd.md) |
| 系统需求 SRD | [`docs/04-srd.md`](./docs/04-srd.md) |
| 架构设计 | [`docs/05-architecture.md`](./docs/05-architecture.md) |
| 数据库设计 | [`docs/06-database-design.md`](./docs/06-database-design.md) |
| UI 设计 | [`docs/07-ui-design.md`](./docs/07-ui-design.md) |
| Sprint 计划 | [`docs/08-sprint-plan.md`](./docs/08-sprint-plan.md) |
| API 规范 | [`docs/api-spec.md`](./docs/api-spec.md) |
| 架构决策 | [`docs/adr/`](./docs/adr) |
| 项目规则 | [`/.trae/rules/project_rules.md`](./.trae/rules/project_rules.md) |
| 接口测试 | [`tests/README.md`](./tests/README.md) |

---

> 本项目为高校学生事务数字化治理样板工程，欢迎基于本仓库进行二次开发与定制。生产部署前请务必修改 `JWT_SECRET`、`MINIO_*` 等敏感配置，并启用 HTTPS。
