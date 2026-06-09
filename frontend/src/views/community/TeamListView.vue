<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import {
  communityApi,
  teamTypeLabel,
  type CommunityTeam,
  type TeamMember,
  TEAM_TYPE_OPTIONS,
  MEMBER_ROLE_OPTIONS,
} from '../../api/community'

const router = useRouter()
const loading = ref(false)
const submitting = ref(false)
const list = ref<CommunityTeam[]>([])
const total = ref(0)

const query = reactive({
  page: 1,
  size: 10,
  teamType: '',
  name: '',
  status: '',
})

// 队伍详情/成员管理
const detailVisible = ref(false)
const currentTeam = ref<CommunityTeam | null>(null)
const memberList = ref<TeamMember[]>([])

// 创建/编辑队伍
const formVisible = ref(false)
const formTitle = ref('创建队伍')
const editingId = ref<number | null>(null)
const form = reactive({
  name: '',
  teamType: 'autonomy',
  description: '',
  quota: 0,
  location: '',
  contactInfo: '',
})

// 添加成员
const addMemberVisible = ref(false)
const memberForm = reactive({
  userId: 0,
  name: '',
  studentNo: '',
  role: 'member',
  joinDate: '',
  termStart: '',
  termEnd: '',
  remark: '',
})

async function fetchList() {
  loading.value = true
  try {
    const data = await communityApi.listTeams(query)
    list.value = data.items
    total.value = data.total
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function reset() {
  query.teamType = ''
  query.name = ''
  query.status = ''
  query.page = 1
  fetchList()
}

function openCreate() {
  formTitle.value = '创建队伍'
  editingId.value = null
  Object.assign(form, { name: '', teamType: 'autonomy', description: '', quota: 0, location: '', contactInfo: '' })
  formVisible.value = true
}

function openEdit(row: CommunityTeam) {
  formTitle.value = '编辑队伍'
  editingId.value = row.id
  Object.assign(form, {
    name: row.name,
    teamType: row.teamType,
    description: row.description,
    quota: row.quota,
    location: row.location,
    contactInfo: row.contactInfo,
  })
  formVisible.value = true
}

async function submitForm() {
  submitting.value = true
  try {
    if (editingId.value) {
      await communityApi.updateTeam(editingId.value, form)
      ElMessage.success('更新成功')
    } else {
      await communityApi.createTeam(form)
      ElMessage.success('创建成功')
    }
    formVisible.value = false
    fetchList()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    submitting.value = false
  }
}

async function disbandTeam(row: CommunityTeam) {
  try {
    await ElMessageBox.confirm(`确定解散队伍「${row.name}」吗？`, '提示', { type: 'warning' })
    await communityApi.deleteTeam(row.id)
    ElMessage.success('已解散')
    fetchList()
  } catch (err) {
    if (err === 'cancel') return
    ElMessage.error((err as Error).message)
  }
}

async function showDetail(row: CommunityTeam) {
  try {
    const data = await communityApi.getTeam(row.id)
    currentTeam.value = data
    memberList.value = data.members || []
    detailVisible.value = true
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

function openAddMember() {
  Object.assign(memberForm, { userId: 0, name: '', studentNo: '', role: 'member', joinDate: new Date().toISOString().slice(0, 10), termStart: '', termEnd: '', remark: '' })
  addMemberVisible.value = true
}

async function submitAddMember() {
  if (!currentTeam.value) return
  try {
    await communityApi.addMember(currentTeam.value.id, memberForm)
    ElMessage.success('添加成功')
    addMemberVisible.value = false
    showDetail(currentTeam.value)
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function removeMember(member: TeamMember) {
  if (!currentTeam.value) return
  try {
    await ElMessageBox.confirm(`确定移除成员「${member.name}」吗？`, '提示', { type: 'warning' })
    await communityApi.removeMember(currentTeam.value.id, member.id)
    ElMessage.success('已移除')
    showDetail(currentTeam.value)
  } catch (err) {
    if (err === 'cancel') return
    ElMessage.error((err as Error).message)
  }
}

function goDutySchedule(teamId: number) {
  router.push({ name: 'CommunityDuty', query: { teamId: String(teamId) } })
}

function goServiceProfile() {
  router.push({ name: 'ServiceProfile' })
}

onMounted(fetchList)
</script>

<template>
  <div class="community-page">
    <div class="toolbar">
      <h2>社区队伍</h2>
      <div>
        <el-button @click="goServiceProfile">我的服务档案</el-button>
        <el-button type="primary" @click="openCreate">创建队伍</el-button>
      </div>
    </div>

    <el-form :inline="true" :model="query" class="filter">
      <el-form-item label="队伍名称">
        <el-input v-model="query.name" placeholder="支持模糊搜索" clearable @keyup.enter="fetchList" />
      </el-form-item>
      <el-form-item label="队伍类型">
        <el-select v-model="query.teamType" placeholder="全部" clearable class="filter-select">
          <el-option v-for="opt in TEAM_TYPE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="fetchList">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" v-loading="loading" stripe border>
      <el-table-column prop="name" label="队伍名称" min-width="160" />
      <el-table-column label="类型" width="120">
        <template #default="{ row }">{{ teamTypeLabel(row.teamType) }}</template>
      </el-table-column>
      <el-table-column prop="quota" label="编制" width="80" />
      <el-table-column prop="location" label="地点" width="140" />
      <el-table-column prop="contactInfo" label="联系方式" width="160" />
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'info'">{{ row.status === 'active' ? '活跃' : '已解散' }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="300" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="showDetail(row)">详情</el-button>
          <el-button link type="primary" @click="openEdit(row)" v-if="row.status === 'active'">编辑</el-button>
          <el-button link type="primary" @click="goDutySchedule(row.id)" v-if="row.status === 'active'">值班安排</el-button>
          <el-button link type="danger" @click="disbandTeam(row)" v-if="row.status === 'active'">解散</el-button>
        </template>
      </el-table-column>
      <template #empty>
        <el-empty description="暂无队伍数据" :image-size="80" />
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

    <!-- 队伍详情/成员管理 -->
    <el-dialog v-model="detailVisible" :title="currentTeam?.name || '队伍详情'" width="800px" destroy-on-close aria-label="队伍详情">
      <template v-if="currentTeam">
        <el-descriptions :column="2" border class="team-info">
          <el-descriptions-item label="队伍类型">{{ teamTypeLabel(currentTeam.teamType) }}</el-descriptions-item>
          <el-descriptions-item label="编制人数">{{ currentTeam.quota || '不限' }}</el-descriptions-item>
          <el-descriptions-item label="活动地点">{{ currentTeam.location || '-' }}</el-descriptions-item>
          <el-descriptions-item label="联系方式">{{ currentTeam.contactInfo || '-' }}</el-descriptions-item>
          <el-descriptions-item label="简介" :span="2">{{ currentTeam.description || '-' }}</el-descriptions-item>
        </el-descriptions>

        <div class="member-header">
          <h4>成员列表</h4>
          <el-button type="primary" size="small" @click="openAddMember">纳新</el-button>
        </div>
        <el-table :data="memberList" stripe border size="small">
          <el-table-column prop="name" label="姓名" width="100" />
          <el-table-column prop="studentNo" label="学号" width="120" />
          <el-table-column label="角色" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="row.role === 'leader' ? 'danger' : row.role === 'vice' ? 'warning' : ''">
                {{ MEMBER_ROLE_OPTIONS.find(o => o.value === row.role)?.label || row.role }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column prop="joinDate" label="入队日期" width="120" />
          <el-table-column label="届次" width="160">
            <template #default="{ row }">{{ row.termStart || '-' }} ~ {{ row.termEnd || '-' }}</template>
          </el-table-column>
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button link type="danger" size="small" @click="removeMember(row)" v-if="row.status === 'active'">移除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </template>
    </el-dialog>

    <!-- 创建/编辑队伍 -->
    <el-dialog v-model="formVisible" :title="formTitle" width="560px" destroy-on-close aria-label="队伍表单">
      <el-form :model="form" label-width="100px">
        <el-form-item label="队伍名称" required>
          <el-input v-model="form.name" maxlength="128" />
        </el-form-item>
        <el-form-item label="队伍类型" required>
          <el-select v-model="form.teamType" class="full-select">
            <el-option v-for="opt in TEAM_TYPE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="编制人数">
          <el-input-number v-model="form.quota" :min="0" />
        </el-form-item>
        <el-form-item label="活动地点">
          <el-input v-model="form.location" maxlength="255" />
        </el-form-item>
        <el-form-item label="联系方式">
          <el-input v-model="form.contactInfo" maxlength="255" />
        </el-form-item>
        <el-form-item label="简介">
          <el-input v-model="form.description" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="formVisible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitForm">确定</el-button>
      </template>
    </el-dialog>

    <!-- 添加成员 -->
    <el-dialog v-model="addMemberVisible" title="纳新 - 添加成员" width="480px" destroy-on-close aria-label="添加成员">
      <el-form :model="memberForm" label-width="100px">
        <el-form-item label="用户ID" required>
          <el-input-number v-model="memberForm.userId" :min="1" class="full-select" />
        </el-form-item>
        <el-form-item label="姓名" required>
          <el-input v-model="memberForm.name" maxlength="64" />
        </el-form-item>
        <el-form-item label="学号" required>
          <el-input v-model="memberForm.studentNo" maxlength="32" />
        </el-form-item>
        <el-form-item label="角色" required>
          <el-select v-model="memberForm.role" class="full-select">
            <el-option v-for="opt in MEMBER_ROLE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
        <el-form-item label="入队日期" required>
          <el-input v-model="memberForm.joinDate" placeholder="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="届次开始">
          <el-input v-model="memberForm.termStart" placeholder="如 2025-09" />
        </el-form-item>
        <el-form-item label="届次结束">
          <el-input v-model="memberForm.termEnd" placeholder="如 2026-06" />
        </el-form-item>
        <el-form-item label="备注">
          <el-input v-model="memberForm.remark" type="textarea" :rows="2" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="addMemberVisible = false">取消</el-button>
        <el-button type="primary" @click="submitAddMember">确定</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.community-page {
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
.filter {
  background: var(--color-surface-sunken);
  padding: var(--space-3);
  border-radius: var(--radius-md);
}
.filter-select {
  width: 160px;
}
.full-select {
  width: 100%;
}
.pagination {
  align-self: flex-end;
}
.team-info {
  margin-bottom: var(--space-5);
}
.member-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: var(--space-3);
}
.member-header h4 {
  margin: 0;
}
</style>
