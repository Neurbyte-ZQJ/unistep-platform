<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type UploadRequestOptions } from 'element-plus'
import {
  memberApi,
  stageLabel,
  type MemberArchive,
  type LeagueApplication,
  type ActivistRecord,
  type DevelopTargetRecord,
  type PoliticalReview,
} from '../../api/member'

const route = useRoute()
const router = useRouter()
const archive = ref<MemberArchive | null>(null)
const loading = ref(false)

const id = computed(() => Number(route.params.id))

const dialog = reactive({
  application: false,
  activist: false,
  develop: false,
  political: false,
})

const application = reactive<LeagueApplication>({ applyDate: '', motivation: '', introducer: '' })
const activist = reactive<ActivistRecord>({ startDate: '', trainer: '', trainPlan: '', score: 0 })
const develop = reactive<DevelopTargetRecord>({ confirmedDate: '', mentor: '', publicityNote: '', conclusion: '' })
const political = reactive<PoliticalReview>({ reviewDate: '', reviewer: '', familyMembers: '', conclusion: '' })
const uploadCategory = ref('application')
const submitting = ref(false)

async function load() {
  loading.value = true
  try {
    archive.value = await memberApi.archive(id.value)
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    loading.value = false
  }
}

async function submitApplication() {
  submitting.value = true
  try {
    await memberApi.addApplication(id.value, { ...application })
    ElMessage.success('已提交入团申请')
    dialog.application = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    submitting.value = false
  }
}
async function submitActivist() {
  submitting.value = true
  try {
    await memberApi.addActivist(id.value, { ...activist })
    ElMessage.success('积极分子记录已保存')
    dialog.activist = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    submitting.value = false
  }
}
async function submitDevelop() {
  submitting.value = true
  try {
    await memberApi.addDevelop(id.value, { ...develop })
    ElMessage.success('发展对象记录已保存')
    dialog.develop = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    submitting.value = false
  }
}
async function submitPolitical() {
  submitting.value = true
  try {
    await memberApi.addPoliticalReview(id.value, { ...political })
    ElMessage.success('政审备案已提交')
    dialog.political = false
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  } finally {
    submitting.value = false
  }
}

async function customUpload(options: UploadRequestOptions) {
  try {
    await memberApi.uploadAttachment(id.value, options.file as File, uploadCategory.value)
    ElMessage.success('附件已上传')
    load()
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

onMounted(load)
</script>

<template>
  <div v-loading="loading" class="member-detail">
    <div class="header">
      <el-button link @click="router.back()">← 返回</el-button>
      <h2 v-if="archive">{{ archive.profile.name }} 的电子档案</h2>
    </div>

    <template v-if="archive">
      <el-descriptions :column="3" border title="基本信息" aria-label="团员详情">
        <el-descriptions-item label="学号">{{ archive.profile.studentNo }}</el-descriptions-item>
        <el-descriptions-item label="性别">{{ archive.profile.gender || '-' }}</el-descriptions-item>
        <el-descriptions-item label="阶段">
          <el-tag>{{ stageLabel(archive.profile.stage) }}</el-tag>
        </el-descriptions-item>
        <el-descriptions-item label="学院">{{ archive.profile.college || '-' }}</el-descriptions-item>
        <el-descriptions-item label="专业">{{ archive.profile.major || '-' }}</el-descriptions-item>
        <el-descriptions-item label="班级">{{ archive.profile.className || '-' }}</el-descriptions-item>
        <el-descriptions-item label="联系电话">{{ archive.profile.phone || '-' }}</el-descriptions-item>
        <el-descriptions-item label="入团日期">{{ archive.profile.joinDate || '-' }}</el-descriptions-item>
        <el-descriptions-item label="身份证号">{{ archive.profile.idCard || '-' }}</el-descriptions-item>
      </el-descriptions>

      <el-card class="section">
        <template #header>
          <div class="section-header">
            <span>团员发展流程</span>
            <div>
              <el-button type="primary" plain @click="dialog.application = true">入团申请</el-button>
              <el-button type="success" plain @click="dialog.activist = true">积极分子培养</el-button>
              <el-button type="warning" plain @click="dialog.develop = true">发展对象</el-button>
              <el-button type="danger" plain @click="dialog.political = true">政审备案</el-button>
            </div>
          </div>
        </template>
        <el-timeline>
          <el-timeline-item
            v-for="(item, idx) in archive.timeline"
            :key="idx"
            :timestamp="`${item.date}  ${item.stage}`"
            placement="top"
          >
            {{ item.text }}
          </el-timeline-item>
          <el-timeline-item v-if="archive.timeline.length === 0" timestamp="暂无记录">
            请点击上方按钮录入入团申请等流程记录。
          </el-timeline-item>
        </el-timeline>
      </el-card>

      <el-card class="section">
        <template #header>
          <div class="section-header">
            <span>档案附件（MinIO）</span>
            <el-select v-model="uploadCategory" size="small" class="filter-select">
              <el-option label="入团申请" value="application" />
              <el-option label="积极分子" value="activist" />
              <el-option label="发展对象" value="develop" />
              <el-option label="政审" value="political" />
              <el-option label="其他" value="other" />
            </el-select>
            <el-upload :show-file-list="false" :http-request="customUpload">
              <el-button type="primary" plain>上传附件</el-button>
            </el-upload>
          </div>
        </template>
        <el-table :data="archive.profile.attachments || []" border>
          <el-table-column prop="fileName" label="文件名" />
          <el-table-column prop="category" label="类别" width="120" />
          <el-table-column prop="size" label="大小(B)" width="120" />
          <el-table-column prop="createdAt" label="上传时间" width="180" />
          <el-table-column label="访问" width="120">
            <template #default="{ row }">
              <el-link :href="row.url" target="_blank" type="primary">打开</el-link>
            </template>
          </el-table-column>
        </el-table>
      </el-card>
    </template>

    <!-- 入团申请 -->
    <el-dialog v-model="dialog.application" title="提交入团申请" width="480px" aria-labelledby="dialog-application-title">
      <el-form :model="application" label-width="100px">
        <el-form-item label="申请日期">
          <el-input v-model="application.applyDate" placeholder="YYYY-MM-DD" />
        </el-form-item>
        <el-form-item label="介绍人">
          <el-input v-model="application.introducer" />
        </el-form-item>
        <el-form-item label="入团动机">
          <el-input v-model="application.motivation" type="textarea" :rows="3" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.application = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitApplication">提交</el-button>
      </template>
    </el-dialog>

    <!-- 积极分子 -->
    <el-dialog v-model="dialog.activist" title="积极分子培养" width="480px" aria-labelledby="dialog-activist-title">
      <el-form :model="activist" label-width="100px">
        <el-form-item label="开始日期"><el-input v-model="activist.startDate" placeholder="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="培养联系人"><el-input v-model="activist.trainer" /></el-form-item>
        <el-form-item label="培养计划"><el-input v-model="activist.trainPlan" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="评分"><el-input-number v-model="activist.score" :precision="1" :min="0" :max="100" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.activist = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitActivist">保存</el-button>
      </template>
    </el-dialog>

    <!-- 发展对象 -->
    <el-dialog v-model="dialog.develop" title="发展对象" width="480px" aria-labelledby="dialog-develop-title">
      <el-form :model="develop" label-width="100px">
        <el-form-item label="确定日期"><el-input v-model="develop.confirmedDate" placeholder="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="联系导师"><el-input v-model="develop.mentor" /></el-form-item>
        <el-form-item label="公示说明"><el-input v-model="develop.publicityNote" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="结论"><el-input v-model="develop.conclusion" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.develop = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitDevelop">保存</el-button>
      </template>
    </el-dialog>

    <!-- 政审 -->
    <el-dialog v-model="dialog.political" title="政审备案" width="480px" aria-labelledby="dialog-political-title">
      <el-form :model="political" label-width="100px">
        <el-form-item label="政审日期"><el-input v-model="political.reviewDate" placeholder="YYYY-MM-DD" /></el-form-item>
        <el-form-item label="审核人"><el-input v-model="political.reviewer" /></el-form-item>
        <el-form-item label="直系亲属"><el-input v-model="political.familyMembers" type="textarea" :rows="3" /></el-form-item>
        <el-form-item label="结论"><el-input v-model="political.conclusion" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialog.political = false">取消</el-button>
        <el-button type="primary" :loading="submitting" @click="submitPolitical">提交</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<style scoped>
.member-detail {
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
  font-weight: 700;
  color: var(--color-ink);
}
.section-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: var(--space-3);
}
.section :deep(.el-card__header) {
  background: var(--color-surface-sunken);
}
.filter-select {
  width: 160px;
}
</style>
