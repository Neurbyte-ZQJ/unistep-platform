<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { workstudyApi, jobStatusLabel, jobStatusType, JOB_STATUS_OPTIONS, type WorkStudyJob } from '../../api/workstudy'

const router = useRouter()
const loading = ref(false)
const list = ref<WorkStudyJob[]>([])
const total = ref(0)

const query = reactive({
  page: 1,
  size: 10,
  status: '',
  department: '',
  title: '',
})

async function fetchList() {
  loading.value = true
  try {
    const data = await workstudyApi.list(query)
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
  query.department = ''
  query.title = ''
  query.page = 1
  fetchList()
}

function goCreate() {
  router.push({ name: 'WorkStudyJobCreate' })
}
function goEdit(row: WorkStudyJob) {
  router.push({ name: 'WorkStudyJobEdit', params: { id: row.id } })
}
function goDetail(row: WorkStudyJob) {
  router.push({ name: 'WorkStudyJobDetail', params: { id: row.id } })
}

async function remove(row: WorkStudyJob) {
  try {
    await ElMessageBox.confirm(`确定删除岗位「${row.title}」吗？`, '提示', { type: 'warning' })
    await workstudyApi.remove(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (err) {
    if (err === 'cancel') return
    ElMessage.error((err as Error).message)
  }
}

async function publishJob(row: WorkStudyJob) {
  try {
    await workstudyApi.publish(row.id)
    ElMessage.success('已发布')
    fetchList()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function closeJob(row: WorkStudyJob) {
  try {
    await ElMessageBox.confirm(`确定关闭岗位「${row.title}」吗？`, '提示', { type: 'warning' })
    await workstudyApi.close(row.id)
    ElMessage.success('已关闭')
    fetchList()
  } catch (err) {
    if (err === 'cancel') return
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
  <div class="workstudy-page">
    <div class="toolbar">
      <h2>勤工助学</h2>
      <el-button type="primary" @click="goCreate">创建岗位</el-button>
    </div>

    <el-form :inline="true" :model="query" class="filter">
      <el-form-item label="岗位名称">
        <el-input v-model="query.title" placeholder="支持模糊搜索" clearable @keyup.enter="fetchList" />
      </el-form-item>
      <el-form-item label="部门">
        <el-input v-model="query.department" placeholder="支持模糊搜索" clearable @keyup.enter="fetchList" />
      </el-form-item>
      <el-form-item label="状态">
        <el-select v-model="query.status" placeholder="全部" clearable class="filter-select">
          <el-option v-for="opt in JOB_STATUS_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="fetchList">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-table :data="list" v-loading="loading" stripe border aria-label="岗位列表">
      <template #empty>
        <el-empty description="暂无岗位数据" />
      </template>
      <el-table-column prop="title" label="岗位名称" min-width="180" />
      <el-table-column prop="department" label="部门" width="140" />
      <el-table-column prop="location" label="地点" width="120" />
      <el-table-column prop="quota" label="名额" width="80" />
      <el-table-column label="时薪" width="120">
        <template #default="{ row }">¥{{ row.salaryPerHour.toFixed(2) }}/时</template>
      </el-table-column>
      <el-table-column label="开始时间" width="160">
        <template #default="{ row }">{{ formatTime(row.startTime) }}</template>
      </el-table-column>
      <el-table-column label="结束时间" width="160">
        <template #default="{ row }">{{ formatTime(row.endTime) }}</template>
      </el-table-column>
      <el-table-column label="状态" width="100">
        <template #default="{ row }">
          <el-tag :type="jobStatusType(row.status)">{{ jobStatusLabel(row.status) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="280" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row)">查看</el-button>
          <el-button link type="primary" @click="goEdit(row)" v-if="row.status === 'draft'">编辑</el-button>
          <el-button link type="warning" @click="publishJob(row)" v-if="row.status === 'draft'">发布</el-button>
          <el-button link type="warning" @click="closeJob(row)" v-if="row.status === 'published'">关闭</el-button>
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
.workstudy-page {
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
  margin: 0;
}
.filter-select {
  width: 160px;
}
.filter {
  background: var(--color-surface-sunken);
  padding: var(--space-3);
  border-radius: var(--radius-md);
}
.pagination {
  align-self: flex-end;
}
</style>
