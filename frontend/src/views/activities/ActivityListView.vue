<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { activityApi, activityStatusLabel, activityStatusType, ACTIVITY_STATUS_OPTIONS, type ClubActivity } from '../../api/activity'

const router = useRouter()
const loading = ref(false)
const list = ref<ClubActivity[]>([])
const total = ref(0)

const query = reactive({
  page: 1,
  size: 10,
  status: '',
  clubName: '',
  title: '',
})

async function fetchList() {
  loading.value = true
  try {
    const data = await activityApi.list(query)
    list.value = data.items
    total.value = data.total
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function reset() {
  query.status = ''
  query.clubName = ''
  query.title = ''
  query.page = 1
  fetchList()
}

function goCreate() {
  router.push({ name: 'ActivityCreate' })
}
function goEdit(row: ClubActivity) {
  router.push({ name: 'ActivityEdit', params: { id: row.id } })
}
function goDetail(row: ClubActivity) {
  router.push({ name: 'ActivityDetail', params: { id: row.id } })
}

async function remove(row: ClubActivity) {
  try {
    await ElMessageBox.confirm(`确定删除活动「${row.title}」吗？`, '提示', { type: 'warning' })
    await activityApi.remove(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (err) {
    if (err === 'cancel') return
    ElMessage.error((err as Error).message)
  }
}

async function submitApproval(row: ClubActivity) {
  try {
    await activityApi.submitForApproval(row.id)
    ElMessage.success('已提交审批')
    fetchList()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

function formatTime(t: string) {
  if (!t) return '-'
  return t.replace('T', ' ').substring(0, 16)
}

onMounted(fetchList)
</script>

<template>
  <div class="activity-page">
    <div class="toolbar">
      <h2>社团活动</h2>
      <el-button type="primary" @click="goCreate">创建活动</el-button>
    </div>

    <el-form :inline="true" :model="query" class="filter">
      <el-form-item label="活动名称">
        <el-input v-model="query.title" placeholder="支持模糊搜索" clearable @keyup.enter="fetchList" />
      </el-form-item>
      <el-form-item label="社团名称">
        <el-input v-model="query.clubName" placeholder="支持模糊搜索" clearable @keyup.enter="fetchList" />
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" placeholder="全部" clearable class="filter-select">
          <el-option v-for="opt in ACTIVITY_STATUS_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="fetchList">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-empty v-if="list.length === 0 && !loading" description="暂无活动数据" />
    <el-table v-else :data="list" v-loading="loading" stripe border aria-label="活动列表">
      <el-table-column prop="title" label="活动名称" min-width="180" />
      <el-table-column prop="clubName" label="社团" width="140" />
      <el-table-column label="开始时间" width="160">
        <template #default="{ row }">{{ formatTime(row.startTime) }}</template>
      </el-table-column>
      <el-table-column label="结束时间" width="160">
        <template #default="{ row }">{{ formatTime(row.endTime) }}</template>
      </el-table-column>
      <el-table-column prop="location" label="地点" width="140" />
      <el-table-column prop="capacity" label="容量" width="80" />
      <el-table-column label="状态" width="120">
        <template #default="{ row }">
          <el-tag :type="activityStatusType(row.status)">{{ activityStatusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row)">查看</el-button>
          <el-button link type="primary" @click="goEdit(row)" v-if="row.status === 'draft' || row.status === 'rejected'">编辑</el-button>
          <el-button link type="warning" @click="submitApproval(row)" v-if="row.status === 'draft' || row.status === 'rejected'">提交审批</el-button>
          <el-button link type="danger" @click="remove(row)" v-if="row.status === 'draft'">删除</el-button>
        </template>
      </el-table-column>
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
  </div>
</template>

<style scoped>
.activity-page {
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
.pagination {
  align-self: flex-end;
}
.filter-select {
  width: 160px;
}
</style>
