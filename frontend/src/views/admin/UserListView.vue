<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { ElMessage } from 'element-plus'

const auth = useAuthStore()

interface UserItem {
  id: number
  username: string
  email: string
  realName: string
  college: string
  roles: string
  status: string
  createdAt: string
}

const users = ref<UserItem[]>([])
const loading = ref(false)

async function fetchUsers() {
  loading.value = true
  try {
    const res = await fetch('/api/v1/admin/users', {
      headers: { Authorization: `Bearer ${auth.token}` },
    })
    const body = await res.json()
    if (res.ok) {
      users.value = body.data ?? []
    } else {
      ElMessage.error(body.message || '加载失败')
    }
  } catch {
    ElMessage.error('网络错误')
  } finally {
    loading.value = false
  }
}

onMounted(fetchUsers)
</script>

<template>
  <div class="admin-page">
    <h2>用户管理</h2>
    <el-table :data="users" v-loading="loading" stripe class="full-select" aria-label="用户列表">
      <el-table-column prop="id" label="ID" width="60" />
      <el-table-column prop="username" label="用户名" min-width="120" />
      <el-table-column prop="realName" label="姓名" min-width="100" />
      <el-table-column prop="email" label="邮箱" min-width="160" />
      <el-table-column prop="college" label="学院" min-width="120" />
      <el-table-column prop="roles" label="角色" min-width="140" />
      <el-table-column prop="status" label="状态" width="80" />
    </el-table>
  </div>
</template>

<style scoped>
.admin-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
.admin-page h2 {
  margin: 0;
  font-size: var(--text-h2);
  color: var(--color-ink);
}
.full-select {
  width: 100%;
}
</style>