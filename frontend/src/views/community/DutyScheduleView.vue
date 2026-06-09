<script setup lang="ts">
import { onMounted, reactive, ref, watch } from 'vue'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  communityApi,
  dutyStatusLabel,
  dutyStatusType,
  type DutySchedule,
  type DutyRecord,
  type TeamMember,
  DUTY_STATUS_OPTIONS,
} from '../../api/community'

const loading = ref(false)
const signingIn = ref(false)
const signingOut = ref(false)
const list = ref<DutySchedule[]>([])
const total = ref(0)
const members = ref<TeamMember[]>([])
const currentTeamId = ref<number>(0)

const query = reactive({
  page: 1,
  size: 10,
  date: '',
  status: '',
})

// 创建值班安排
const formVisible = ref(false)
const form = reactive({
  date: '',
  startTime: '',
  endTime: '',
  location: '',
  memberIds: [] as number[],
})

// 签到/签退
const checkinVisible = ref(false)
const currentScheduleId = ref(0)
const checkinUserId = ref(0)

async function fetchMembers() {
  if (!currentTeamId.value) return
  try {
    members.value = await communityApi.listMembers(currentTeamId.value, { status: 'active' })
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function fetchList() {
  if (!currentTeamId.value) return
  loading.value = true
  try {
    const data = await communityApi.listDuties(currentTeamId.value, query)
    list.value = data.items
    total.value = data.total
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function reset() {
  query.date = ''
  query.status = ''
  query.page = 1
  fetchList()
}

function openCreate() {
  Object.assign(form, { date: '', startTime: '', endTime: '', location: '', memberIds: [] })
  formVisible.value = true
}

async function submitForm() {
  try {
    await communityApi.createDuty(currentTeamId.value, form as any)
    ElMessage.success('排班成功')
    formVisible.value = false
    fetchList()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

function openCheckin(scheduleId: number) {
  currentScheduleId.value = scheduleId
  checkinUserId.value = 0
  checkinVisible.value = true
}

async function doCheckin() {
  signingIn.value = true
  try {
    await communityApi.dutyCheckin(currentTeamId.value, currentScheduleId.value, checkinUserId.value)
    ElMessage.success('签到成功')
    checkinVisible.value = false
    fetchList()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    signingIn.value = false
  }
}

async function doCheckout(scheduleId: number, userId: number) {
  signingOut.value = true
  try {
    await ElMessageBox.confirm('确定签退？系统将自动计算值班时长。', '签退确认', { type: 'info' })
    await communityApi.dutyCheckout(currentTeamId.value, scheduleId, userId)
    ElMessage.success('签退成功')
    fetchList()
  } catch (err) {
    if (err === 'cancel') return
    ElMessage.error((err as Error).message)
  } finally {
    signingOut.value = false
  }
}

function formatTime(t: string | null | undefined) {
  if (!t) return '-'
  return t.replace('T', ' ').substring(0, 16)
}

// 从 URL 参数获取 teamId
const urlParams = new URLSearchParams(window.location.search)
const teamIdFromUrl = urlParams.get('teamId')
if (teamIdFromUrl) {
  currentTeamId.value = parseInt(teamIdFromUrl)
}

watch(currentTeamId, () => {
  if (currentTeamId.value) {
    fetchMembers()
    fetchList()
  }
})

onMounted(() => {
  if (currentTeamId.value) {
    fetchMembers()
    fetchList()
  }
})
</script>

<template>
  <div class="duty-page">
    <div class="toolbar">
      <h2>值班安排</h2>
      <el-button type="primary" @click="openCreate" :disabled="!currentTeamId">创建排班</el-button>
    </div>

    <el-alert v-if="!currentTeamId" title="请从队伍列表进入值班安排" type="info" show-icon :closable="false" class="toolbar-gap" />

    <template v-if="currentTeamId">
      <el-form :inline="true" :model="query" class="filter">
        <el-form-item label="日期">
          <el-input v-model="query.date" placeholder="YYYY-MM-DD" clearable class="filter-select" @keyup.enter="fetchList" />
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="query.status" placeholder="全部" clearable class="filter-select-sm">
            <el-option v-for="opt in DUTY_STATUS_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item>
          <el-button type="primary" @click="fetchList">查询</el-button>
          <el-button @click="reset">重置</el-button>
        </el-form-item>
      </el-form>

      <el-table :data="list" v-loading="loading" stripe border>
        <el-table-column prop="date" label="日期" width="120" />
        <el-table-column label="时段" width="140">
          <template #default="{ row }">{{ row.startTime }} - {{ row.endTime }}</template>
        </el-table-column>
        <el-table-column prop="location" label="地点" width="140" />
        <el-table-column label="状态" width="100">
          <template #default="{ row }">
            <el-tag :type="dutyStatusType(row.status)" size="small">{{ dutyStatusLabel(row.status) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="值班人员" min-width="200">
          <template #default="{ row }">
            <template v-if="row.records && row.records.length">
              <div v-for="r in row.records" :key="r.id" class="duty-record">
                <span>{{ r.name }}</span>
                <el-tag v-if="r.status === 'completed'" type="success" size="small">已完成 {{ r.duration?.toFixed(1) }}h</el-tag>
                <el-tag v-else-if="r.status === 'active'" type="primary" size="small">值班中</el-tag>
                <el-tag v-else-if="r.status === 'absent'" type="danger" size="small">缺勤</el-tag>
                <el-tag v-else type="info" size="small">待值班</el-tag>
              </div>
            </template>
            <span v-else>-</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button link type="primary" size="small" @click="openCheckin(row.id)">签到</el-button>
            <template v-if="row.records">
              <el-button v-for="r in row.records" :key="'co-'+r.id" link type="warning" size="small" @click="doCheckout(row.id, r.userId)" v-show="r.status === 'active'" :loading="signingOut">
                签退({{ r.name }})
              </el-button>
            </template>
          </template>
        </el-table-column>
        <template #empty>
          <el-empty description="暂无值班安排" :image-size="80" />
        </template>
      </el-table>

      <el-pagination
        class="pagination"
        v-model:current-page="query.page"
        v-model:page-size="query.size"
        :page-sizes="[10, 20, 50]"
        :total="total"
        layout="total, sizes, prev, pager, next"
        @current-change="fetchList"
        @size-change="fetchList"
      />
    </template>

    <!-- 创建排班 -->
    <el-dialog v-model="formVisible" title="创建值班安排" width="480px" destroy-on-close aria-label="创建值班安排">
      <el-form :model="form" label-width="100px">
        <el-form-item label="日期" required>
          <el-input v-model="form.date" placeholder="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="开始时间" required>
          <el-input v-model="form.startTime" placeholder="HH:mm" />
        </el-form-item>
        <el-form-item label="结束时间" required>
          <el-input v-model="form.endTime" placeholder="HH:mm" />
        </el-form-item>
        <el-form-item label="值班地点">
          <el-input v-model="form.location" />
        </el-form-item>
        <el-form-item label="值班人员" required>
          <el-select v-model="form.memberIds" multiple placeholder="选择成员" class="full-select">
            <el-option v-for="m in members" :key="m.id" :label="`${m.name} (${m.studentNo})`" :value="m.id" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 签到 -->
    <el-dialog v-model="checkinVisible" title="值班签到" width="360px" destroy-on-close aria-label="值班签到">
      <el-form label-width="80px">
        <el-form-item label="用户ID">
          <el-input-number v-model="checkinUserId" :min="1" class="full-select" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="checkinVisible = false">取消</el-button>
        <el-button type="primary" :loading="signingIn" @click="doCheckin">签到</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.duty-page {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
.toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
}
.toolbar h2 {
  font-size: var(--text-h2);
  color: var(--color-ink);
  font-weight: 600;
}
.toolbar-gap {
  margin-bottom: var(--space-4);
}
.filter {
  background: var(--color-surface-sunken);
  padding: var(--space-3);
  border-radius: var(--radius-md);
}
.filter-select {
  width: 160px;
}
.filter-select-sm {
  width: 140px;
}
.full-select {
  width: 100%;
}
.pagination {
  align-self: flex-end;
}
.duty-record {
  display: inline-flex;
  align-items: center;
  gap: var(--space-2);
  margin-right: var(--space-3);
  margin-bottom: var(--space-1);
}
</style>
