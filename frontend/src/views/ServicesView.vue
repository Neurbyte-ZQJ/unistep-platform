<template>
  <div class="services-page">
    <div class="page-header">
      <h1>服务入口</h1>
      <p class="page-desc">选择以下服务模块，快速进入对应业务</p>
    </div>
    <el-row :gutter="16">
      <el-col
        :xs="24"
        :sm="12"
        :lg="6"
        v-for="service in services"
        :key="service.name"
      >
        <el-card
          class="service-card"
          shadow="hover"
          role="link"
          tabindex="0"
          :aria-label="`进入${service.name}：${service.desc}`"
          @click="goTo(service.route)"
          @keydown.enter="goTo(service.route)"
        >
          <div class="card-icon" :style="{ background: service.bgColor }">
            <el-icon :size="28" :color="service.iconColor"><component :is="service.icon" /></el-icon>
          </div>
          <div class="card-body">
            <span class="card-name">{{ service.name }}</span>
            <span class="card-desc">{{ service.desc }}</span>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { useRouter } from 'vue-router'
import { User, Flag, House, Briefcase } from '@element-plus/icons-vue'

const router = useRouter()

/* Design-token-aligned background constants */
const COLOR_SUCCESS_BG = '#f0fdf4'  // light tint of --color-success
const COLOR_WARNING_BG = '#fffbeb'  // light tint of --color-warning
const COLOR_DANGER_BG = '#fef2f2'   // light tint of --color-danger

const services = [
  { name: '团员发展', desc: '管理团员档案与发展流程', route: 'MemberList', icon: User, bgColor: 'var(--color-brand-subtle)', iconColor: 'var(--color-brand)' },
  { name: '社团活动', desc: '组织与记录社团活动', route: 'ActivityList', icon: Flag, bgColor: COLOR_SUCCESS_BG, iconColor: 'var(--color-success)' },
  { name: '自治队伍', desc: '社区自治队伍管理', route: 'CommunityList', icon: House, bgColor: COLOR_WARNING_BG, iconColor: 'var(--color-warning)' },
  { name: '勤工助学', desc: '勤工助学岗位与申请', route: 'JobList', icon: Briefcase, bgColor: COLOR_DANGER_BG, iconColor: 'var(--color-danger)' },
]

function goTo(name: string) {
  router.push({ name })
}
</script>

<style scoped>
.services-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-6);
}
.page-header h1 {
  margin: 0;
  font-size: var(--text-h1);
  font-weight: 700;
  color: var(--color-ink);
}
.page-desc {
  margin: var(--space-1) 0 0;
  font-size: var(--text-body);
  color: var(--color-ink-muted);
}
.service-card {
  cursor: pointer;
  border-radius: var(--radius-lg);
  transition: transform var(--duration-normal) var(--ease-out),
              box-shadow var(--duration-normal) var(--ease-out);
}
.service-card:hover {
  transform: translateY(-2px);
  box-shadow: var(--shadow-md);
}
.service-card:focus-visible {
  outline: 2px solid var(--color-brand);
  outline-offset: 2px;
}
.service-card :deep(.el-card__body) {
  display: flex;
  align-items: center;
  gap: var(--space-4);
  padding: var(--space-5);
}
.card-icon {
  display: flex;
  align-items: center;
  justify-content: center;
  width: var(--icon-xl);
  height: var(--icon-xl);
  border-radius: var(--radius-md);
  flex-shrink: 0;
}
.card-body {
  display: flex;
  flex-direction: column;
  gap: var(--space-1);
}
.card-name {
  font-size: var(--text-h3);
  font-weight: 600;
  color: var(--color-ink);
}
.card-desc {
  font-size: var(--text-caption);
  color: var(--color-ink-muted);
}

@media (prefers-reduced-motion: reduce) {
  .service-card {
    transition: none;
  }
  .service-card:hover {
    transform: none;
  }
}
</style>
