package authz

import (
	"strings"

	"gorm.io/gorm"

	"unistep-platform/backend/internal/models"
)

// ScopeWeight 数据作用域宽窄权重，取最宽
var ScopeWeight = map[string]int{
	"self":    1,
	"team":    2,
	"college": 3,
	"all":     4,
}

// UserAuthz 表示用户聚合后的授权信息
type UserAuthz struct {
	Roles       []string   `json:"roles"`       // 角色编码列表
	Permissions []string   `json:"permissions"` // 权限码（去重）
	Menus       []MenuItem `json:"menus"`       // 可见菜单（按矩阵过滤后的有序列表）
	DataScope   string     `json:"dataScope"`   // 数据作用域（取最宽）
}

// Resolve 根据用户当前的 roles 字符串聚合权限信息
// 优先从数据库 roles / role_permissions 中加载；若数据库未配置（兼容旧逻辑），
// 则回退到 BuiltinRoles + RoleMenuMatrix 的内存定义
func Resolve(db *gorm.DB, userRoles string) UserAuthz {
	roleCodes := splitRoles(userRoles)
	authz := UserAuthz{Roles: roleCodes}

	if len(roleCodes) == 0 {
		return authz
	}

	// 1. 解析数据作用域：从 roles 表读取后取最宽
	authz.DataScope = resolveDataScope(db, roleCodes)

	// 2. 解析权限码（去重）
	permSet := map[string]struct{}{}
	if db != nil {
		var roleIDs []uint
		db.Model(&models.Role{}).Where("code IN ?", roleCodes).Pluck("id", &roleIDs)
		if len(roleIDs) > 0 {
			var codes []string
			db.Table("role_permissions AS rp").
				Select("p.code").
				Joins("JOIN permissions p ON p.id = rp.permission_id").
				Where("rp.role_id IN ?", roleIDs).
				Scan(&codes)
			for _, c := range codes {
				permSet[c] = struct{}{}
			}
		}
	}
	// 兜底：若数据库尚未有权限数据，则按角色矩阵注入菜单权限码
	if len(permSet) == 0 {
		for _, code := range roleCodes {
			for _, menuKey := range RoleMenuMatrix[code] {
				permSet[MenuPermissionCode(menuKey)] = struct{}{}
			}
		}
	}
	authz.Permissions = mapKeys(permSet)

	// 3. 按权限码过滤菜单，保持 BuiltinMenus 顺序
	authz.Menus = filterMenus(permSet)
	return authz
}

func resolveDataScope(db *gorm.DB, codes []string) string {
	best := ""
	if db != nil {
		var scopes []string
		db.Model(&models.Role{}).Where("code IN ?", codes).Pluck("data_scope", &scopes)
		for _, s := range scopes {
			if ScopeWeight[s] > ScopeWeight[best] {
				best = s
			}
		}
	}
	if best != "" {
		return best
	}
	// 兜底：根据内置角色推断
	for _, code := range codes {
		for _, r := range BuiltinRoles {
			if r.Code == code && ScopeWeight[r.DataScope] > ScopeWeight[best] {
				best = r.DataScope
			}
		}
	}
	if best == "" {
		best = "self"
	}
	return best
}

func filterMenus(permSet map[string]struct{}) []MenuItem {
	out := make([]MenuItem, 0, len(BuiltinMenus))
	for _, m := range BuiltinMenus {
		if _, ok := permSet[MenuPermissionCode(m.Key)]; ok {
			out = append(out, m)
		}
	}
	return out
}

func splitRoles(s string) []string {
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	seen := map[string]struct{}{}
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		if _, ok := seen[p]; ok {
			continue
		}
		seen[p] = struct{}{}
		out = append(out, p)
	}
	return out
}

func mapKeys(m map[string]struct{}) []string {
	out := make([]string, 0, len(m))
	for k := range m {
		out = append(out, k)
	}
	return out
}
