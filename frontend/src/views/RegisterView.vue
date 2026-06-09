<script setup lang="ts">
import { reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage } from 'element-plus'
import { useAuthStore } from '../stores/auth'

const router = useRouter()
const auth = useAuthStore()
const loading = ref(false)

const form = reactive({
  username: '',
  email: '',
  password: '',
})

async function handleRegister() {
  loading.value = true
  try {
    await auth.register(form.username, form.password, form.email)
    ElMessage.success('注册成功，请登录')
    await router.push('/login')
  } catch (error) {
    ElMessage.error(error instanceof Error ? error.message : '注册失败')
  } finally {
    loading.value = false
  }
}
</script>

<template>
  <div class="auth-page">
    <el-card class="auth-card">
      <h1>用户注册</h1>
      <p>创建账号后即可进入学生“一站式”服务平台。</p>
      <el-form label-position="top" aria-label="注册表单" @submit.prevent="handleRegister">
        <el-form-item label="账号">
          <el-input v-model="form.username" placeholder="请输入账号，至少 3 位" autocomplete="username" />
        </el-form-item>
        <el-form-item label="邮箱">
          <el-input v-model="form.email" placeholder="请输入邮箱，可选" />
        </el-form-item>
        <el-form-item label="密码">
          <el-input v-model="form.password" type="password" placeholder="请输入密码，至少 6 位" show-password autocomplete="new-password" />
        </el-form-item>
        <el-button type="primary" class="full-width" :loading="loading" @click="handleRegister">注册</el-button>
        <div class="auth-link">
          已有账号？<router-link to="/login">返回登录</router-link>
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
