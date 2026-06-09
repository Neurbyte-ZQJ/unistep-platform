<script setup lang="ts">
import { useRouter } from 'vue-router'
const router = useRouter()
</script>

<template>
  <div class="forbidden-page">
    <el-result icon="warning" title="403 - 权限不足" sub-title="您没有访问此页面的权限，请联系管理员。" role="alert">
      <template #extra>
        <el-button type="primary" @click="router.push('/')">返回工作台</el-button>
      </template>
    </el-result>
  </div>
</template>

<style scoped>
.forbidden-page {
  display: flex;
  justify-content: center;
  align-items: center;
  min-height: 60vh;
  padding: var(--space-6);
}
</style>