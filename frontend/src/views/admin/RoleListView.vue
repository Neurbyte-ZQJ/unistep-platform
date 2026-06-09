<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useAuthStore } from '../../stores/auth'
import { ElMessage } from 'element-plus'

const auth = useAuthStore()

interface RoleItem {
  id: number
  code: string
  name: string
  dataScope: string
  description: string
  builtin: boolean
}

interface PermissionItem {
  id: number
  code: string
  name: string
  module: string
  type: string
}

const roles = ref<RoleItem[]>([])
const permissions = ref<PermissionItem[]>([])
const loading = ref(false)

async function fetchData() {
  loading.value = true
  try {
    const [rolesRes, permsRes] = await Promise.all([
      fetch('/api/v1/admin/roles', {
        headers: { Authorization: `Bearer ${auth.token}` },
      }),
      fetch('/api/v1/admin/permissions', {
        headers: { Authorization: `Bearer ${auth.token}` },
      }),
    ])
    const rolesBody = await rolesRes.json()
    const permsBody = await permsRes.json()
    if (rolesRes.ok) roles.value = rolesBody.data ?? []
    if (permsRes.ok) permissions.value = permsBody.data ?? []
  } catch {
    ElMessage.error('网络错误')
  } finally {
    loading.value = false
  }
}

onMounted(fetchData)
</script>

<template>
  <div class="admin-page">
    <h2>角色权限管理</h2>
    <el-card shadow="hover" class="section-card">
      <template #header><span>角色列表</span></template>
      <el-table :data="roles" v-loading="loading" stripe aria-label="角色列表">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="code" label="角色编码" width="140" />
        <el-table-column prop="name" label="角色名称" width="120" />
        <el-table-column prop="dataScope" label="数据作用域" width="120" />
        <el-table-column prop="description" label="描述" min-width="160" />
        <el-table-column prop="builtin" label="内置" width="60">
          <template #default="{ row }">
            <el-tag :type="row.builtin ? 'success' : 'info'" size="small">{{ row.builtin ? '是' : '否' }}</el-tag>
          </template>
        </el-table-column>
      </el-table>
    </el-card>

    <el-card shadow="hover">
      <template #header><span>权限列表</span></template>
      <el-table :data="permissions" v-loading="loading" stripe>
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="code" label="权限码" width="200" />
        <el-table-column prop="name" label="名称" width="160" />
        <el-table-column prop="module" label="模块" width="100" />
        <el-table-column prop="type" label="类型" width="80" />
      </el-table>
    </el-card>
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
.section-card {
  margin-bottom: var(--space-5);
}
</style>