<script setup lang="ts">
import { computed, onMounted, reactive, ref } from 'vue'
import { useRoute, useRouter } from 'vue-router'
import { ElMessage, type FormInstance } from 'element-plus'
import { memberApi, STAGE_OPTIONS, type MemberProfile } from '../../api/member'

const route = useRoute()
const router = useRouter()
const formRef = ref<FormInstance>()
const submitting = ref(false)

const id = computed(() => {
  const raw = route.params.id
  return raw ? Number(raw) : undefined
})
const isEdit = computed(() => Boolean(id.value))

const form = reactive<Partial<MemberProfile>>({
  name: '',
  studentNo: '',
  gender: '男',
  birthday: '',
  idCard: '',
  nation: '',
  phone: '',
  college: '',
  major: '',
  className: '',
  stage: 'applicant',
  joinDate: '',
  remark: '',
})

const rules = {
  name: [{ required: true, message: '请输入姓名', trigger: 'blur' }],
  studentNo: [{ required: true, message: '请输入学号', trigger: 'blur' }],
}

async function loadDetail() {
  if (!id.value) return
  try {
    const data = await memberApi.detail(id.value)
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
        await memberApi.update(id.value, form)
        ElMessage.success('已更新档案')
      } else {
        const created = await memberApi.create(form)
        ElMessage.success('已创建档案')
        router.replace({ name: 'MemberEdit', params: { id: created.id } })
        return
      }
      router.push({ name: 'MemberList' })
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
  <div class="member-form">
    <div class="header">
      <el-button link @click="router.back()">← 返回</el-button>
      <h2>{{ isEdit ? '编辑团员档案' : '新增团员档案' }}</h2>
    </div>

    <el-form ref="formRef" :model="form" :rules="rules" label-width="100px" aria-label="编辑团员">
      <el-row :gutter="16">
        <el-col :xs="24" :sm="12">
          <el-form-item label="姓名" prop="name">
            <el-input v-model="form.name" placeholder="请输入姓名" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="学号" prop="studentNo">
            <el-input v-model="form.studentNo" placeholder="请输入学号" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="性别">
            <el-radio-group v-model="form.gender">
              <el-radio label="男">男</el-radio>
              <el-radio label="女">女</el-radio>
            </el-radio-group>
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="生日">
            <el-input v-model="form.birthday" placeholder="YYYY-MM-DD" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="身份证号">
            <el-input v-model="form.idCard" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="民族">
            <el-input v-model="form.nation" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="联系电话">
            <el-input v-model="form.phone" />
          </el-form-item>
        </el-col>
        <el-col :xs="24" :sm="12">
          <el-form-item label="入团日期">
            <el-input v-model="form.joinDate" placeholder="YYYY-MM-DD" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="学院">
            <el-input v-model="form.college" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="专业">
            <el-input v-model="form.major" />
          </el-form-item>
        </el-col>
        <el-col :span="8">
          <el-form-item label="班级">
            <el-input v-model="form.className" />
          </el-form-item>
        </el-col>
        <el-col :span="12">
          <el-form-item label="当前阶段">
            <el-select v-model="form.stage" class="full-select">
              <el-option v-for="opt in STAGE_OPTIONS" :key="opt.value" :label="opt.label" :value="opt.value" />
            </el-select>
          </el-form-item>
        </el-col>
        <el-col :span="24">
          <el-form-item label="备注">
            <el-input v-model="form.remark" type="textarea" :rows="3" />
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
.member-form {
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
  font-weight: 700;
  color: var(--color-ink);
}
.full-select {
  width: 100%;
}
</style>
