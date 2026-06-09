<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance } from 'element-plus'
import { activityApi, type ClubActivity } from '../../api/activity'

const route = useRoute()
const router = useRouter()
const formRef = ref<FormInstance>()
const submitting = ref(false)

const id = computed(() => {
  const raw = route.params.id
  return raw ? Number(raw) : undefined
})
const isEdit = computed(() => Boolean(id.value))

const form = reactive<Partial<ClubActivity>>({
  clubName: '',
  title: '',
  startTime: '',
  endTime: '',
  location: '',
  capacity: 50,
  description: '',
  budget: null,
})

const rules = {
  clubName: [{ required: true, message: '请输入社团名称', trigger: 'blur' }],
  title: [{ required: true, message: '请输入活动名称', trigger: 'blur' }],
  startTime: [{ required: true, message: '请选择开始时间', trigger: 'change' }],
  endTime: [{ required: true, message: '请选择结束时间', trigger: 'change' }],
  location: [{ required: true, message: '请输入活动地点', trigger: 'blur' }],
  capacity: [{ required: true, message: '请输入活动容量', trigger: 'blur' }],
  description: [{ required: true, message: '请输入活动描述', trigger: 'blur' }],
}

async function loadDetail() {
  if (!id.value) return
  try {
    const data = await activityApi.detail(id.value)
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
        await activityApi.update(id.value, form)
        ElMessage.success('已更新活动')
      } else {
        const created = await activityApi.create(form)
        ElMessage.success('已创建活动')
        router.replace({ name: 'ActivityEdit', params: { id: created.id } })
        return
      }
      router.push({ name: 'ActivityList' })
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
  <div class="activity-form">
    <div class="header">
      <el-button link @click="router.back()">← 返回</el-button>
      <h2>{{ isEdit ? '编辑活动' : '创建活动' }}</h2>
    </div>

    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" aria-label="编辑活动">
      <el-row :gutter="16">
        <el-col :xs="24" :sm="12">
          <el-form-item label="社团名称" prop="clubName">
            <el-input v-model="form.clubName" placeholder="请输入社团名称" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="活动名称" prop="title">
            <el-input v-model="form.title" placeholder="请输入活动名称" />
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
          <el-form-item label="活动地点" prop="location">
            <el-input v-model="form.location" placeholder="请输入活动地点" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="容量" prop="capacity">
            <el-input-number v-model="form.capacity" :min="1" class="full-select" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="预算(元)">
            <el-input-number v-model="form.budget" :precision="2" :min="0" class="full-select" />
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-form-item label="活动描述" prop="description">
            <el-input v-model="form.description" type="textarea" :rows="4" placeholder="请输入活动描述" />
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
.activity-form {
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
