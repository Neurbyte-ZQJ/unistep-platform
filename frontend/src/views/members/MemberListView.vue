<script setup lang="ts">
import { onMounted, reactive, ref } from 'vue'
import { useRouter } from 'vue-router'
import { ElMessage, ElMessageBox } from 'element-plus'
import { memberApi, stageLabel, STAGE_OPTIONS, type MemberProfile } from '../../api/member'

const router = useRouter()
const loading = ref(false)
const list = ref<MemberProfile[]>([])
const total = ref(0)

const query = reactive({
  page: 1,
  size: 10,
  stage: '',
  name: '',
})

async function fetchList() {
  loading.value = true
  try {
    const data = await memberApi.list(query)
    list.value = data.items
    total.value = data.total
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function reset() {
  query.stage = ''
  query.name = ''
  query.page = 1
  fetchList()
}

function goCreate() {
  router.push({ name: 'MemberCreate' })
}
function goEdit(row: MemberProfile) {
  router.push({ name: 'MemberEdit', params: { id: row.id } })
}
function goDetail(row: MemberProfile) {
  router.push({ name: 'MemberDetail', params: { id: row.id } })
}

async function remove(row: MemberProfile) {
  try {
    await ElMessageBox.confirm(`确定删除 ${row.name} 的档案吗？`, '提示', { type: 'warning' })
    await memberApi.remove(row.id)
    ElMessage.success('删除成功')
    fetchList()
  } catch (err) {
    if (err === 'cancel') return
    ElMessage.error((err as Error).message)
  }
}

onMounted(fetchList)
</script>

<template>
  <div class="member-page">
    <div class="toolbar">
      <h2>团员发展</h2>
      <el-button type="primary" @click="goCreate">新增档案</el-button>
    </div>

    <el-form :inline="true" :model="query" class="filter">
      <el-form-item label="姓名">
        <el-input v-model="query.name" placeholder="支持模糊搜索" clearable @keyup.enter="fetchList" />
      </el-form-item>
      <el-form-item label="阶段">
        <el-select v-model="query.stage" placeholder="全部" clearable class="filter-select">
          <el-option v-for="opt in STAGE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
        </el-select>
      </el-form-item>
      <el-form-item>
        <el-button type="primary" @click="fetchList">查询</el-button>
        <el-button @click="reset">重置</el-button>
      </el-form-item>
    </el-form>

    <el-empty v-if="list.length === 0 && !loading" description="暂无团员数据" />
    <el-table v-else :data="list" v-loading="loading" stripe border aria-label="团员列表">
      <el-table-column prop="studentNo" label="学号" width="140" />
      <el-table-column prop="name" label="姓名" width="120" />
      <el-table-column prop="college" label="学院" />
      <el-table-column prop="major" label="专业" />
      <el-table-column prop="className" label="班级" />
      <el-table-column label="阶段" width="140">
        <template #default="{ row }">
          <el-tag>{{ stageLabel(row.stage) }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="240" fixed="right">
        <template #default="{ row }">
          <el-button link type="primary" @click="goDetail(row)">查看</el-button>
          <el-button link type="primary" @click="goEdit(row)">编辑</el-button>
          <el-button link type="danger" @click="remove(row)">删除</el-button>
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
.member-page {
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
  margin: 0;
  font-size: var(--text-h2);
  font-weight: 700;
  color: var(--color-ink);
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
