<template>
  <div class="dashboard">
    <!-- 标题区 -->
    <header class="dashboard-header">
      <div class="dashboard-header-text">
        <h1>{{ titleByRole }}</h1>
        <p class="dashboard-tip">{{ tipByRole }}</p>
      </div>
      <el-tag :type="roleTagType" effect="dark" size="large">{{ roleLabel }}</el-tag>
    </header>

    <!-- 加载态 -->
    <div v-if="loading" class="dashboard-skeleton" aria-label="加载中">
      <el-row :gutter="16">
        <el-col v-for="i in 4" :key="i" :xs="12" :sm="6">
          <el-card shadow="never" class="stat-card stat-card--skeleton">
            <el-skeleton :rows="2" animated />
          </el-card>
        </el-col>
      </el-row>
      <el-row :gutter="16" class="dashboard-skeleton-charts">
        <el-col :xs="24" :md="12">
          <el-card shadow="never"><el-skeleton :rows="8" animated /></el-card>
        </el-col>
        <el-col :xs="24" :md="12">
          <el-card shadow="never"><el-skeleton :rows="8" animated /></el-card>
        </el-col>
      </el-row>
    </div>

    <!-- 错误态 -->
    <el-result
      v-else-if="loadError"
      icon="warning"
      title="数据加载失败"
      :sub-title="loadError"
    >
      <template #extra>
        <el-button type="primary" @click="fetchData">重新加载</el-button>
      </template>
    </el-result>

    <!-- 正常内容 -->
    <template v-else-if="overview">

      <!-- 概览卡片 -->
      <section class="stat-section" aria-label="数据概览">
        <el-row :gutter="16">
          <el-col v-for="card in statCards" :key="card.label" :xs="12" :sm="6">
            <el-card
              shadow="hover"
              class="stat-card"
              :class="{ 'stat-card--clickable': card.route }"
              :tabindex="card.route ? 0 : undefined"
              :role="card.route ? 'link' : undefined"
              @click="card.route && $router.push(card.route)"
              @keydown.enter="card.route && $router.push(card.route)"
            >
              <div class="stat-card-inner">
                <div class="stat-card-icon" :style="{ background: card.bgColor }">
                  <el-icon :size="20" :style="{ color: card.color }"><component :is="card.icon" /></el-icon>
                </div>
                <div class="stat-card-content">
                  <div class="stat-value" :style="{ color: card.color }">{{ card.value }}</div>
                  <div class="stat-label">{{ card.label }}</div>
                </div>
              </div>
            </el-card>
          </el-col>
        </el-row>
      </section>

      <!-- 图表区域：根据角色选择呈现 -->
      <section class="chart-section" aria-label="数据图表">
        <el-row :gutter="16">
          <el-col v-if="showMemberCharts" :xs="24" :md="12">
            <el-card shadow="hover" class="chart-card">
              <template #header><span>团员阶段分布</span></template>
              <div ref="memberPieRef" class="chart-container"></div>
            </el-card>
          </el-col>
          <el-col v-if="showActivityCharts" :xs="24" :md="12">
            <el-card shadow="hover" class="chart-card">
              <template #header><span>活动状态分布</span></template>
              <div ref="activityPieRef" class="chart-container"></div>
            </el-card>
          </el-col>
        </el-row>

        <el-row :gutter="16" v-if="showTrendCharts">
          <el-col :xs="24" :md="12">
            <el-card shadow="hover" class="chart-card">
              <template #header><span>团员发展趋势</span></template>
              <div ref="memberTrendRef" class="chart-container"></div>
            </el-card>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-card shadow="hover" class="chart-card">
              <template #header><span>活动趋势</span></template>
              <div ref="activityTrendRef" class="chart-container"></div>
            </el-card>
          </el-col>
        </el-row>

        <el-row :gutter="16">
          <el-col v-if="showServiceTrend" :xs="24" :md="12">
            <el-card shadow="hover" class="chart-card">
              <template #header><span>服务时长趋势</span></template>
              <div ref="serviceTrendRef" class="chart-container"></div>
            </el-card>
          </el-col>
          <el-col v-if="showWorkstudyChart" :xs="24" :md="12">
            <el-card shadow="hover" class="chart-card">
              <template #header><span>勤工助学岗位统计</span></template>
              <div ref="workstudyPieRef" class="chart-container"></div>
            </el-card>
          </el-col>
        </el-row>
      </section>

      <!-- 学生视图：个人画像 -->
      <section v-if="primaryRole === 'student'" class="profile-section" aria-label="个人画像">
        <el-card shadow="hover" class="profile-card">
          <template #header><span>我的服务画像</span></template>
          <el-descriptions :column="2" border>
            <el-descriptions-item label="学院">{{ auth.user?.college || '—' }}</el-descriptions-item>
            <el-descriptions-item label="班级">{{ auth.user?.className || '—' }}</el-descriptions-item>
            <el-descriptions-item label="参与活动">{{ overview?.activities?.totalRegistrations ?? 0 }} 次</el-descriptions-item>
            <el-descriptions-item label="服务时长">{{ overview?.services?.totalHours?.toFixed(1) ?? '0.0' }} 小时</el-descriptions-item>
          </el-descriptions>
          <div class="quick-actions">
            <el-button type="primary" @click="$router.push('/services')">前往服务入口</el-button>
            <el-button @click="$router.push('/community/duty')">值班签到</el-button>
            <el-button @click="$router.push('/workstudy')">浏览岗位</el-button>
          </div>
        </el-card>
      </section>

      <!-- admin / teacher：服务 + 勤工助学明细 -->
      <section v-if="showDetails" class="detail-section" aria-label="数据明细">
        <el-row :gutter="16">
          <el-col :xs="24" :md="12">
            <el-card shadow="hover" class="detail-card">
              <template #header><span>服务时长明细</span></template>
              <el-descriptions :column="2" border>
                <el-descriptions-item label="志愿服务时长">{{ overview?.services?.totalServiceHours?.toFixed(1) ?? '0.0' }} 小时</el-descriptions-item>
                <el-descriptions-item label="值班时长">{{ overview?.services?.totalDutyHours?.toFixed(1) ?? '0.0' }} 小时</el-descriptions-item>
                <el-descriptions-item label="志愿服务次数">{{ overview?.services?.totalVolunteerCount ?? 0 }} 次</el-descriptions-item>
                <el-descriptions-item label="值班记录数">{{ overview?.services?.totalDutyRecordCount ?? 0 }} 条</el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-col>
          <el-col :xs="24" :md="12">
            <el-card shadow="hover" class="detail-card">
              <template #header><span>勤工助学明细</span></template>
              <el-descriptions :column="2" border>
                <el-descriptions-item label="岗位总数">{{ overview?.workstudy?.totalJobs ?? 0 }}</el-descriptions-item>
                <el-descriptions-item label="报名总数">{{ overview?.workstudy?.totalApplications ?? 0 }}</el-descriptions-item>
                <el-descriptions-item label="录用人数">{{ overview?.workstudy?.totalAccepted ?? 0 }}</el-descriptions-item>
                <el-descriptions-item label="已发薪资">¥{{ overview?.workstudy?.totalSalaryPaid?.toFixed(2) ?? '0.00' }}</el-descriptions-item>
                <el-descriptions-item label="总工时">{{ overview?.workstudy?.totalWorkHours?.toFixed(1) ?? '0.0' }} 小时</el-descriptions-item>
                <el-descriptions-item label="活动报名">{{ overview?.activities?.totalRegistrations ?? 0 }} 人</el-descriptions-item>
              </el-descriptions>
            </el-card>
          </el-col>
        </el-row>
      </section>

      <!-- admin 专属：快捷管理入口 -->
      <section v-if="primaryRole === 'admin'" class="admin-section" aria-label="管理快捷入口">
        <el-card shadow="hover" class="admin-card">
          <template #header><span>管理快捷入口</span></template>
          <div class="quick-actions">
            <el-button type="primary" @click="$router.push('/admin/users')">用户管理</el-button>
            <el-button type="success" @click="$router.push('/admin/roles')">角色权限</el-button>
          </div>
        </el-card>
      </section>

    </template>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted, nextTick, watch, shallowRef } from 'vue'
import * as echarts from 'echarts'
import { dashboardApi, type DashboardOverview, type TrendItem, type ServiceTrendData } from '../api/dashboard'
import { useAuthStore } from '../stores/auth'
import { User, Flag, Timer, Briefcase, Calendar, Document } from '@element-plus/icons-vue'

const auth = useAuthStore()
const overview = ref<DashboardOverview | null>(null)
const memberTrendData = ref<TrendItem[]>([])
const activityTrendData = ref<TrendItem[]>([])
const serviceTrendData = ref<ServiceTrendData | null>(null)
const loading = ref(true)
const loadError = ref('')

// ===== 角色驱动的视图开关 =====
const primaryRole = computed(() => auth.primaryRole)
const roleLabels: Record<string, string> = {
  admin: '系统管理员',
  teacher: '教师',
  student_cadre: '学生干部',
  student: '普通学生',
}
const roleLabel = computed(() => roleLabels[primaryRole.value] || primaryRole.value)
const roleTagType = computed(() => {
  switch (primaryRole.value) {
    case 'admin':
      return 'danger'
    case 'teacher':
      return 'warning'
    case 'student_cadre':
      return 'success'
    default:
      return 'info'
  }
})

const titleByRole = computed(() => {
  switch (primaryRole.value) {
    case 'admin':
      return '平台总览'
    case 'teacher':
      return '教师工作台'
    case 'student_cadre':
      return '学生干部工作台'
    default:
      return '我的工作台'
  }
})

const tipByRole = computed(() => {
  switch (primaryRole.value) {
    case 'admin':
      return '全平台数据汇总，含用户、活动、社区与勤工助学统计。'
    case 'teacher':
      return '聚焦审批与统计，可查看本学院的团员发展、活动与服务情况。'
    case 'student_cadre':
      return '管理本社团 / 队伍的活动与值班，关注成员的成长画像。'
    default:
      return '查看你参与的活动、服务和岗位申请进度。'
  }
})

// 各角色可见的统计卡片
const statCards = computed(() => {
  const o = overview.value
  if (!o) return []
  const role = primaryRole.value
  if (role === 'student') {
    return [
      { label: '我可报名活动', value: o.activities?.total ?? 0, color: 'var(--color-success)', bgColor: 'var(--color-brand-subtle)', icon: Calendar, route: '/activities' },
      { label: '我的服务时长(h)', value: (o.services?.totalHours ?? 0).toFixed(1), color: 'var(--color-warning)', bgColor: COLOR_WARNING_BG, icon: Timer, route: '/community/profile' },
      { label: '在招岗位', value: o.workstudy?.totalJobs ?? 0, color: 'var(--color-danger)', bgColor: COLOR_DANGER_BG, icon: Briefcase, route: '/workstudy' },
      { label: '我的报名记录', value: o.activities?.totalRegistrations ?? 0, color: 'var(--color-brand)', bgColor: 'var(--color-brand-subtle)', icon: Document, route: undefined },
    ]
  }
  return [
    { label: '团员总数', value: o.members?.total ?? 0, color: 'var(--color-brand)', bgColor: 'var(--color-brand-subtle)', icon: User, route: '/members' },
    { label: '活动总数', value: o.activities?.total ?? 0, color: 'var(--color-success)', bgColor: COLOR_SUCCESS_BG, icon: Flag, route: '/activities' },
    { label: '服务总时长(h)', value: (o.services?.totalHours ?? 0).toFixed(1), color: 'var(--color-warning)', bgColor: COLOR_WARNING_BG, icon: Timer, route: '/community/profile' },
    { label: '勤工助学岗位', value: o.workstudy?.totalJobs ?? 0, color: 'var(--color-danger)', bgColor: COLOR_DANGER_BG, icon: Briefcase, route: '/workstudy' },
  ]
})

// 图表显示开关
const showMemberCharts = computed(() => ['admin', 'teacher', 'student_cadre'].includes(primaryRole.value))
const showActivityCharts = computed(() => primaryRole.value !== 'student')
const showTrendCharts = computed(() => ['admin', 'teacher'].includes(primaryRole.value))
const showServiceTrend = computed(() => ['admin', 'teacher', 'student_cadre'].includes(primaryRole.value))
const showWorkstudyChart = computed(() => ['admin', 'teacher'].includes(primaryRole.value))
const showDetails = computed(() => ['admin', 'teacher'].includes(primaryRole.value))

// 图表 DOM 引用
const memberPieRef = ref<HTMLElement>()
const activityPieRef = ref<HTMLElement>()
const memberTrendRef = ref<HTMLElement>()
const activityTrendRef = ref<HTMLElement>()
const serviceTrendRef = ref<HTMLElement>()
const workstudyPieRef = ref<HTMLElement>()

const charts = shallowRef<echarts.ECharts[]>([])

const stageLabels: Record<string, string> = {
  applicant: '入团申请人',
  activist: '积极分子',
  develop_target: '发展对象',
  political_review: '政审备案',
  league_member: '正式团员',
}
const activityStatusLabels: Record<string, string> = {
  draft: '草稿',
  pending: '待审批',
  rejected: '已驳回',
  reg_open: '报名开放',
  reg_closed: '报名截止',
  in_progress: '进行中',
  completed: '已完成',
  archived: '已归档',
}
const jobStatusLabels: Record<string, string> = {
  draft: '草稿',
  published: '已发布',
  closed: '已关闭',
  completed: '已完成',
}
/* Design-token-aligned color constants (CSS variables cannot be used in ECharts / JS :style directly) */
const COLOR_BRAND = '#1d4ed8'          // --color-brand
const COLOR_SUCCESS = '#67C23A'        // --color-success
const COLOR_WARNING = '#E6A23C'        // --color-warning
const COLOR_DANGER = '#F56C6C'         // --color-danger
const COLOR_INFO = '#909399'           // --color-info
const COLOR_TEAL = '#00D1B2'
const COLOR_PURPLE = '#8B5CF6'
const COLOR_PINK = '#EC4899'
const COLOR_SURFACE_RAISED = '#ffffff'  // --color-surface-raised

const COLOR_SUCCESS_BG = '#f0f9eb'     // light tint of --color-success
const COLOR_WARNING_BG = '#fef3e2'     // light tint of --color-warning
const COLOR_DANGER_BG = '#fef0f0'      // light tint of --color-danger

const colorPalette = [COLOR_BRAND, COLOR_SUCCESS, COLOR_WARNING, COLOR_DANGER, COLOR_INFO, COLOR_TEAL, COLOR_PURPLE, COLOR_PINK]

function initChart(el: HTMLElement): echarts.ECharts {
  const chart = echarts.init(el)
  charts.value = [...charts.value, chart]
  return chart
}

function renderMemberPie(data: DashboardOverview) {
  if (!memberPieRef.value) return
  const chart = initChart(memberPieRef.value)
  const pieData = (data.members.stageBreakdown || []).map(item => ({
    name: stageLabels[item.stage] || item.stage,
    value: item.count,
  }))
  chart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    color: colorPalette,
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: true,
      itemStyle: { borderRadius: 6, borderColor: COLOR_SURFACE_RAISED, borderWidth: 2 },
      label: { show: true, formatter: '{b}\n{d}%', fontSize: 12 },
      data: pieData,
    }],
  })
}

function renderActivityPie(data: DashboardOverview) {
  if (!activityPieRef.value) return
  const chart = initChart(activityPieRef.value)
  const pieData = (data.activities.statusBreakdown || []).map(item => ({
    name: activityStatusLabels[item.status] || item.status,
    value: item.count,
  }))
  chart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    color: colorPalette,
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: true,
      itemStyle: { borderRadius: 6, borderColor: COLOR_SURFACE_RAISED, borderWidth: 2 },
      label: { show: true, formatter: '{b}\n{d}%', fontSize: 12 },
      data: pieData,
    }],
  })
}

function renderWorkstudyPie(data: DashboardOverview) {
  if (!workstudyPieRef.value) return
  const chart = initChart(workstudyPieRef.value)
  const pieData = (data.workstudy.statusBreakdown || []).map(item => ({
    name: jobStatusLabels[item.status] || item.status,
    value: item.count,
  }))
  chart.setOption({
    tooltip: { trigger: 'item', formatter: '{b}: {c} ({d}%)' },
    color: colorPalette,
    series: [{
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: true,
      itemStyle: { borderRadius: 6, borderColor: COLOR_SURFACE_RAISED, borderWidth: 2 },
      label: { show: true, formatter: '{b}\n{d}%', fontSize: 12 },
      data: pieData,
    }],
  })
}

function renderMemberTrend(trend: TrendItem[]) {
  if (!memberTrendRef.value) return
  const chart = initChart(memberTrendRef.value)
  chart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: trend.map(i => i.month), boundaryGap: false },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{
      type: 'line',
      data: trend.map(i => i.count),
      smooth: true,
      areaStyle: { opacity: 0.12 },
      itemStyle: { color: COLOR_BRAND },
    }],
  })
}

function renderActivityTrend(trend: TrendItem[]) {
  if (!activityTrendRef.value) return
  const chart = initChart(activityTrendRef.value)
  chart.setOption({
    tooltip: { trigger: 'axis' },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: trend.map(i => i.month), boundaryGap: false },
    yAxis: { type: 'value', minInterval: 1 },
    series: [{
      type: 'line',
      data: trend.map(i => i.count),
      smooth: true,
      areaStyle: { opacity: 0.12 },
      itemStyle: { color: COLOR_SUCCESS },
    }],
  })
}

function renderServiceTrend(data: ServiceTrendData) {
  if (!serviceTrendRef.value) return
  const chart = initChart(serviceTrendRef.value)
  const allMonths = [...new Set([
    ...data.volunteer.map(i => i.month),
    ...data.duty.map(i => i.month),
  ])].sort()
  const volMap = Object.fromEntries(data.volunteer.map(i => [i.month, i.hours]))
  const dutyMap = Object.fromEntries(data.duty.map(i => [i.month, i.hours]))
  chart.setOption({
    tooltip: { trigger: 'axis' },
    legend: { data: ['志愿服务', '值班'] },
    grid: { left: '3%', right: '4%', bottom: '3%', containLabel: true },
    xAxis: { type: 'category', data: allMonths, boundaryGap: false },
    yAxis: { type: 'value' },
    series: [
      {
        name: '志愿服务',
        type: 'line',
        data: allMonths.map(m => volMap[m] ?? 0),
        smooth: true,
        areaStyle: { opacity: 0.12 },
        itemStyle: { color: COLOR_WARNING },
      },
      {
        name: '值班',
        type: 'line',
        data: allMonths.map(m => dutyMap[m] ?? 0),
        smooth: true,
        areaStyle: { opacity: 0.12 },
        itemStyle: { color: COLOR_DANGER },
      },
    ],
  })
}

function disposeCharts() {
  charts.value.forEach(c => c.dispose())
  charts.value = []
}

async function renderAll() {
  if (!overview.value) return
  await nextTick()
  disposeCharts()
  if (showMemberCharts.value) renderMemberPie(overview.value)
  if (showActivityCharts.value) renderActivityPie(overview.value)
  if (showWorkstudyChart.value) renderWorkstudyPie(overview.value)
  if (showTrendCharts.value) {
    renderMemberTrend(memberTrendData.value)
    renderActivityTrend(activityTrendData.value)
  }
  if (showServiceTrend.value && serviceTrendData.value) {
    renderServiceTrend(serviceTrendData.value)
  }
}

function handleResize() {
  charts.value.forEach(c => c.resize())
}

async function fetchData() {
  loading.value = true
  loadError.value = ''
  try {
    const [o, m, a, s] = await Promise.all([
      dashboardApi.overview(),
      dashboardApi.memberTrend(),
      dashboardApi.activityTrend(),
      dashboardApi.serviceTrend(),
    ])
    overview.value = o
    memberTrendData.value = m.trend
    activityTrendData.value = a.trend
    serviceTrendData.value = s
    await renderAll()
  } catch (e: unknown) {
    const msg = e instanceof Error ? e.message : '请检查网络连接后重试'
    loadError.value = msg
  } finally {
    loading.value = false
  }
}

onMounted(async () => {
  await fetchData()
  window.addEventListener('resize', handleResize)
})

watch(primaryRole, () => {
  renderAll()
})

onUnmounted(() => {
  disposeCharts()
  window.removeEventListener('resize', handleResize)
})
</script>

<style scoped>
.dashboard {
  display: flex;
  flex-direction: column;
  gap: var(--space-5);
}

/* 骨架屏图表行 */
.dashboard-skeleton-charts {
  margin-top: var(--space-4);
}

/* 标题区 */
.dashboard-header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: var(--space-4);
}

.dashboard-header h1 {
  margin: 0;
  font-size: var(--text-h1);
  font-weight: 700;
  line-height: 1.3;
  color: var(--color-ink);
}

.dashboard-tip {
  color: var(--color-ink-muted);
  font-size: var(--text-body);
  margin: var(--space-1) 0 0;
  line-height: 1.5;
}

/* 概览卡片 */
.stat-section {
  margin: 0;
}

.stat-card {
  transition: box-shadow var(--duration-fast) var(--ease-out);
}

.stat-card--clickable {
  cursor: pointer;
}

.stat-card--clickable:hover {
  box-shadow: var(--shadow-md);
}

.stat-card--clickable:focus-visible {
  outline: 2px solid var(--color-brand);
  outline-offset: 2px;
}

.stat-card--skeleton {
  min-height: var(--stat-card-min-h);
}

.stat-card-inner {
  display: flex;
  align-items: center;
  gap: var(--space-4);
}

.stat-card-icon {
  flex-shrink: 0;
  width: var(--icon-lg);
  height: var(--icon-lg);
  border-radius: var(--radius-lg);
  display: flex;
  align-items: center;
  justify-content: center;
}

.stat-card-content {
  min-width: 0;
}

.stat-value {
  font-size: var(--text-stat);
  font-weight: 700;
  line-height: 1.3;
}

.stat-label {
  font-size: var(--text-caption);
  color: var(--color-ink-muted);
  margin-top: 2px;
}

/* 图表区 */
.chart-section {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}

.chart-card {
  margin-bottom: 0;
}

.chart-container {
  width: 100%;
  height: var(--chart-height);
}

@media (max-width: 767px) {
  .chart-container {
    height: var(--chart-height-mobile);
  }
}

/* 画像区 */
.profile-section,
.detail-section,
.admin-section {
  margin: 0;
}

.profile-card,
.detail-card,
.admin-card {
  margin-bottom: 0;
}

.quick-actions {
  display: flex;
  gap: var(--space-3);
  margin-top: var(--space-4);
  flex-wrap: wrap;
}

/* 响应式 */
@media (max-width: 767px) {
  .dashboard-header {
    flex-direction: column;
    gap: var(--space-2);
  }

  .dashboard-header h1 {
    font-size: var(--text-h2);
  }

  .stat-card-inner {
    gap: var(--space-3);
  }

  .stat-card-icon {
    width: var(--icon-md);
    height: var(--icon-md);
  }

  .stat-value {
    font-size: var(--text-h3);
  }
}
</style>
