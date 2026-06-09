<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance } from 'element-plus'
import { workstudyApi, type WorkStudyJob } from '../../api/workstudy'

const route = useRoute()
const router = useRouter()
const formRef = ref<FormInstance>()
const submitting = ref(false)

const id = computed(() => {
  const raw = route.params.id
  return raw ? Number(raw) : undefined
})
const isEdit = computed(() => Boolean(id.value))

const form = reactive<Partial<WorkStudyJob>>({
  title: '',
  department: '',
  location: '',
  description: '',
  quota: 1,
  salaryPerHour: 0,
  startTime: '',
  endTime: '',
  contactPerson: '',
  contactPhone: '',
})

const rules = {
  title: [{ required: true, message: '请输入岗位名称', trigger: 'blur' }],
  department: [{ required: true, message: '请输入部门', trigger: 'blur' }],
  location: [{ required: true, message: '请输入工作地点', trigger: 'blur' }],
  description: [{ required: true, message: '请输入岗位描述', trigger: 'blur' }],
  quota: [{ required: true, message: '请输入招聘名额', trigger: 'blur' }],
  salaryPerHour: [{ required: true, message: '请输入时薪', trigger: 'blur' }],
  startTime: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  endTime: [{ required: true, message: '请选择结束时间', trigger: 'change' }],
  contactPerson: [{ required: true, message: '请输入联系人', trigger: 'blur' }],
  contactPhone: [{ required: true, message: '请输入联系电话', trigger: 'blur' }],
}

async function loadDetail() {
  if (!id.value) return
  try {
    const data = await workstudyApi.detail(id.value)
    Object.assign(form, data)
  } catch (err) {
    ElMessage.error((err as Error).message)
  }
}

async function submit() {
  if (!formRef.value) return
  await formRef.value.validate(async (valid) => {
    if (!valid) return
    submitting.value = true
    try {
      if (isEdit.value && id.value) {
        await workstudyApi.update(id.value, form)
        ElMessage.success('已更新岗位')
      } else {
        const created = await workstudyApi.create(form)
        ElMessage.success('已创建岗位')
        router.replace({ name: 'WorkStudyJobEdit', params: { id: created.id } })
        return
      }
      router.push({ name: 'WorkStudyJobList' })
    } catch (err) {
      ElMessage.error((err as Error).message)
    } finally {
      submitting.value = false
    }
  })
}

onMounted(loadDetail)
</script>

<template>
  <div class="workstudy-form">
    <div class="header">
      <el-button link @click="router.back()">← 返回</el-button>
      <h2>{{ isEdit ? '编辑岗位' : '创建岗位' }}</h2>
    </div>

    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" aria-label="编辑岗位">
      <el-row :gutter="16">
        <el-col :xs="24" :sm="12">
          <el-form-item label="岗位名称" prop="title">
            <el-input v-model="form.title" placeholder="请输入岗位名称" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="部门" prop="department">
            <el-input v-model="form.department" placeholder="请输入部门" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="工作地点" prop="location">
            <el-input v-model="form.location" placeholder="请输入工作地点" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="6">
          <el-form-item label="名额" prop="quota">
            <el-input-number v-model="form.quota" :min="1" class="full-select" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="6">
          <el-form-item label="时薪(元)" prop="salaryPerHour">
            <el-input-number v-model="form.salaryPerHour" :precision="2" :min="0" class="full-select" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="开始时间" prop="startTime">
            <el-date-picker v-model="form.startTime" type="datetime" placeholder="选择开始时间" value-format="YYYY-MM-DDTHH:mm:ssZ" class="full-select" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="结束时间" prop="endTime">
            <el-date-picker v-model="form.endTime" type="datetime" placeholder="选择结束时间" value-format="YYYY-MM-DDTHH:mm:ssZ" class="full-select" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="联系人" prop="contactPerson">
            <el-input v-model="form.contactPerson" placeholder="请输入联系人" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="联系电话" prop="contactPhone">
            <el-input v-model="form.contactPhone" placeholder="请输入联系电话" />
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-form-item label="岗位描述" prop="description">
            <el-input v-model="form.description" type="textarea" :rows="4" placeholder="请输入岗位描述" />
          </el-form-item>
        </el-col>
      </el-row>

      <el-form-item>
        <el-button type="primary" :loading="submitting" @click="submit">保存</el-button>
        <el-button @click="router.back()">取消</el-button>
      </el-form-item>
    </el-form>
  </div>
</template>

<style scoped>
.workstudy-form {
  background: var(--color-surface-raised);
  padding: var(--space-6);
  border-radius: var(--radius-lg);
}
.header {
  display: flex;
  align-items: center;
  gap: var(--space-3);
  margin-bottom: var(--space-4);
}
.header h2 {
  margin: 0;
  font-size: var(--text-h2);
  color: var(--color-ink);
  font-weight: 600;
}
.full-select {
  width: 100%;
}
</style>
