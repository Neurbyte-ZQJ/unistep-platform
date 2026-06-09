<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { ElMessage } from 'element-plus'
import {
  communityApi,
  serviceCategoryLabel,
  memberRoleLabel,
  type ServiceProfile as ServiceProfileType,
  type VolunteerService,
  type DutyRecord,
  type TeamMember,
  SERVICE_CATEGORY_OPTIONS,
  MEMBER_ROLE_OPTIONS,
} from '../../api/community'

const loading = ref(false)
const profile = ref<ServiceProfileType | null>(null)

// 记录志愿服务
const serviceFormVisible = ref(false)
const selectedTeamId = ref(0)
const serviceForm = ref({
  userId: 0,
  name: '',
  studentNo: '',
  title: '',
  date: '',
  hours: 0,
  category: 'community',
  description: '',
})

async function fetchProfile() {
  loading.value = true
  try {
    profile.value = await communityApi.serviceProfile()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function openAddService(teamId: number) {
  selectedTeamId.value = teamId
  Object.assign(serviceForm.value, {
    userId: 0,
    name: '',
    studentNo: '',
    title: '',
    date: new Date().toISOString().slice(0, 10),
    hours: 0,
    category: 'community',
    description: '',
  })
  serviceFormVisible.value = true
}

async function submitService() {
  try {
    await communityApi.createService(selectedTeamId.value, serviceForm.value as any)
    ElMessage.success('记录成功')
    serviceFormVisible.value = false
    fetchProfile()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function verifyService(teamId: number, serviceId: number) {
  try {
    await communityApi.verifyService(teamId, serviceId, true)
    ElMessage.success('已核实')
    fetchProfile()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

onMounted(fetchProfile)
</script>

<template>
  <div class="profile-page" v-loading="loading">
    <h2>服务时长个人档案</h2>

    <template v-if="profile">
      <!-- 总览卡片 -->
      <el-row :gutter="16" class="stat-cards">
        <el-col :xs="24" :sm="12" :lg="8">
          <el-card shadow="hover" aria-label="志愿服务时长">
            <div class="stat-item">
              <div class="stat-value">{{ profile.totalServiceHours.toFixed(1) }}</div>
              <div class="stat-label">志愿服务时长(h)</div>
            </div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :lg="8">
          <el-card shadow="hover" aria-label="值班时长">
            <div class="stat-item">
              <div class="stat-value">{{ profile.totalDutyHours.toFixed(1) }}</div>
              <div class="stat-label">值班时长(h)</div>
            </div>
          </el-card>
        </el-col>
        <el-col :xs="24" :sm="12" :lg="8">
          <el-card shadow="hover" aria-label="总服务时长">
            <div class="stat-item">
              <div class="stat-value highlight">{{ profile.totalHours.toFixed(1) }}</div>
              <div class="stat-label">总服务时长(h)</div>
            </div>
          </el-card>
        </el-col>
      </el-row>

      <!-- 所属队伍 -->
      <el-card class="section-card">
        <template #header>
          <div class="section-header">
            <span>所属队伍</span>
          </div>
        </template>
        <el-table :data="profile.teamMemberships" stripe border size="small" v-if="profile.teamMemberships.length" aria-label="所属队伍列表">
          <el-table-column prop="name" label="姓名" width="100" />
          <el-table-column label="角色" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="row.role === 'leader' ? 'danger' : row.role === 'vice' ? 'warning' : ''">
                {{ memberRoleLabel(row.role) }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="joinDate" label="入队日期" width="120" />
          <el-table-column label="届次" width="160">
            <template #default="{ row }">{{ row.termStart || '-' }} ~ {{ row.termEnd || '-' }}</template>
          </el-table-column>
          <el-table-column label="操作" width="100">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="openAddService(row.teamId)">记录服务</el-button>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无队伍" :image-size="60" />
      </el-card>

      <!-- 志愿服务明细 -->
      <el-card class="section-card">
        <template #header>
          <span>志愿服务明细</span>
        </template>
        <el-table :data="profile.services" stripe border size="small" v-if="profile.services.length" aria-label="志愿服务明细">
          <el-table-column prop="title" label="服务名称" min-width="160" />
          <el-table-column prop="date" label="日期" width="120" />
          <el-table-column prop="hours" label="时长(h)" width="90" />
          <el-table-column label="类别" width="100">
            <template #default="{ row }">{{ serviceCategoryLabel(row.category) }}</template>
          </el-table-column>
          <el-table-column label="核实状态" width="100">
            <template #default="{ row }">
              <el-tag :type="row.verified ? 'success' : 'warning'" size="small">
                {{ row.verified ? '已核实' : '待核实' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="description" label="描述" min-width="120" />
        </el-table>
        <el-empty v-else description="暂无志愿服务记录" :image-size="60" />
      </el-card>

      <!-- 值班记录明细 -->
      <el-card class="section-card">
        <template #header>
          <span>值班记录明细</span>
        </template>
        <el-table :data="profile.dutyRecords" stripe border size="small" v-if="profile.dutyRecords.length" aria-label="值班记录明细">
          <el-table-column prop="name" label="姓名" width="100" />
          <el-table-column label="签到时间" width="160">
            <template #default="{ row }">{{ row.checkinTime ? row.checkinTime.replace('T', ' ').substring(0, 16) : '-' }}</template>
          </el-table-column>
          <el-table-column label="签退时间" width="160">
            <template #default="{ row }">{{ row.checkoutTime ? row.checkoutTime.replace('T', ' ').substring(0, 16) : '-' }}</template>
          </el-table-column>
          <el-table-column label="时长(h)" width="90">
            <template #default="{ row }">{{ row.duration?.toFixed(1) || '-' }}</template>
          </el-table-column>
          <el-table-column label="状态" width="80">
            <template #default="{ row }">
              <el-tag :type="row.status === 'completed' ? 'success' : 'info'" size="small">
                {{ row.status === 'completed' ? '已完成' : row.status }}
              </el-tag>
            </template>
          </el-table-column>
        </el-table>
        <el-empty v-else description="暂无值班记录" :image-size="60" />
      </el-card>
    </template>

    <!-- 记录志愿服务 -->
    <el-dialog v-model="serviceFormVisible" title="记录志愿服务" width="520px" destroy-on-close aria-label="记录志愿服务">
      <el-form :model="serviceForm" label-width="100px">
        <el-form-item label="用户ID" required>
          <el-input-number v-model="serviceForm.userId" :min="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="姓名" required>
          <el-input v-model="serviceForm.name" maxlength="64" />
        </el-form-item>
        <el-form-item label="学号" required>
          <el-input v-model="serviceForm.studentNo" maxlength="32" />
        </el-form-item>
        <el-form-item label="服务名称" required>
          <el-input v-model="serviceForm.title" maxlength="255" />
        </el-form-item>
        <el-form-item label="服务日期" required>
          <el-input v-model="serviceForm.date" placeholder="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="服务时长(h)" required>
          <el-input-number v-model="serviceForm.hours" :min="0.5" :step="0.5" :precision="1" style="width: 100%" />
        </el-form-item>
        <el-form-item label="服务类别" required>
          <el-select v-model="serviceForm.category" style="width: 100%">
            <el-option v-for="opt in SERVICE_CATEGORY_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="serviceForm.description" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="serviceFormVisible = false">取消</el-button>
        <el-button type="primary" @click="submitService">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.profile-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
.stat-cards {
  margin-bottom: var(--space-2);
}
.stat-item {
  text-align: center;
  padding: var(--space-3) 0;
}
.stat-value {
  font-size: var(--text-stat);
  font-weight: 700;
  color: var(--color-brand);
}
.stat-value.highlight {
  color: var(--color-success);
}
.stat-label {
  font-size: var(--text-body);
  color: var(--color-ink-muted);
  margin-top: var(--space-1);
}
.section-card {
  margin-bottom: 0;
}
.section-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
</style>
