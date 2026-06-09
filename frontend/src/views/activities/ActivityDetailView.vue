<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type UploadRequestOptions } from 'element-plus'
import {
  activityApi,
  activityStatusLabel,
  activityStatusType,
  ACTIVITY_STATUS_OPTIONS,
  type ClubActivity,
  type ActivityFile,
} from '../../api/activity'

const route = useRoute()
const router = useRouter()
const activity = ref<ClubActivity | null>(null)
const loading = ref(false)
const id = computed(() => Number(route.params.id))

// 审批对话框
const approvalDialog = reactive({ visible: false, opinion: '', approve: false })
// 签到对话框
const checkinDialog = reactive({ visible: false, studentId: 0 })
// 总结对话框
const summaryDialog = reactive({ visible: false, summary: '' })
// 状态变更
const statusDialog = reactive({ visible: false, status: '' })
// 上传类型
const uploadFileType = ref('image')
const submitting = ref(false)

async function load() {
  loading.value = true
  try {
    activity.value = await activityApi.detail(id.value)
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

function formatTime(t: string) {
  if (!t) return '-'
  return t.replace('T', ' ').substring(0, 16)
}

// 提交审批
async function submitForApproval() {
  try {
    await activityApi.submitForApproval(id.value)
    ElMessage.success('已提交审批')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

// 审批操作
function openApproval(approve: boolean) {
  approvalDialog.approve = approve
  approvalDialog.opinion = ''
  approvalDialog.visible = true
}

async function doApproval() {
  submitting.value = true
  try {
    await activityApi.approve(id.value, approvalDialog.opinion, approvalDialog.approve)
    ElMessage.success(approvalDialog.approve ? '已通过审批' : '已驳回')
    approvalDialog.visible = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    submitting.value = false
  }
}

// 报名
async function doRegister() {
  try {
    await activityApi.register(id.value)
    ElMessage.success('报名成功')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

// 取消报名
async function doCancelRegistration() {
  try {
    await ElMessageBox.confirm('确定取消报名吗？', '提示', { type: 'warning' })
    await activityApi.cancelRegistration(id.value)
    ElMessage.success('已取消报名')
    load()
  } catch (err) {
    if (err === 'cancel') return
    ElMessage.error((err as Error).message)
  }
}

// 签到
async function doCheckin() {
  submitting.value = true
  try {
    await activityApi.checkin(id.value, checkinDialog.studentId)
    ElMessage.success('签到成功')
    checkinDialog.visible = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    submitting.value = false
  }
}

// 上传文件
async function customUpload(options: UploadRequestOptions) {
  try {
    await activityApi.uploadFile(id.value, options.file as File, uploadFileType.value)
    ElMessage.success('文件已上传')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

// 提交总结
async function doSubmitSummary() {
  submitting.value = true
  try {
    await activityApi.submitSummary(id.value, summaryDialog.summary)
    ElMessage.success('总结已提交，活动已归档')
    summaryDialog.visible = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    submitting.value = false
  }
}

// 更新状态
async function doUpdateStatus() {
  submitting.value = true
  try {
    await activityApi.updateStatus(id.value, statusDialog.status)
    ElMessage.success('状态已更新')
    statusDialog.visible = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    submitting.value = false
  }
}

onMounted(load)
</script>

<template>
  <div v-loading="loading" class="activity-detail">
    <div class="header">
      <el-button link @click="router.back()">← 返回</el-button>
      <h2 v-if="activity">{{ activity.title }}</h2>
    </div>

    <template v-if="activity">
      <!-- 基本信息 -->
      <el-descriptions :column="3" border title="活动信息" aria-label="活动详情">
        <el-descriptions-item label="社团名称">{{ activity.clubName }}</el-descriptions-item>
        <el-descriptions-item label="活动名称">{{ activity.title }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="activityStatusType(activity.status)">{{ activityStatusLabel(activity.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ formatTime(activity.startTime) }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ formatTime(activity.endTime) }}</el-descriptions-item>
        <el-descriptions-item label="活动地点">{{ activity.location }}</el-descriptions-item>
        <el-descriptions-item label="容量">{{ activity.capacity }}</el-descriptions-item>
        <el-descriptions-item label="预算">{{ activity.budget != null ? `¥${activity.budget}` : '-' }}</el-descriptions-item>
        <el-descriptions-item label="审批意见">{{ activity.approvalOpinion || '-' }}</el-descriptions-item>
        <el-descriptions-item label="活动描述" :span="3">{{ activity.description }}</el-descriptions-item>
      </el-descriptions>

      <!-- 操作按钮 -->
      <el-card class="section">
        <template #header><span>操作</span></template>
        <div class="action-buttons">
          <el-button type="primary" @click="submitForApproval" v-if="activity.status === 'draft' || activity.status === 'rejected'">提交审批</el-button>
          <el-button type="success" @click="openApproval(true)" v-if="activity.status === 'pending'">通过审批</el-button>
          <el-button type="danger" @click="openApproval(false)" v-if="activity.status === 'pending'">驳回</el-button>
          <el-button type="success" @click="doRegister" v-if="activity.status === 'reg_open'">报名</el-button>
          <el-button type="warning" @click="doCancelRegistration" v-if="activity.status === 'reg_open'">取消报名</el-button>
          <el-button type="primary" @click="checkinDialog.visible = true" v-if="['reg_open', 'reg_closed', 'in_progress'].includes(activity.status)">签到</el-button>
          <el-button type="success" @click="summaryDialog.visible = true; summaryDialog.summary = activity.summary || ''" v-if="activity.status === 'completed'">提交总结</el-button>
          <el-button @click="statusDialog.visible = true; statusDialog.status = activity.status">变更状态</el-button>
          <el-button @click="router.push({ name: 'ActivityEdit', params: { id: activity.id } })" v-if="activity.status === 'draft' || activity.status === 'rejected'">编辑</el-button>
        </div>
      </el-card>

      <!-- 报名列表 -->
      <el-card class="section">
        <template #header><span>报名列表 ({{ activity.registrations?.length || 0 }})</span></template>
        <el-empty v-if="(activity.registrations || []).length === 0" description="暂无报名数据" />
        <el-table v-else :data="activity.registrations || []" border size="small">
          <el-table-column prop="studentId" label="学生ID" width="100" />
          <el-table-column prop="status" label="状态" width="120">
            <template #default="{ row }">
              <el-tag :type="row.status === 'registered' ? 'success' : 'info'" size="small">
                {{ row.status === 'registered' ? '已报名' : '已取消' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="报名时间" width="180">
            <template #default="{ row }">{{ formatTime(row.registeredAt) }}</template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 签到列表 -->
      <el-card class="section">
        <template #header><span>签到记录 ({{ activity.checkins?.length || 0 }})</span></template>
        <el-empty v-if="(activity.checkins || []).length === 0" description="暂无签到数据" />
        <el-table v-else :data="activity.checkins || []" border size="small">
          <el-table-column prop="studentId" label="学生ID" width="100" />
          <el-table-column label="签到时间" width="180">
            <template #default="{ row }">{{ formatTime(row.checkinTime) }}</template>
          </el-table-column>
          <el-table-column prop="checkinMethod" label="签到方式" width="120" />
        </el-table>
      </el-card>

      <!-- 活动图片/文件 -->
      <el-card class="section">
        <template #header>
          <div class="section-header">
            <span>活动图片/文件</span>
            <div class="upload-area">
              <el-select v-model="uploadFileType" size="small" class="status-tag">
                <el-option label="图片" value="image" />
                <el-option label="文档" value="document" />
                <el-option label="总结" value="summary" />
              </el-select>
              <el-upload :show-file-list="false" :http-request="customUpload">
                <el-button type="primary" plain size="small">上传文件</el-button>
              </el-upload>
            </div>
          </div>
        </template>
        <el-table :data="activity.files || []" border size="small">
          <el-table-column prop="fileName" label="文件名" />
          <el-table-column prop="fileType" label="类型" width="100" />
          <el-table-column prop="size" label="大小(B)" width="100" />
          <el-table-column label="上传时间" width="180">
            <template #default="{ row }">{{ formatTime(row.createdAt) }}</template>
          </el-table-column>
          <el-table-column label="访问" width="80">
            <template #default="{ row }">
              <el-link :href="row.url" target="_blank" type="primary">打开</el-link>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 活动总结 -->
      <el-card class="section" v-if="activity.summary">
        <template #header><span>活动总结</span></template>
        <div class="summary-content">{{ activity.summary }}</div>
      </el-card>
    </template>

    <!-- 审批对话框 -->
    <el-dialog v-model="approvalDialog.visible" :title="approvalDialog.approve ? '通过审批' : '驳回活动'" width="480px" aria-labelledby="dialog-approval-title">
      <el-form label-width="80px">
        <el-form-item label="审批意见">
          <el-input v-model="approvalDialog.opinion" type="textarea" :rows="3" placeholder="请输入审批意见" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="approvalDialog.visible = false">取消</el-button>
        <el-button :type="approvalDialog.approve ? 'success' : 'danger'" :loading="submitting" @click="doApproval">确认</el-button>
      </template>
    </el-dialog>

    <!-- 签到对话框 -->
    <el-dialog v-model="checkinDialog.visible" title="活动签到" width="400px" aria-labelledby="dialog-checkin-title">
      <el-form label-width="80px">
        <el-form-item label="学生ID">
          <el-input-number v-model="checkinDialog.studentId" :min="1" class="full-select" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="checkinDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="doCheckin">确认签到</el-button>
      </template>
    </el-dialog>

    <!-- 总结对话框 -->
    <el-dialog v-model="summaryDialog.visible" title="提交活动总结" width="560px" aria-labelledby="dialog-summary-title">
      <el-form label-width="80px">
        <el-form-item label="活动总结">
          <el-input v-model="summaryDialog.summary" type="textarea" :rows="6" placeholder="请输入活动总结" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="summaryDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="doSubmitSummary">提交并归档</el-button>
      </template>
    </el-dialog>

    <!-- 状态变更对话框 -->
    <el-dialog v-model="statusDialog.visible" title="变更活动状态" width="400px" aria-labelledby="dialog-status-title">
      <el-form label-width="80px">
        <el-form-item label="新状态">
          <el-select v-model="statusDialog.status" class="full-select">
            <el-option v-for="opt in ACTIVITY_STATUS_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="statusDialog.visible = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="doUpdateStatus">确认</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.activity-detail {
  display: flex;
  flex-direction: column;
  gap: var(--space-4);
}
.header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
}
.header h2 {
  margin: 0;
  font-size: var(--text-h2);
  color: var(--color-ink);
  font-weight: 600;
}
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
}
.upload-area {
  display: flex;
  align-items: center;
}
.action-buttons {
  display: flex;
  gap: var(--space-2);
  flex-wrap: wrap;
}
.summary-content {
  white-space: pre-wrap;
  line-height: 1.6;
}
.section :deep(.el-card__header) {
  background: var(--color-surface-sunken);
}
.status-tag {
  width: 120px;
  margin-right: var(--space-2);
}
.full-select {
  width: 100%;
}
</style>
