<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, ElMessageBox, type UploadRequestOptions } from 'element-plus'
import {
  workstudyApi,
  jobStatusLabel,
  jobStatusType,
  applicationStatusLabel,
  applicationStatusType,
  salaryStatusLabel,
  salaryStatusType,
  JOB_STATUS_OPTIONS,
  APPLICATION_STATUS_OPTIONS,
  SALARY_STATUS_OPTIONS,
  type WorkStudyJob,
  type JobApplication,
  type WorkAttendance,
  type SalaryRecord,
} from '../../api/workstudy'

const route = useRoute()
const router = useRouter()
const job = ref<WorkStudyJob | null>(null)
const loading = ref(false)
const submitting = ref(false)
const id = computed(() => Number(route.params.id))

// 报名申请对话框
const applicationDialog = reactive({ visible: false, appId: 0, accept: true, remark: '' })
// 考勤对话框
const attendanceDialog = reactive({ visible: false, studentId: 0, date: '', method: 'manual' })
// 薪资计算对话框
const salaryDialog = reactive({ visible: false, month: '' })
// 上传文件类型
const uploadFileType = ref('document')

// 独立加载的关联数据
const applications = ref<JobApplication[]>([])
const attendances = ref<WorkAttendance[]>([])
const salaries = ref<SalaryRecord[]>([])

async function load() {
  loading.value = true
  try {
    job.value = await workstudyApi.detail(id.value)
    // 加载关联数据
    const [appData, attData, salData] = await Promise.all([
      workstudyApi.listApplications(id.value),
      workstudyApi.listAttendances(id.value),
      workstudyApi.listSalaries(id.value),
    ])
    applications.value = appData
    attendances.value = attData
    salaries.value = salData
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

// 发布岗位
async function publishJob() {
  try {
    await workstudyApi.publish(id.value)
    ElMessage.success('已发布')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

// 关闭岗位
async function closeJob() {
  try {
    await ElMessageBox.confirm('确定关闭该岗位吗？', '提示', { type: 'warning' })
    await workstudyApi.close(id.value)
    ElMessage.success('已关闭')
    load()
  } catch (err) {
    if (err === 'cancel') return
    ElMessage.error((err as Error).message)
  }
}

// 报名
async function doApply() {
  try {
    await workstudyApi.apply(id.value)
    ElMessage.success('报名成功')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

// 取消报名
async function doCancelApplication() {
  try {
    await ElMessageBox.confirm('确定取消报名吗？', '提示', { type: 'warning' })
    await workstudyApi.cancelApplication(id.value)
    ElMessage.success('已取消报名')
    load()
  } catch (err) {
    if (err === 'cancel') return
    ElMessage.error((err as Error).message)
  }
}

// 打开申请处理对话框
function openApplicationDialog(app: JobApplication, accept: boolean) {
  applicationDialog.appId = app.id
  applicationDialog.accept = accept
  applicationDialog.remark = ''
  applicationDialog.visible = true
}

// 处理申请
async function doApplicationAction() {
  try {
    if (applicationDialog.accept) {
      await workstudyApi.acceptApplication(id.value, applicationDialog.appId, applicationDialog.remark)
      ElMessage.success('已录用')
    } else {
      await workstudyApi.rejectApplication(id.value, applicationDialog.appId, applicationDialog.remark)
      ElMessage.success('已拒绝')
    }
    applicationDialog.visible = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

// 创建考勤
async function doCreateAttendance() {
  try {
    await workstudyApi.createAttendance(id.value, attendanceDialog.studentId, attendanceDialog.date, attendanceDialog.method)
    ElMessage.success('签到成功')
    attendanceDialog.visible = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

// 签退
async function doCheckout(att: WorkAttendance) {
  try {
    await workstudyApi.checkoutAttendance(att.id)
    ElMessage.success('签退成功')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

// 计算薪资
async function doCalculateSalary() {
  try {
    const result = await workstudyApi.calculateSalary(id.value, salaryDialog.month)
    ElMessage.success(`已计算 ${result.length} 条薪资记录`)
    salaryDialog.visible = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

// 发放薪资
async function doPaySalary(sal: SalaryRecord) {
  try {
    await ElMessageBox.confirm(`确定发放该薪资记录吗？金额：¥${sal.amount.toFixed(2)}`, '提示', { type: 'warning' })
    await workstudyApi.paySalary(sal.id)
    ElMessage.success('薪资已发放')
    load()
  } catch (err) {
    if (err === 'cancel') return
    ElMessage.error((err as Error).message)
  }
}

// 上传文件
async function customUpload(options: UploadRequestOptions) {
  try {
    await workstudyApi.uploadFile(id.value, options.file as File, uploadFileType.value)
    ElMessage.success('文件已上传')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

onMounted(load)
</script>

<template>
  <div v-loading="loading" class="workstudy-detail">
    <div class="header">
      <el-button link @click="router.back()">← 返回</el-button>
      <h2 v-if="job">{{ job.title }}</h2>
    </div>

    <template v-if="job">
      <!-- 基本信息 -->
      <el-descriptions :column="3" border title="岗位信息" aria-label="岗位信息">
        <el-descriptions-item label="岗位名称">{{ job.title }}</el-descriptions-item>
        <el-descriptions-item label="部门">{{ job.department }}</el-descriptions-item>
        <el-descriptions-item label="状态">
          <el-tag :type="jobStatusType(job.status)">{{ jobStatusLabel(job.status) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="工作地点">{{ job.location }}</el-descriptions-item>
        <el-descriptions-item label="名额">{{ job.quota }}</el-descriptions-item>
        <el-descriptions-item label="时薪">¥{{ job.salaryPerHour.toFixed(2) }}/时</el-descriptions-item>
        <el-descriptions-item label="开始时间">{{ formatTime(job.startTime) }}</el-descriptions-item>
        <el-descriptions-item label="结束时间">{{ formatTime(job.endTime) }}</el-descriptions-item>
        <el-descriptions-item label="联系人">{{ job.contactPerson }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ job.contactPhone }}</el-descriptions-item>
        <el-descriptions-item label="岗位描述" :span="3">{{ job.description }}</el-descriptions-item>
      </el-descriptions>

      <!-- 操作按钮 -->
      <el-card class="section">
        <template #header><span>操作</span></template>
        <div class="action-buttons">
          <el-button type="primary" @click="publishJob" v-if="job.status === 'draft'">发布</el-button>
          <el-button type="warning" @click="closeJob" v-if="job.status === 'published'">关闭</el-button>
          <el-button type="success" @click="doApply" v-if="job.status === 'published'">报名</el-button>
          <el-button type="warning" @click="doCancelApplication" v-if="job.status === 'published'">取消报名</el-button>
          <el-button @click="router.push({ name: 'WorkStudyJobEdit', params: { id: job.id } })" v-if="job.status === 'draft'">编辑</el-button>
        </div>
      </el-card>

      <!-- 报名列表 -->
      <el-card class="section">
        <template #header><span>报名列表 ({{ applications.length }})</span></template>
        <el-table :data="applications" border size="small">
          <template #empty>
            <el-empty description="暂无报名数据" />
          </template>
          <el-table-column prop="studentId" label="学生ID" width="100" />
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="applicationStatusType(row.status)" size="small">{{ applicationStatusLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="报名时间" width="180">
            <template #default="{ row }">{{ formatTime(row.appliedAt) }}</template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" min-width="120" />
          <el-table-column label="操作" width="160">
            <template #default="{ row }">
              <el-button link type="success" size="small" @click="openApplicationDialog(row, true)" v-if="row.status === 'applied'">录用</el-button>
              <el-button link type="danger" size="small" @click="openApplicationDialog(row, false)" v-if="row.status === 'applied'">拒绝</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 考勤记录 -->
      <el-card class="section">
        <template #header>
          <div class="section-header">
            <span>考勤记录 ({{ attendances.length }})</span>
            <el-button type="primary" plain size="small" @click="attendanceDialog.visible = true; attendanceDialog.studentId = 0; attendanceDialog.date = ''; attendanceDialog.method = 'manual'">新建签到</el-button>
          </div>
        </template>
        <el-table :data="attendances" border size="small">
          <template #empty>
            <el-empty description="暂无考勤记录" />
          </template>
          <el-table-column prop="studentId" label="学生ID" width="100" />
          <el-table-column prop="date" label="日期" width="120" />
          <el-table-column label="签到时间" width="180">
            <template #default="{ row }">{{ formatTime(row.checkinTime) }}</template>
          </el-table-column>
          <el-table-column label="签退时间" width="180">
            <template #default="{ row }">{{ row.checkoutTime ? formatTime(row.checkoutTime) : '-' }}</template>
          </el-table-column>
          <el-table-column prop="hours" label="工时(h)" width="100" />
          <el-table-column prop="method" label="方式" width="80" />
          <el-table-column prop="remark" label="备注" min-width="100" />
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button link type="primary" size="small" @click="doCheckout(row)" v-if="!row.checkoutTime">签退</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 薪资记录 -->
      <el-card class="section">
        <template #header>
          <div class="section-header">
            <span>薪资记录 ({{ salaries.length }})</span>
            <el-button type="primary" plain size="small" @click="salaryDialog.visible = true; salaryDialog.month = ''">计算薪资</el-button>
          </div>
        </template>
        <el-table :data="salaries" border size="small">
          <template #empty>
            <el-empty description="暂无薪资记录" />
          </template>
          <el-table-column prop="studentId" label="学生ID" width="100" />
          <el-table-column prop="month" label="月份" width="100" />
          <el-table-column prop="hours" label="工时(h)" width="100" />
          <el-table-column label="金额" width="120">
            <template #default="{ row }">¥{{ row.amount.toFixed(2) }}</template>
          </el-table-column>
          <el-table-column label="状态" width="100">
            <template #default="{ row }">
              <el-tag :type="salaryStatusType(row.status)" size="small">{{ salaryStatusLabel(row.status) }}</el-tag>
            </template>
          </el-table-column>
          <el-table-column label="发放时间" width="180">
            <template #default="{ row }">{{ row.paidAt ? formatTime(row.paidAt) : '-' }}</template>
          </el-table-column>
          <el-table-column prop="remark" label="备注" min-width="100" />
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button link type="success" size="small" @click="doPaySalary(row)" v-if="row.status === 'pending'">发放</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-card>

      <!-- 文件列表 -->
      <el-card class="section">
        <template #header>
          <div class="section-header">
            <span>相关文件</span>
            <div class="upload-area">
              <el-select v-model="uploadFileType" size="small" class="status-tag">
                <el-option label="图片" value="image" />
                <el-option label="文档" value="document" />
                <el-option label="其他" value="other" />
              </el-select>
              <el-upload :show-file-list="false" :http-request="customUpload">
                <el-button type="primary" plain size="small">上传文件</el-button>
              </el-upload>
            </div>
          </div>
        </template>
        <el-table :data="job.files || []" border size="small">
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
    </template>

    <!-- 申请处理对话框 -->
    <el-dialog v-model="applicationDialog.visible" :title="applicationDialog.accept ? '录用申请' : '拒绝申请'" width="480px" aria-label="申请处理">
      <el-form label-width="80px">
        <el-form-item label="备注">
          <el-input v-model="applicationDialog.remark" type="textarea" :rows="3" placeholder="请输入备注" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="applicationDialog.visible = false">取消</el-button>
        <el-button :type="applicationDialog.accept ? 'success' : 'danger'" @click="doApplicationAction" :loading="submitting">确认</el-button>
      </template>
    </el-dialog>

    <!-- 考勤对话框 -->
    <el-dialog v-model="attendanceDialog.visible" title="新建签到" width="480px" aria-label="新建签到">
      <el-form label-width="80px">
        <el-form-item label="学生ID">
          <el-input-number v-model="attendanceDialog.studentId" :min="1" class="full-select" />
        </el-form-item>
        <el-form-item label="日期">
          <el-date-picker v-model="attendanceDialog.date" type="date" placeholder="选择日期" value-format="YYYY-MM-DD" class="full-select" />
        </el-form-item>
        <el-form-item label="签到方式">
          <el-select v-model="attendanceDialog.method" class="full-select">
            <el-option label="手动" value="manual" />
            <el-option label="扫码" value="qrcode" />
            <el-option label="人脸" value="face" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="attendanceDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doCreateAttendance" :loading="submitting">确认签到</el-button>
      </template>
    </el-dialog>

    <!-- 薪资计算对话框 -->
    <el-dialog v-model="salaryDialog.visible" title="计算薪资" width="400px" aria-label="计算薪资">
      <el-form label-width="80px">
        <el-form-item label="月份">
          <el-date-picker v-model="salaryDialog.month" type="month" placeholder="选择月份" value-format="YYYY-MM" class="full-select" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="salaryDialog.visible = false">取消</el-button>
        <el-button type="primary" @click="doCalculateSalary" :loading="submitting">计算</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.workstudy-detail {
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
.status-tag {
  width: 120px;
  margin-right: var(--space-2);
}
.full-select {
  width: 100%;
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
.section :deep(.el-card__header) {
  background: var(--color-surface-sunken);
}
</style>
