package database

import (
	"os"
	"path/filepath"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"unistep-platform/backend/internal/authz"
	"unistep-platform/backend/internal/models"
)

func Connect(path string) (*gorm.DB, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return nil, err
	}

	db, err := gorm.Open(sqlite.Open(path), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := configureSQLite(db); err != nil {
		return nil, err
	}

	// Sprint3：新增团员发展模块的多张表
	// Sprint4：新增社团活动模块的多张表
	// Sprint5：新增学生社区与自治队伍模块的多张表
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
		// Sprint6 / RBAC：角色、权限、角色-权限关联表
		&models.Role{},
		&models.Permission{},
		&models.RolePermission{},
	); err != nil {
		return nil, err
	}

	return db, nil
}

// Seed 向数据库插入测试账号，按用户名逐个插入（已存在则跳过）
func Seed(db *gorm.DB) error {
	// ---- RBAC 种子：内置角色 ----
	if err := seedRoles(db); err != nil {
		return err
	}

	// ---- RBAC 种子：菜单权限 + API 权限 ----
	if err := seedPermissions(db); err != nil {
		return err
	}

	// ---- RBAC 种子：角色-权限关联 ----
	if err := seedRolePermissions(db); err != nil {
		return err
	}

	// ---- 测试用户 ----
	testAccounts := []struct {
		Username   string
		Password   string
		Email      string
		Roles      string
		RealName   string
		College    string
		ClassName  string
		Status     string
	}{
		{"admin", "admin123", "admin@unistep.edu.cn", "admin", "系统管理员", "学工部", "", "active"},
		{"teacher_wang", "teacher123", "wang@unistep.edu.cn", "teacher", "王老师", "计算机学院", "", "active"},
		{"student_li", "student123", "li@unistep.edu.cn", "student_cadre", "李同学", "计算机学院", "计科2101", "active"},
		{"student_zhang", "student123", "zhang@unistep.edu.cn", "student", "张同学", "计算机学院", "计科2101", "active"},
		{"student_liu", "student123", "liu@unistep.edu.cn", "student", "刘同学", "计算机学院", "计科2102", "active"},
	}

	for _, a := range testAccounts {
		var count int64
		db.Model(&models.User{}).Where("username = ?", a.Username).Count(&count)
		if count > 0 {
			continue
		}

		hash, err := bcrypt.GenerateFromPassword([]byte(a.Password), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user := models.User{
			Username:  a.Username,
			Password:  string(hash),
			Email:     a.Email,
			Roles:     a.Roles,
			RealName:  a.RealName,
			College:   a.College,
			ClassName: a.ClassName,
			Status:    a.Status,
		}
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	return nil
}

// seedRoles 播种内置角色
func seedRoles(db *gorm.DB) error {
	for _, r := range authz.BuiltinRoles {
		var count int64
		db.Model(&models.Role{}).Where("code = ?", r.Code).Count(&count)
		if count > 0 {
			continue
		}
		if err := db.Create(&models.Role{
			Code:        r.Code,
			Name:        r.Name,
			DataScope:   r.DataScope,
			Description: r.Description,
			Builtin:     true,
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedPermissions 播种所有菜单权限码
func seedPermissions(db *gorm.DB) error {
	for _, m := range authz.BuiltinMenus {
		code := authz.MenuPermissionCode(m.Key)
		var count int64
		db.Model(&models.Permission{}).Where("code = ?", code).Count(&count)
		if count > 0 {
			continue
		}
		module := "menu"
		if m.Key == "dashboard" || m.Key == "services" {
			module = m.Key
		}
		if err := db.Create(&models.Permission{
			Code:   code,
			Name:   "菜单:" + m.Title,
			Module: module,
			Type:   "menu",
		}).Error; err != nil {
			return err
		}
	}
	return nil
}

// seedRolePermissions 播种角色-菜单关联
func seedRolePermissions(db *gorm.DB) error {
	for roleCode, menuKeys := range authz.RoleMenuMatrix {
		var role models.Role
		if err := db.Where("code = ?", roleCode).First(&role).Error; err != nil {
			continue // 角色不存在，跳过
		}
		for _, menuKey := range menuKeys {
			code := authz.MenuPermissionCode(menuKey)
			var perm models.Permission
			if err := db.Where("code = ?", code).First(&perm).Error; err != nil {
				continue
			}
			// 检查关联是否已存在
			var count int64
			db.Model(&models.RolePermission{}).
				Where("role_id = ? AND permission_id = ?", role.ID, perm.ID).
				Count(&count)
			if count > 0 {
				continue
			}
			if err := db.Create(&models.RolePermission{
				RoleID:       role.ID,
				PermissionID: perm.ID,
			}).Error; err != nil {
				return err
			}
		}
	}
	return nil
}

func configureSQLite(db *gorm.DB) error {
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}

	pragmas := []string{
		"PRAGMA foreign_keys = ON;",
		"PRAGMA journal_mode = WAL;",
		"PRAGMA busy_timeout = 5000;",
	}

	for _, pragma := range pragmas {
		if _, err := sqlDB.Exec(pragma); err != nil {
			return err
		}
	}

	return nil
}
