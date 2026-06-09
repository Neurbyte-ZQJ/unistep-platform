import { defineStore } from 'pinia'

// 菜单项（与后端 authz.MenuItem 对齐）
export interface MenuItem {
  key: string
  title: string
  path: string
  icon?: string
}

interface UserProfile {
  id: number
  username: string
  email?: string
  realName?: string
  college?: string
  className?: string
  roles: string[] // 角色编码数组，如 ["teacher", "student_cadre"]
  permissions: string[] // 权限码集合
  menus: MenuItem[] // 可见菜单
  dataScope?: string // 数据作用域：all/college/team/self
}

interface ApiResponse<T> {
  code: string
  message: string
  data: T
}

const TOKEN_KEY = 'unistep_token'
const USER_KEY = 'unistep_user'

// 将后端可能返回的 roles 字段（字符串或数组）统一为数组
function normalizeRoles(value: unknown): string[] {
  if (Array.isArray(value)) return value.map((v) => String(v).trim()).filter(Boolean)
  if (typeof value === 'string')
    return value
      .split(',')
      .map((v) => v.trim())
      .filter(Boolean)
  return []
}

function normalizeUser(raw: any): UserProfile {
  return {
    id: raw?.id ?? 0,
    username: raw?.username ?? '',
    email: raw?.email,
    realName: raw?.realName,
    college: raw?.college,
    className: raw?.className,
    roles: normalizeRoles(raw?.roles),
    permissions: Array.isArray(raw?.permissions) ? raw.permissions : [],
    menus: Array.isArray(raw?.menus) ? raw.menus : [],
    dataScope: raw?.dataScope,
  }
}

export const useAuthStore = defineStore('auth', {
  state: () => ({
    token: localStorage.getItem(TOKEN_KEY) || '',
    user: localStorage.getItem(USER_KEY)
      ? normalizeUser(JSON.parse(localStorage.getItem(USER_KEY) as string))
      : (null as UserProfile | null),
  }),
  getters: {
    isAuthenticated: (state) => Boolean(state.token),
    // 角色集合（Set 便于 O(1) 判断）
    roleSet: (state) => new Set(state.user?.roles ?? []),
    // 权限码集合
    permissionSet: (state) => new Set(state.user?.permissions ?? []),
    // 是否拥有任意一个角色
    hasRole(): (role: string | string[]) => boolean {
      return (role: string | string[]) => {
        const required = Array.isArray(role) ? role : [role]
        if (required.length === 0) return true
        return required.some((r) => this.roleSet.has(r))
      }
    },
    // 是否拥有指定权限码
    hasPermission(): (code: string) => boolean {
      return (code: string) => this.permissionSet.has(code)
    },
    // 主要角色（按权限高低排序后返回第一个；用于 Dashboard 视图选择）
    primaryRole(state): string {
      const roles = state.user?.roles ?? []
      const priority = ['admin', 'teacher', 'student_cadre', 'student']
      for (const p of priority) {
        if (roles.includes(p)) return p
      }
      return roles[0] ?? 'student'
    },
  },
  actions: {
    setSession(token: string, user: UserProfile) {
      this.token = token
      this.user = user
      localStorage.setItem(TOKEN_KEY, token)
      localStorage.setItem(USER_KEY, JSON.stringify(user))
    },
    async login(username: string, password: string) {
      const res = await fetch('/api/v1/auth/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      })
      const body = (await res.json()) as ApiResponse<{ token: string; user: any }>
      if (!res.ok) throw new Error(body.message || '登录失败')
      this.setSession(body.data.token, normalizeUser(body.data.user))
    },
    async register(username: string, password: string, email: string) {
      const res = await fetch('/api/v1/auth/register', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password, email }),
      })
      const body = (await res.json()) as ApiResponse<any>
      if (!res.ok) throw new Error(body.message || '注册失败')
    },
    // 拉取最新授权信息（菜单 + 权限），用于刷新或首次加载
    async fetchProfile() {
      if (!this.token) return
      const res = await fetch('/api/v1/users/me', {
        headers: { Authorization: `Bearer ${this.token}` },
      })
      const body = (await res.json()) as ApiResponse<any>
      if (!res.ok) throw new Error(body.message || '获取用户信息失败')
      // 保留 token，刷新 user
      const merged = normalizeUser({ ...body.data, id: body.data.userId ?? body.data.id })
      this.setSession(this.token, merged)
    },
    logout() {
      this.token = ''
      this.user = null
      localStorage.removeItem(TOKEN_KEY)
      localStorage.removeItem(USER_KEY)
    },
  },
})
