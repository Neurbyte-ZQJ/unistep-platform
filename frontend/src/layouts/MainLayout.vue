<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { useRouter, useRoute } from 'vue-router'
import { useAuthStore } from '../stores/auth'
import { Fold, Expand } from '@element-plus/icons-vue'

const router = useRouter()
const route = useRoute()
const auth = useAuthStore()

// 移动端侧边栏控制
const sidebarVisible = ref(false)

// 保底：如果登录了但菜单为空（旧 localStorage 数据），尝试刷新授权信息
onMounted(async () => {
  if (auth.isAuthenticated && (!auth.user?.menus || auth.user.menus.length === 0)) {
    try {
      await auth.fetchProfile()
    } catch {
      // 静默失败，导航守卫会拦
    }
  }
})

// 当前激活菜单：根据路由 path 匹配
const activeMenu = computed(() => {
  // 优先精确匹配菜单 path；否则取路径前缀（如 /members/123 -> /members）
  const menus = auth.user?.menus ?? []
  const path = route.path
  const exact = menus.find((m) => m.path === path)
  if (exact) return exact.path
  // 找最长前缀匹配
  const prefix = menus
    .filter((m) => m.path !== '/' && path.startsWith(m.path))
    .sort((a, b) => b.path.length - a.path.length)[0]
  return prefix?.path ?? '/'
})

// 角色中文标签
const roleLabels: Record<string, string> = {
  admin: '系统管理员',
  teacher: '教师',
  student_cadre: '学生干部',
  student: '普通学生',
}
const roleLabel = computed(() =>
  (auth.user?.roles ?? []).map((r) => roleLabels[r] || r).join(' / ') || '访客',
)

function handleLogout() {
  auth.logout()
  router.push('/login')
}
</script>

<template>
  <el-container class="app-shell">
    <!-- 移动端遮罩 -->
    <div v-if="sidebarVisible" class="sidebar-overlay" @click="sidebarVisible = false" aria-hidden="true" />

    <el-aside width="240px" class="app-aside" :class="{ 'is-mobile-open': sidebarVisible }" role="navigation" aria-label="主导航">
      <div class="brand">一站式服务</div>
      <!-- 菜单根据后端返回的 user.menus 动态渲染 -->
      <el-menu router :default-active="activeMenu" @select="sidebarVisible = false">
        <el-menu-item v-for="item in auth.user?.menus ?? []" :key="item.key" :index="item.path">
          <el-icon v-if="item.icon"><component :is="item.icon" /></el-icon>
          {{ item.title }}
        </el-menu-item>
      </el-menu>
    </el-aside>

    <el-container>
      <el-header class="app-header" role="banner">
        <div class="header-left">
          <el-button
            class="hamburger-btn"
            :icon="sidebarVisible ? Fold : Expand"
            text
            aria-label="打开菜单"
            :aria-expanded="sidebarVisible.toString()"
            @click="sidebarVisible = !sidebarVisible"
          />
          <span>学生"一站式"自主管理过程管理系统</span>
        </div>
        <div class="user-actions">
          <span class="user-name">
            {{ auth.user?.realName || auth.user?.username || '当前用户' }}
            <el-tag size="small" type="info" effect="plain" class="role-tag">{{ roleLabel }}</el-tag>
          </span>
          <el-button type="primary" plain @click="handleLogout">退出登录</el-button>
        </div>
      </el-header>
      <el-main>
        <router-view />
      </el-main>
    </el-container>
  </el-container>
</template>

<style scoped>
.user-actions {
  display: inline-flex;
  align-items: center;
  gap: var(--space-3);
}
.user-name {
  display: inline-flex;
  align-items: center;
}
.role-tag {
  margin-left: var(--space-2);
}
.header-left {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
}
.hamburger-btn {
  display: none;
}
.sidebar-overlay {
  display: none;
}

/* 移动端响应式 */
@media (max-width: 768px) {
  .hamburger-btn {
    display: inline-flex;
  }
  .app-aside {
    position: fixed;
    top: 0;
    left: 0;
    bottom: 0;
    z-index: var(--z-modal-backdrop);
    transform: translateX(-100%);
    transition: transform var(--duration-normal) var(--ease-out);
  }
  .app-aside.is-mobile-open {
    transform: translateX(0);
  }
  .sidebar-overlay {
    display: block;
    position: fixed;
    inset: 0;
    z-index: calc(var(--z-modal-backdrop) - 1);
    background: var(--color-overlay);
  }
}

@media (prefers-reduced-motion: reduce) {
  .app-aside {
    transition: none;
  }
}
</style>