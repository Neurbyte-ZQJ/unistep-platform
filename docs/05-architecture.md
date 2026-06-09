# 系统架构设计文档 — 学生“一站式”自主管理过程管理系统

> 文档版本：v1.0  
> 创建日期：2026-06-05  
> 输入文档：`02-prd.md`、`04-srd.md`、`docs/adr/`  
> 架构决策：Go + Gin、SQLite、Vue3 + Element Plus、JWT、前后端分离  

---

## 1. 架构原则

1. **前后端分离**：Vue3 前端负责页面、路由、权限菜单和交互；Go + Gin 后端负责认证、授权、业务规则和数据访问。
2. **模块化单体优先**：MVP 阶段采用模块化单体架构，按业务域划分模块，保留后续微服务拆分边界。
3. **后端强制鉴权**：前端只控制菜单和按钮展示，所有接口权限和数据权限均由后端校验。
4. **低成本部署**：MVP 使用 Docker Compose 单机部署，SQLite 单文件数据库，MinIO 或本地文件存储。
5. **可演进设计**：通过 Repository 层隔离数据库实现，通过 S3 兼容接口隔离文件存储，通过 Service 层隔离业务流程。
6. **安全默认开启**：密码哈希、JWT、RBAC、输入校验、文件白名单、结构化日志作为基础能力。

---

## 2. 系统架构图

```mermaid
flowchart TB
    subgraph Client[客户端]
        Browser[PC 浏览器]
        Mobile[移动端 Web]
    end

    subgraph Frontend[前端层：Vue3 + Element Plus]
        SPA[单页应用 SPA]
        Router[Vue Router 路由]
        Store[Pinia 状态管理]
        Menu[动态菜单与按钮权限]
        Forms[业务表单/列表/审批页面]
    end

    subgraph Gateway[接入层]
        Nginx[Nginx / 静态资源服务 / 反向代理]
        HTTPS[HTTPS]
    end

    subgraph Backend[后端层：Go + Gin REST API]
        Middleware[中间件\nJWT认证 / RBAC鉴权 / RequestID / 日志 / CORS]
        Handler[Handler/API 层\n参数绑定 / 响应封装]
        Service[Service 层\n业务规则 / 状态流转 / 事务控制]
        Repo[Repository 层\nGORM 数据访问]
        Infra[Infrastructure\n配置 / JWT / bcrypt / 文件存储 / 日志]
    end

    subgraph Domain[业务域]
        Basic[基础能力\n用户认证 / RBAC / 文件上传]
        League[团员发展\n入团申请 / 审批 / 积极分子 / 培养记录]
        Club[社团活动\n活动审批 / 报名 / 二维码签到 / 图片上传]
        Community[社区自治\n队伍 / 纳新 / 排班]
        Work[勤工助学\n岗位 / 报名 / 录用 / 考勤]
    end

    subgraph Data[数据与存储层]
        SQLite[(SQLite 数据库)]
        FileStore[MinIO / 本地文件存储]
        LogStore[结构化日志]
        Backup[每日备份]
    end

    Browser --> HTTPS
    Mobile --> HTTPS
    HTTPS --> Nginx
    Nginx --> SPA
    SPA --> Router
    SPA --> Store
    Router --> Menu
    Router --> Forms
    Forms -->|Authorization: Bearer JWT / JSON API| Nginx
    Nginx -->|/api/v1| Middleware
    Middleware --> Handler
    Handler --> Service
    Service --> Basic
    Service --> League
    Service --> Club
    Service --> Community
    Service --> Work
    Service --> Repo
    Repo --> SQLite
    Service --> Infra
    Infra --> FileStore
    Middleware --> LogStore
    SQLite --> Backup
    FileStore --> Backup
```

---

## 3. 模块架构图

```mermaid
flowchart TB
    subgraph FE[前端模块]
        FEAuth[登录/注册]
        FERBAC[权限路由/动态菜单]
        FELeague[团员发展页面]
        FEClub[社团活动页面]
        FECommunity[社区自治页面]
        FEWork[勤工助学页面]
        FEFile[文件上传组件]
    end

    subgraph API[后端 API 模块]
        AuthAPI[Auth API\n/auth/login /auth/me]
        UserAPI[User API\n/users /roles /permissions]
        LeagueAPI[League API\n/league/applications]
        ClubAPI[Club API\n/club/activities]
        CommunityAPI[Community API\n/community/teams]
        WorkAPI[Work API\n/work/positions]
        FileAPI[File API\n/files]
    end

    subgraph Service[业务服务模块]
        AuthSvc[认证服务\n密码校验 / JWT签发]
        RbacSvc[权限服务\n角色权限 / 数据权限]
        LeagueSvc[团员发展服务\n申请状态机 / 审批 / 积极分子生成]
        ClubSvc[社团活动服务\n活动状态 / 报名约束 / 签到令牌]
        CommunitySvc[社区自治服务\n队伍 / 纳新 / 排班冲突检测]
        WorkSvc[勤工助学服务\n岗位录用 / 在岗约束 / 工时计算]
        FileSvc[文件服务\n类型校验 / 对象Key / 元数据]
    end

    subgraph Repo[数据访问模块]
        UserRepo[User/Role/Permission Repository]
        LeagueRepo[League Repository]
        ClubRepo[Club Repository]
        CommunityRepo[Community Repository]
        WorkRepo[Work Repository]
        FileRepo[File Repository]
    end

    subgraph DB[核心数据模型]
        Users[(users)]
        Roles[(roles / permissions)]
        LeagueTables[(league_applications / activists / cultivation_records)]
        ClubTables[(club_activities / registrations / checkins)]
        CommunityTables[(community_teams / recruitments / schedules)]
        WorkTables[(work_positions / applications / assignments / attendances)]
        Files[(files)]
    end

    FEAuth --> AuthAPI
    FERBAC --> UserAPI
    FELeague --> LeagueAPI
    FEClub --> ClubAPI
    FECommunity --> CommunityAPI
    FEWork --> WorkAPI
    FEFile --> FileAPI

    AuthAPI --> AuthSvc
    UserAPI --> RbacSvc
    LeagueAPI --> LeagueSvc
    ClubAPI --> ClubSvc
    CommunityAPI --> CommunitySvc
    WorkAPI --> WorkSvc
    FileAPI --> FileSvc

    AuthSvc --> UserRepo
    RbacSvc --> UserRepo
    LeagueSvc --> LeagueRepo
    ClubSvc --> ClubRepo
    CommunitySvc --> CommunityRepo
    WorkSvc --> WorkRepo
    FileSvc --> FileRepo

    UserRepo --> Users
    UserRepo --> Roles
    LeagueRepo --> LeagueTables
    ClubRepo --> ClubTables
    CommunityRepo --> CommunityTables
    WorkRepo --> WorkTables
    FileRepo --> Files

    LeagueSvc --> RbacSvc
    ClubSvc --> RbacSvc
    CommunitySvc --> RbacSvc
    WorkSvc --> RbacSvc
    FileSvc --> RbacSvc
```

---

## 4. 部署架构图

```mermaid
flowchart TB
    subgraph UserNet[校园网 / 互联网访问]
        Teacher[教师/管理员 PC]
        Student[学生手机/PC]
    end

    subgraph Server[校内服务器 / 云主机]
        subgraph Docker[Docker Compose]
            Nginx[Nginx 容器\n静态资源 + 反向代理]
            API[Go API 容器\nGin REST 服务]
            MinIO[MinIO 容器\n对象存储]
        end

        subgraph Volumes[持久化卷]
            WebDist[前端 dist 静态文件]
            DBFile[SQLite 数据库文件]
            ObjectData[MinIO 对象数据 / 本地上传目录]
            Logs[应用日志目录]
            BackupDir[备份目录]
        end

        BackupJob[定时备份任务\n每日全量备份]
        Health[健康检查\n/health]
    end

    subgraph Future[后续演进]
        PostgreSQL[(PostgreSQL)]
        OSS[云对象存储 / S3]
        LB[负载均衡 / 网关]
    end

    Teacher -->|HTTPS| Nginx
    Student -->|HTTPS| Nginx
    Nginx -->|静态资源| WebDist
    Nginx -->|/api/v1| API
    API --> DBFile
    API --> MinIO
    MinIO --> ObjectData
    API --> Logs
    BackupJob --> DBFile
    BackupJob --> ObjectData
    BackupJob --> BackupDir
    API --> Health

    DBFile -.数据规模/并发增长后迁移.-> PostgreSQL
    MinIO -.存储规模增长后迁移.-> OSS
    Nginx -.多实例部署后接入.-> LB
```

---

## 5. 数据流图

```mermaid
flowchart LR
    subgraph Actor[用户角色]
        Student[普通学生]
        ClubLeader[社团负责人]
        Teacher[社团指导老师]
        LeagueAdmin[团委管理员]
        WorkAdmin[学工处管理员]
        CommunityAdmin[社区管理员]
        SysAdmin[系统管理员]
    end

    subgraph FE[Vue3 前端]
        LoginPage[登录页]
        BusinessPage[业务页面]
        UploadPage[上传组件]
        PermissionView[权限菜单/按钮]
    end

    subgraph API[Go REST API]
        AuthFlow[认证流程]
        AuthzFlow[鉴权与数据权限过滤]
        BizFlow[业务处理流程]
        FileFlow[文件处理流程]
        LogFlow[日志记录]
    end

    subgraph Data[数据存储]
        UserData[(用户/角色/权限)]
        BizData[(业务数据)]
        FileMeta[(文件元数据)]
        ObjectStore[MinIO / 本地文件]
        AuditLog[结构化日志]
    end

    Student --> LoginPage
    ClubLeader --> LoginPage
    Teacher --> LoginPage
    LeagueAdmin --> LoginPage
    WorkAdmin --> LoginPage
    CommunityAdmin --> LoginPage
    SysAdmin --> LoginPage

    LoginPage -->|账号/密码| AuthFlow
    AuthFlow -->|bcrypt校验| UserData
    AuthFlow -->|JWT + 用户信息 + 角色权限| LoginPage
    LoginPage --> PermissionView

    PermissionView --> BusinessPage
    BusinessPage -->|Bearer JWT + 请求参数| AuthzFlow
    AuthzFlow -->|校验角色权限| UserData
    AuthzFlow -->|按本人/组织/角色过滤| BizFlow
    BizFlow -->|状态流转/业务规则/事务| BizData
    BizData -->|结果数据| BizFlow
    BizFlow -->|统一响应 JSON| BusinessPage

    UploadPage -->|文件 + Bearer JWT| FileFlow
    FileFlow -->|类型/大小/MIME校验| ObjectStore
    FileFlow -->|保存文件元数据| FileMeta
    FileFlow -->|文件ID/URL| UploadPage

    AuthFlow --> LogFlow
    AuthzFlow --> LogFlow
    BizFlow --> LogFlow
    FileFlow --> LogFlow
    LogFlow --> AuditLog
```

### 5.1 典型业务数据流：社团活动报名与签到

```mermaid
sequenceDiagram
    autonumber
    participant S as 普通学生
    participant F as Vue3前端
    participant A as Go API
    participant R as RBAC/数据权限
    participant D as SQLite
    participant L as 结构化日志

    S->>F: 登录并进入活动列表
    F->>A: GET /api/v1/club/activities + JWT
    A->>R: 校验 club_activity:read
    R-->>A: 允许访问
    A->>D: 查询报名中活动
    D-->>A: 活动列表
    A-->>F: 返回活动列表

    S->>F: 点击报名
    F->>A: POST /club/activities/{id}/registrations + JWT
    A->>R: 校验 activity_registration:create 和本人身份
    R-->>A: 允许访问
    A->>D: 校验人数上限、重复报名、活动状态
    A->>D: 写入 activity_registrations
    A->>L: 记录报名操作日志
    A-->>F: 报名成功

    S->>F: 扫描签到二维码
    F->>A: POST /club/activities/{id}/checkins + 签到令牌 + JWT
    A->>R: 校验 club_checkin:checkin 和本人身份
    R-->>A: 允许访问
    A->>D: 校验已报名、未签到、二维码有效期与签名
    A->>D: 写入 activity_checkins
    A->>L: 记录签到操作日志
    A-->>F: 签到成功
```

---

## 6. 权限架构图

```mermaid
flowchart TB
    subgraph Identity[身份认证]
        Login[账号密码登录]
        Password[bcrypt 密码哈希校验]
        JWT[JWT 签发\nuser_id / account / token_exp]
        Token[Authorization: Bearer Token]
    end

    subgraph RBAC[RBAC 权限模型]
        User[User 用户]
        UserRole[UserRole 用户角色]
        Role[Role 角色]
        RolePermission[RolePermission 角色权限]
        Permission[Permission 权限\n资源:动作]
    end

    subgraph Roles[业务角色]
        SYS[SYS_ADMIN\n系统管理员]
        LEAGUE[LEAGUE_ADMIN\n团委管理员]
        AFFAIRS[STUDENT_AFFAIRS_ADMIN\n学工处管理员]
        COMMUNITY[COMMUNITY_ADMIN\n社区管理员]
        TEACHER[CLUB_TEACHER\n社团指导老师]
        CLUB[CLUB_LEADER\n社团负责人]
        C_LEADER[COMMUNITY_LEADER\n自治队伍负责人]
        STUDENT[STUDENT\n普通学生]
    end

    subgraph Enforcement[权限执行点]
        FrontMenu[前端菜单/按钮控制\n体验层控制]
        AuthMiddleware[后端 JWT 中间件\n认证必选]
        RbacMiddleware[后端 RBAC 中间件\n接口权限校验]
        DataScope[Service 数据权限\n本人/本社团/本队伍/业务范围]
        BizRule[业务规则校验\n状态/容量/重复/时间]
    end

    subgraph Resources[资源与动作]
        UserPerm[user:create/read/update/delete]
        LeaguePerm[league_application:create/read/update/approve]
        ClubPerm[club_activity:create/read/update/approve]
        CheckinPerm[club_checkin:create/read/checkin]
        CommunityPerm[community_team / recruitment / duty_schedule]
        WorkPerm[work_position / work_application / work_attendance]
        FilePerm[file:upload/read/delete]
    end

    Login --> Password
    Password --> JWT
    JWT --> Token
    Token --> AuthMiddleware

    User --> UserRole
    UserRole --> Role
    Role --> RolePermission
    RolePermission --> Permission

    SYS --> Role
    LEAGUE --> Role
    AFFAIRS --> Role
    COMMUNITY --> Role
    TEACHER --> Role
    CLUB --> Role
    C_LEADER --> Role
    STUDENT --> Role

    Permission --> UserPerm
    Permission --> LeaguePerm
    Permission --> ClubPerm
    Permission --> CheckinPerm
    Permission --> CommunityPerm
    Permission --> WorkPerm
    Permission --> FilePerm

    Permission --> FrontMenu
    AuthMiddleware --> RbacMiddleware
    RbacMiddleware --> DataScope
    DataScope --> BizRule
    BizRule --> Resources
```

### 6.1 权限控制边界

```mermaid
flowchart LR
    Request[API 请求] --> Auth{JWT 是否有效?}
    Auth -- 否 --> E401[401 未认证]
    Auth -- 是 --> Perm{是否拥有接口权限?}
    Perm -- 否 --> E403[403 权限不足]
    Perm -- 是 --> Scope{是否满足数据范围?}
    Scope -- 否 --> E403Scope[403 越权访问]
    Scope -- 是 --> Rule{是否满足业务规则?}
    Rule -- 否 --> E422[422/409 业务冲突]
    Rule -- 是 --> Execute[执行业务操作]
    Execute --> Log[记录结构化日志]
    Log --> Response[返回统一响应]
```

---

## 7. 关键架构决策映射

| 决策项 | 当前方案 | 架构影响 |
|--------|----------|----------|
| 后端框架 | Go + Gin | 轻量 REST API、单二进制部署、适合模块化单体 |
| 数据库 | SQLite + GORM | MVP 零配置部署，Repository 层预留 PostgreSQL 迁移 |
| 前端 | Vue3 + Element Plus | 快速构建中后台页面，学生端通过响应式适配 |
| 认证 | JWT Bearer Token | 无状态认证，适合前后端分离和未来多端复用 |
| 总体架构 | 前后端分离 | 前端静态资源与后端 API 独立构建、独立部署 |
| 文件存储 | MinIO / 本地降级 | S3 兼容接口，后续可迁移云对象存储 |
| 权限模型 | RBAC + 数据权限 | 角色权限控制接口访问，服务端过滤本人/组织范围数据 |

---

## 8. 风险与演进方向

| 风险 | 影响 | 当前缓解措施 | 演进方向 |
|------|------|--------------|----------|
| SQLite 写并发瓶颈 | 报名、签到高峰可能出现锁竞争 | 控制事务范围、建立索引、避免长事务 | 迁移 PostgreSQL |
| JWT 主动失效较弱 | 用户禁用或权限变更无法立即完全生效 | 缩短有效期，后端实时查询关键权限 | 引入刷新令牌、Token 黑名单或 token_version |
| Element Plus 移动端体验有限 | 学生手机端表单和签到体验受限 | 响应式布局、简化移动端页面 | 引入移动端组件或小程序端 |
| 文件上传安全 | 恶意文件、越权访问、URL 泄露 | 白名单、随机对象Key、受控URL | 私有桶、临时签名URL、内容检测 |
| 审批规则变化 | 后续多级审批需求增加 | Service 层封装状态流转 | 抽象审批流引擎 |

---

## 9. 架构验收要点

1. 前端通过 `/api/v1` 调用后端 REST API，非公开接口必须携带 JWT。
2. 后端所有接口必须经过认证中间件、RBAC 校验和数据权限过滤。
3. 核心业务模块按 Handler / Service / Repository 分层实现。
4. 所有列表接口分页，关键外键、状态、时间字段建立索引。
5. 文件上传必须校验类型、大小、MIME，并保存文件元数据。
6. 关键操作必须记录结构化日志，日志不得包含密码、JWT 完整值和敏感配置。
7. SQLite 数据库文件、上传文件目录必须纳入每日备份。
8. 部署环境提供 `/health` 健康检查接口。
