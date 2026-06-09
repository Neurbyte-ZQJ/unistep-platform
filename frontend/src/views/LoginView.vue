<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'

const route = useRoute()
const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)

const form = reactive({
  username: '',
  password: '',
})

async function handleLogin() {
  loading.value = true
  try {
    await auth.login(form.username, form.password)
    ElMessage.success('登录成功')
    await router.push((route.query.redirect as string) || '/')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '登录失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <h1>统一登录</h1>
      <p>请输入账号信息进入系统。</p>
      <el-form label-position="top" aria-label="登录表单" @submit.prevent="handleLogin">
        <el-form-item label="账号">
          <el-input v-model="form.username" placeholder="学号 / 工号 / 管理员账号" autocomplete="username" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="请输入密码" show-password autocomplete="current-password" />
        </el-form-item>
        <el-button type="primary" class="full-width" :loading="loading" @click="handleLogin">登录</el-button>
        <div class="auth-link">
          没有账号？<router-link to="/register">立即注册</router-link>
        </div>
      </el-form>
    </el-card>
  </div>
</template>

<style scoped>
.auth-card h1 {
  margin: 0 0 var(--space-2);
  font-size: var(--text-h2);
  font-weight: 700;
  color: var(--color-ink);
}
.auth-card p {
  color: var(--color-ink-muted);
  margin: 0 0 var(--space-4);
  font-size: var(--text-body);
}
</style>
