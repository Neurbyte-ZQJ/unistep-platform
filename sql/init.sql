PRAGMA foreign_keys = ON;
PRAGMA journal_mode = WAL;
PRAGMA busy_timeout = 5000;

CREATE TABLE IF NOT EXISTS users (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    account TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    password_hash TEXT NOT NULL,
    user_type TEXT NOT NULL CHECK (user_type IN ('student', 'staff', 'admin')),
    department TEXT,
    grade TEXT,
    phone TEXT,
    email TEXT,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'locked', 'disabled')),
    failed_login_count INTEGER NOT NULL DEFAULT 0 CHECK (failed_login_count >= 0),
    locked_until DATETIME,
    last_login_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME
);

CREATE TABLE IF NOT EXISTS roles (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    description TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS permissions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    code TEXT NOT NULL UNIQUE,
    name TEXT NOT NULL,
    resource TEXT NOT NULL,
    action TEXT NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_roles (
    user_id INTEGER NOT NULL,
    role_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (user_id, role_id),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS role_permissions (
    role_id INTEGER NOT NULL,
    permission_id INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY (role_id, permission_id),
    FOREIGN KEY (role_id) REFERENCES roles(id) ON DELETE CASCADE,
    FOREIGN KEY (permission_id) REFERENCES permissions(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    original_name TEXT NOT NULL,
    object_key TEXT NOT NULL UNIQUE,
    bucket TEXT NOT NULL,
    content_type TEXT NOT NULL,
    size INTEGER NOT NULL CHECK (size >= 0),
    url TEXT NOT NULL,
    biz_type TEXT NOT NULL,
    uploaded_by INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (uploaded_by) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS league_applications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL,
    reason TEXT NOT NULL,
    resume TEXT,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending', 'approved', 'rejected')),
    approval_opinion TEXT,
    approved_by INTEGER,
    approved_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS activists (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    student_id INTEGER NOT NULL UNIQUE,
    application_id INTEGER NOT NULL UNIQUE,
    status TEXT NOT NULL DEFAULT 'cultivating' CHECK (status IN ('cultivating', 'developed', 'eliminated')),
    registered_date DATE NOT NULL DEFAULT (date('now')),
    status_reason TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (application_id) REFERENCES league_applications(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS cultivation_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    activist_id INTEGER NOT NULL,
    record_type TEXT NOT NULL CHECK (record_type IN ('thought_report', 'training', 'practice', 'review')),
    content TEXT NOT NULL,
    file_id INTEGER,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'confirmed')),
    created_by INTEGER NOT NULL,
    confirmed_by INTEGER,
    confirmed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (activist_id) REFERENCES activists(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE SET NULL,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (confirmed_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS club_activities (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    club_name TEXT NOT NULL,
    title TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    location TEXT NOT NULL,
    capacity INTEGER NOT NULL CHECK (capacity > 0),
    description TEXT NOT NULL,
    budget NUMERIC(10,2) CHECK (budget IS NULL OR budget >= 0),
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'pending', 'rejected', 'reg_open', 'reg_closed', 'in_progress', 'completed', 'archived')),
    approval_opinion TEXT,
    created_by INTEGER NOT NULL,
    approved_by INTEGER,
    approved_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    CHECK (end_time > start_time),
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (approved_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS activity_registrations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    activity_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,
    status TEXT NOT NULL DEFAULT 'registered' CHECK (status IN ('registered', 'cancelled')),
    registered_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    cancelled_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (activity_id, student_id),
    FOREIGN KEY (activity_id) REFERENCES club_activities(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS activity_checkins (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    activity_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,
    checkin_time DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    checkin_method TEXT NOT NULL DEFAULT 'qr' CHECK (checkin_method IN ('qr', 'manual')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (activity_id, student_id),
    FOREIGN KEY (activity_id) REFERENCES club_activities(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS activity_files (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    activity_id INTEGER NOT NULL,
    file_id INTEGER NOT NULL,
    file_type TEXT NOT NULL DEFAULT 'image' CHECK (file_type IN ('image', 'document', 'summary')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (activity_id, file_id),
    FOREIGN KEY (activity_id) REFERENCES club_activities(id) ON DELETE CASCADE,
    FOREIGN KEY (file_id) REFERENCES files(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS community_teams (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    area TEXT,
    description TEXT,
    leader_id INTEGER,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'inactive')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (leader_id) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS community_members (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    role TEXT NOT NULL DEFAULT 'member' CHECK (role IN ('leader', 'member')),
    joined_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'left')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (team_id, user_id),
    FOREIGN KEY (team_id) REFERENCES community_teams(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS recruitment_notices (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id INTEGER NOT NULL,
    title TEXT NOT NULL,
    requirements TEXT,
    quota INTEGER NOT NULL CHECK (quota > 0),
    deadline DATETIME NOT NULL,
    status TEXT NOT NULL DEFAULT 'draft' CHECK (status IN ('draft', 'published', 'closed')),
    created_by INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (team_id) REFERENCES community_teams(id) ON DELETE CASCADE,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS recruitment_applications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    notice_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,
    statement TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected')),
    review_opinion TEXT,
    reviewed_by INTEGER,
    reviewed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (notice_id, student_id),
    FOREIGN KEY (notice_id) REFERENCES recruitment_notices(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS duty_schedules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    duty_date DATE NOT NULL,
    time_slot TEXT NOT NULL,
    start_time DATETIME NOT NULL,
    end_time DATETIME NOT NULL,
    status TEXT NOT NULL DEFAULT 'scheduled' CHECK (status IN ('scheduled', 'completed', 'absent', 'cancelled')),
    created_by INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (user_id, duty_date, time_slot),
    CHECK (end_time > start_time),
    FOREIGN KEY (team_id) REFERENCES community_teams(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS duty_attendances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    schedule_id INTEGER NOT NULL UNIQUE,
    user_id INTEGER NOT NULL,
    attendance_status TEXT NOT NULL CHECK (attendance_status IN ('present', 'absent', 'leave')),
    hours NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (hours >= 0),
    remark TEXT,
    confirmed_by INTEGER,
    confirmed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (schedule_id) REFERENCES duty_schedules(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (confirmed_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS work_positions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL,
    department TEXT NOT NULL,
    quota INTEGER NOT NULL CHECK (quota > 0),
    salary NUMERIC(10,2) NOT NULL CHECK (salary >= 0),
    salary_unit TEXT NOT NULL DEFAULT 'hour' CHECK (salary_unit IN ('hour', 'month', 'task')),
    requirements TEXT,
    work_time TEXT,
    deadline DATETIME,
    status TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'closed', 'archived')),
    created_by INTEGER NOT NULL,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at DATETIME,
    FOREIGN KEY (created_by) REFERENCES users(id) ON DELETE RESTRICT
);

CREATE TABLE IF NOT EXISTS work_applications (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    position_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,
    statement TEXT,
    status TEXT NOT NULL DEFAULT 'pending' CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    review_opinion TEXT,
    reviewed_by INTEGER,
    reviewed_at DATETIME,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (position_id, student_id),
    FOREIGN KEY (position_id) REFERENCES work_positions(id) ON DELETE CASCADE,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (reviewed_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS work_assignments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    position_id INTEGER NOT NULL,
    student_id INTEGER NOT NULL,
    application_id INTEGER,
    start_date DATE NOT NULL DEFAULT (date('now')),
    end_date DATE,
    status TEXT NOT NULL DEFAULT 'active' CHECK (status IN ('active', 'ended')),
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (position_id) REFERENCES work_positions(id) ON DELETE RESTRICT,
    FOREIGN KEY (student_id) REFERENCES users(id) ON DELETE RESTRICT,
    FOREIGN KEY (application_id) REFERENCES work_applications(id) ON DELETE SET NULL
);

CREATE TABLE IF NOT EXISTS work_attendances (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    assignment_id INTEGER NOT NULL,
    work_date DATE NOT NULL,
    clock_in_at DATETIME,
    clock_out_at DATETIME,
    hours NUMERIC(5,2) NOT NULL DEFAULT 0 CHECK (hours >= 0),
    status TEXT NOT NULL DEFAULT 'normal' CHECK (status IN ('normal', 'exception', 'adjusted')),
    exception_type TEXT CHECK (exception_type IS NULL OR exception_type IN ('late', 'early_leave', 'missing_clock', 'other')),
    adjusted_by INTEGER,
    remark TEXT,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (assignment_id, work_date),
    CHECK (clock_out_at IS NULL OR clock_in_at IS NULL OR clock_out_at > clock_in_at),
    FOREIGN KEY (assignment_id) REFERENCES work_assignments(id) ON DELETE CASCADE,
    FOREIGN KEY (adjusted_by) REFERENCES users(id) ON DELETE SET NULL
);

CREATE INDEX IF NOT EXISTS idx_users_user_type ON users(user_type);
CREATE INDEX IF NOT EXISTS idx_users_department_grade ON users(department, grade);
CREATE INDEX IF NOT EXISTS idx_users_status ON users(status);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

CREATE INDEX IF NOT EXISTS idx_permissions_resource_action ON permissions(resource, action);
CREATE INDEX IF NOT EXISTS idx_files_biz_type ON files(biz_type);
CREATE INDEX IF NOT EXISTS idx_files_uploaded_by ON files(uploaded_by);
CREATE INDEX IF NOT EXISTS idx_files_created_at ON files(created_at);
CREATE INDEX IF NOT EXISTS idx_files_deleted_at ON files(deleted_at);

CREATE INDEX IF NOT EXISTS idx_league_applications_student ON league_applications(student_id);
CREATE INDEX IF NOT EXISTS idx_league_applications_status_created ON league_applications(status, created_at);
CREATE INDEX IF NOT EXISTS idx_league_applications_approved_by ON league_applications(approved_by);
CREATE INDEX IF NOT EXISTS idx_league_applications_deleted_at ON league_applications(deleted_at);
CREATE UNIQUE INDEX IF NOT EXISTS ux_league_applications_active_student ON league_applications(student_id) WHERE deleted_at IS NULL AND status IN ('draft', 'pending', 'approved');

CREATE INDEX IF NOT EXISTS idx_activists_status ON activists(status);
CREATE INDEX IF NOT EXISTS idx_activists_deleted_at ON activists(deleted_at);
CREATE INDEX IF NOT EXISTS idx_cultivation_records_activist ON cultivation_records(activist_id);
CREATE INDEX IF NOT EXISTS idx_cultivation_records_type_status ON cultivation_records(record_type, status);
CREATE INDEX IF NOT EXISTS idx_cultivation_records_created_by ON cultivation_records(created_by);

CREATE INDEX IF NOT EXISTS idx_club_activities_status_start ON club_activities(status, start_time);
CREATE INDEX IF NOT EXISTS idx_club_activities_club_name ON club_activities(club_name);
CREATE INDEX IF NOT EXISTS idx_club_activities_created_by ON club_activities(created_by);
CREATE INDEX IF NOT EXISTS idx_club_activities_approved_by ON club_activities(approved_by);
CREATE INDEX IF NOT EXISTS idx_club_activities_deleted_at ON club_activities(deleted_at);
CREATE INDEX IF NOT EXISTS idx_activity_registrations_activity_status ON activity_registrations(activity_id, status);
CREATE INDEX IF NOT EXISTS idx_activity_registrations_student_status ON activity_registrations(student_id, status);
CREATE INDEX IF NOT EXISTS idx_activity_checkins_activity_time ON activity_checkins(activity_id, checkin_time);
CREATE INDEX IF NOT EXISTS idx_activity_checkins_student ON activity_checkins(student_id);
CREATE INDEX IF NOT EXISTS idx_activity_files_activity ON activity_files(activity_id);
CREATE INDEX IF NOT EXISTS idx_activity_files_file ON activity_files(file_id);

CREATE INDEX IF NOT EXISTS idx_community_teams_area_status ON community_teams(area, status);
CREATE INDEX IF NOT EXISTS idx_community_teams_leader ON community_teams(leader_id);
CREATE INDEX IF NOT EXISTS idx_community_teams_deleted_at ON community_teams(deleted_at);
CREATE INDEX IF NOT EXISTS idx_community_members_user_status ON community_members(user_id, status);
CREATE INDEX IF NOT EXISTS idx_community_members_team_status ON community_members(team_id, status);
CREATE INDEX IF NOT EXISTS idx_recruitment_notices_team_status_deadline ON recruitment_notices(team_id, status, deadline);
CREATE INDEX IF NOT EXISTS idx_recruitment_notices_deleted_at ON recruitment_notices(deleted_at);
CREATE INDEX IF NOT EXISTS idx_recruitment_applications_student_status ON recruitment_applications(student_id, status);
CREATE INDEX IF NOT EXISTS idx_recruitment_applications_notice_status ON recruitment_applications(notice_id, status);
CREATE INDEX IF NOT EXISTS idx_duty_schedules_team_date ON duty_schedules(team_id, duty_date);
CREATE INDEX IF NOT EXISTS idx_duty_schedules_user_date ON duty_schedules(user_id, duty_date);
CREATE INDEX IF NOT EXISTS idx_duty_schedules_status ON duty_schedules(status);
CREATE INDEX IF NOT EXISTS idx_duty_attendances_user_status ON duty_attendances(user_id, attendance_status);

CREATE INDEX IF NOT EXISTS idx_work_positions_department_status ON work_positions(department, status);
CREATE INDEX IF NOT EXISTS idx_work_positions_deadline ON work_positions(deadline);
CREATE INDEX IF NOT EXISTS idx_work_positions_created_by ON work_positions(created_by);
CREATE INDEX IF NOT EXISTS idx_work_positions_deleted_at ON work_positions(deleted_at);
CREATE INDEX IF NOT EXISTS idx_work_applications_position_status ON work_applications(position_id, status);
CREATE INDEX IF NOT EXISTS idx_work_applications_student_status ON work_applications(student_id, status);
CREATE INDEX IF NOT EXISTS idx_work_assignments_position_status ON work_assignments(position_id, status);
CREATE INDEX IF NOT EXISTS idx_work_assignments_student_status ON work_assignments(student_id, status);
CREATE UNIQUE INDEX IF NOT EXISTS ux_work_assignments_active_student ON work_assignments(student_id) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_work_attendances_assignment_date ON work_attendances(assignment_id, work_date);
CREATE INDEX IF NOT EXISTS idx_work_attendances_work_date ON work_attendances(work_date);
CREATE INDEX IF NOT EXISTS idx_work_attendances_status ON work_attendances(status);

INSERT OR IGNORE INTO roles (code, name, description) VALUES
('SYS_ADMIN', '系统管理员', '负责用户、角色、权限、系统配置与数据备份'),
('LEAGUE_ADMIN', '团委管理员', '负责入团申请审批、积极分子与培养记录管理'),
('STUDENT_AFFAIRS_ADMIN', '学工处管理员', '负责勤工岗位、报名录用与考勤管理'),
('COMMUNITY_ADMIN', '社区管理员', '负责自治队伍、排班与志愿服务管理'),
('CLUB_TEACHER', '社团指导老师', '负责社团活动审批与活动数据查看'),
('CLUB_LEADER', '社团负责人', '负责活动策划、报名管理、签到与活动材料上传'),
('COMMUNITY_LEADER', '自治队伍负责人', '负责纳新公告、报名审核和队伍值班管理'),
('STUDENT', '普通学生', '办理入团申请、活动报名签到、自治队伍报名、勤工报名与打卡');

INSERT OR IGNORE INTO permissions (code, name, resource, action) VALUES
('user:create', '创建用户', 'user', 'create'),
('user:read', '查看用户', 'user', 'read'),
('user:update', '更新用户', 'user', 'update'),
('user:delete', '删除用户', 'user', 'delete'),
('role:manage', '管理角色权限', 'role', 'manage'),
('league_application:create', '提交入团申请', 'league_application', 'create'),
('league_application:read', '查看入团申请', 'league_application', 'read'),
('league_application:update', '更新入团申请', 'league_application', 'update'),
('league_application:approve', '审批入团申请', 'league_application', 'approve'),
('activist:read', '查看积极分子', 'activist', 'read'),
('activist:update', '更新积极分子', 'activist', 'update'),
('cultivation_record:create', '创建培养记录', 'cultivation_record', 'create'),
('cultivation_record:read', '查看培养记录', 'cultivation_record', 'read'),
('cultivation_record:approve', '确认培养记录', 'cultivation_record', 'approve'),
('club_activity:create', '创建社团活动', 'club_activity', 'create'),
('club_activity:read', '查看社团活动', 'club_activity', 'read'),
('club_activity:update', '更新社团活动', 'club_activity', 'update'),
('club_activity:approve', '审批社团活动', 'club_activity', 'approve'),
('activity_registration:create', '活动报名', 'activity_registration', 'create'),
('activity_registration:read', '查看活动报名', 'activity_registration', 'read'),
('activity_registration:delete', '取消活动报名', 'activity_registration', 'delete'),
('club_checkin:create', '生成活动签到码', 'club_checkin', 'create'),
('club_checkin:read', '查看活动签到', 'club_checkin', 'read'),
('club_checkin:checkin', '活动签到', 'club_checkin', 'checkin'),
('community_team:create', '创建自治队伍', 'community_team', 'create'),
('community_team:read', '查看自治队伍', 'community_team', 'read'),
('community_team:update', '更新自治队伍', 'community_team', 'update'),
('community_team:delete', '删除自治队伍', 'community_team', 'delete'),
('community_member:create', '添加队伍成员', 'community_member', 'create'),
('community_member:read', '查看队伍成员', 'community_member', 'read'),
('recruitment:create', '发布纳新公告', 'recruitment', 'create'),
('recruitment:read', '查看纳新公告', 'recruitment', 'read'),
('recruitment_application:create', '提交纳新报名', 'recruitment_application', 'create'),
('recruitment_application:read', '查看纳新报名', 'recruitment_application', 'read'),
('recruitment_application:approve', '审核纳新报名', 'recruitment_application', 'approve'),
('duty_schedule:create', '创建值班排班', 'duty_schedule', 'create'),
('duty_schedule:read', '查看值班排班', 'duty_schedule', 'read'),
('duty_schedule:update', '更新值班排班', 'duty_schedule', 'update'),
('work_position:create', '发布勤工岗位', 'work_position', 'create'),
('work_position:read', '查看勤工岗位', 'work_position', 'read'),
('work_position:update', '更新勤工岗位', 'work_position', 'update'),
('work_application:create', '提交岗位报名', 'work_application', 'create'),
('work_application:read', '查看岗位报名', 'work_application', 'read'),
('work_application:approve', '审核岗位报名', 'work_application', 'approve'),
('work_attendance:checkin', '勤工打卡', 'work_attendance', 'checkin'),
('work_attendance:read', '查看勤工考勤', 'work_attendance', 'read'),
('work_attendance:update', '调整勤工考勤', 'work_attendance', 'update'),
('file:upload', '上传文件', 'file', 'upload'),
('file:read', '查看文件', 'file', 'read'),
('file:delete', '删除文件', 'file', 'delete');
