import { createRouter, createWebHistory, type RouteRecordRaw } from 'vue-router'
import { useAuthStore } from '../stores/auth'

// 角色编码常量
export const ROLES = {
  ADMIN: 'admin',
  TEACHER: 'teacher',
  CADRE: 'student_cadre',
  STUDENT: 'student',
} as const

// 所有登录后可访问的路由都挂在主布局下
// meta.roles 声明可访问的角色编码白名单（不写表示登录即可访问）
const layoutChildren: RouteRecordRaw[] = [
  {
    path: '',
    name: 'Dashboard',
    component: () => import('../views/DashboardView.vue'),
    meta: { title: '工作台', menuKey: 'dashboard' },
  },
  {
    path: 'services',
    name: 'Services',
    component: () => import('../views/ServicesView.vue'),
    meta: { title: '服务入口', menuKey: 'services' },
  },
  // ---- 团员发展 ----
  {
    path: 'members',
    name: 'MemberList',
    component: () => import('../views/members/MemberListView.vue'),
    meta: {
      title: '团员发展',
      menuKey: 'members',
      roles: [ROLES.ADMIN, ROLES.TEACHER, ROLES.CADRE],
    },
  },
  {
    path: 'members/new',
    name: 'MemberCreate',
    component: () => import('../views/members/MemberEditView.vue'),
    meta: { roles: [ROLES.ADMIN, ROLES.TEACHER, ROLES.CADRE] },
  },
  {
    path: 'members/:id/edit',
    name: 'MemberEdit',
    component: () => import('../views/members/MemberEditView.vue'),
    meta: { roles: [ROLES.ADMIN, ROLES.TEACHER, ROLES.CADRE] },
  },
  {
    path: 'members/:id',
    name: 'MemberDetail',
    component: () => import('../views/members/MemberDetailView.vue'),
    meta: { roles: [ROLES.ADMIN, ROLES.TEACHER, ROLES.CADRE] },
  },
  // ---- 社团活动 ----
  {
    path: 'activities',
    name: 'ActivityList',
    component: () => import('../views/activities/ActivityListView.vue'),
    meta: {
      title: '社团活动',
      menuKey: 'activities',
      roles: [ROLES.ADMIN, ROLES.TEACHER, ROLES.CADRE],
    },
  },
  {
    path: 'activities/new',
    name: 'ActivityCreate',
    component: () => import('../views/activities/ActivityEditView.vue'),
    meta: { roles: [ROLES.ADMIN, ROLES.TEACHER, ROLES.CADRE] },
  },
  {
    path: 'activities/:id/edit',
    name: 'ActivityEdit',
    component: () => import('../views/activities/ActivityEditView.vue'),
    meta: { roles: [ROLES.ADMIN, ROLES.TEACHER, ROLES.CADRE] },
  },
  {
    path: 'activities/:id',
    name: 'ActivityDetail',
    component: () => import('../views/activities/ActivityDetailView.vue'),
    meta: { roles: [ROLES.ADMIN, ROLES.TEACHER, ROLES.CADRE] },
  },
  // ---- 社区队伍 ----
  {
    path: 'community',
    name: 'CommunityList',
    component: () => import('../views/community/TeamListView.vue'),
    meta: {
      title: '社区队伍',
      menuKey: 'community',
      roles: [ROLES.ADMIN, ROLES.TEACHER, ROLES.CADRE],
    },
  },
  {
    path: 'community/duty',
    name: 'CommunityDuty',
    component: () => import('../views/community/DutyScheduleView.vue'),
    meta: { title: '值班签到', menuKey: 'community.duty' },
  },
  {
    path: 'community/profile',
    name: 'ServiceProfile',
    component: () => import('../views/community/ServiceProfileView.vue'),
    meta: { title: '服务画像', menuKey: 'community.profile' },
  },
  // ---- 勤工助学 ----
  {
    path: 'workstudy',
    name: 'JobList',
    component: () => import('../views/workstudy/JobListView.vue'),
    meta: { title: '勤工助学', menuKey: 'workstudy' },
  },
  {
    path: 'workstudy/new',
    name: 'JobCreate',
    component: () => import('../views/workstudy/JobEditView.vue'),
    meta: { roles: [ROLES.ADMIN, ROLES.TEACHER] },
  },
  {
    path: 'workstudy/:id/edit',
    name: 'JobEdit',
    component: () => import('../views/workstudy/JobEditView.vue'),
    meta: { roles: [ROLES.ADMIN, ROLES.TEACHER] },
  },
  {
    path: 'workstudy/:id',
    name: 'JobDetail',
    component: () => import('../views/workstudy/JobDetailView.vue'),
  },
  // ---- 管理员后台 ----
  {
    path: 'admin/users',
    name: 'AdminUsers',
    component: () => import('../views/admin/UserListView.vue'),
    meta: { title: '用户管理', menuKey: 'admin.users', roles: [ROLES.ADMIN] },
  },
  {
    path: 'admin/roles',
    name: 'AdminRoles',
    component: () => import('../views/admin/RoleListView.vue'),
    meta: { title: '角色权限', menuKey: 'admin.roles', roles: [ROLES.ADMIN] },
  },
]

const router = createRouter({
  history: createWebHistory(),
  routes: [
    {
      path: '/login',
      name: 'Login',
      component: () => import('../views/LoginView.vue'),
      meta: { public: true },
    },
    {
      path: '/register',
      name: 'Register',
      component: () => import('../views/RegisterView.vue'),
      meta: { public: true },
    },
    {
      path: '/403',
      name: 'Forbidden',
      component: () => import('../views/ForbiddenView.vue'),
      meta: { public: true },
    },
    {
      path: '/',
      component: () => import('../layouts/MainLayout.vue'),
      children: layoutChildren,
    },
  ],
})

// 路由守卫：登录态校验 + 角色校验
router.beforeEach(async (to) => {
  const auth = useAuthStore()

  // 公开页：已登录用户访问登录/注册时回到首页
  if (to.meta.public) {
    if (auth.isAuthenticated && (to.name === 'Login' || to.name === 'Register')) {
      return { name: 'Dashboard' }
    }
    return true
  }

  // 未登录：重定向到登录页
  if (!auth.isAuthenticated) {
    return { name: 'Login', query: { redirect: to.fullPath } }
  }

  // 已登录但本地缺少权限信息（旧 localStorage 升级场景），尝试拉取一次
  // 注意：旧版 user 可能 roles 有值但 menus 为空，所以两个条件都要判断
  if (!auth.user?.roles?.length || !auth.user?.menus?.length) {
    try {
      await auth.fetchProfile()
    } catch {
      // 拉取失败：登出并回登录页
      auth.logout()
      return { name: 'Login', query: { redirect: to.fullPath } }
    }
  }

  // 角色校验
  const allow = (to.meta.roles as string[] | undefined) ?? []
  if (allow.length > 0 && !auth.hasRole(allow)) {
    return { name: 'Forbidden' }
  }

  return true
})

export default router
