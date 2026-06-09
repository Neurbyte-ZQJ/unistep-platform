import { useAuthStore } from '../stores/auth'

interface ApiResponse<T> {
  code: string
  message: string
  data: T
}

async function request<T>(url: string, init: RequestInit = {}): Promise<T> {
  const auth = useAuthStore()
  const headers = new Headers(init.headers)
  if (!headers.has('Content-Type') && !(init.body instanceof FormData)) {
    headers.set('Content-Type', 'application/json')
  }
  if (auth.token) headers.set('Authorization', `Bearer ${auth.token}`)

  const res = await fetch(url, { ...init, headers })
  const text = await res.text()
  const body = text ? (JSON.parse(text) as ApiResponse<T>) : ({} as ApiResponse<T>)
  if (!res.ok) {
    throw new Error(body.message || `请求失败 (${res.status})`)
  }
  return body.data
}

// ---------- 类型定义 ----------

export interface WorkStudyJob {
  id: number
  title: string
  department: string
  location: string
  description: string
  quota: number
  salaryPerHour: number
  startTime: string
  endTime: string
  contactPerson: string
  contactPhone: string
  status: string
  createdBy: number
  createdAt: string
  updatedAt: string
  applications?: JobApplication[]
  attendances?: WorkAttendance[]
  salaries?: SalaryRecord[]
  files?: WorkStudyFile[]
}

export interface JobApplication {
  id: number
  jobId: number
  studentId: number
  status: string
  remark: string
  appliedAt: string
  acceptedAt?: string | null
  rejectedAt?: string | null
  cancelledAt?: string | null
}

export interface WorkAttendance {
  id: number
  jobId: number
  studentId: number
  date: string
  checkinTime: string
  checkoutTime?: string | null
  hours: number
  method: string
  remark: string
}

export interface SalaryRecord {
  id: number
  jobId: number
  studentId: number
  month: string
  hours: number
  amount: number
  status: string
  paidAt?: string | null
  remark: string
}

export interface WorkStudyFile {
  id: number
  jobId: number
  fileName: string
  objectKey: string
  url: string
  fileType: string
  size: number
  uploadedBy: number
  createdAt: string
}

export interface JobListResponse {
  items: WorkStudyJob[]
  total: number
  page: number
  size: number
}

export interface WorkStudyStatistics {
  totalJobs: number
  statusBreakdown: { status: string; count: number }[]
  totalApplications: number
  totalAccepted: number
  totalSalaryPaid: number
  recentJobs: WorkStudyJob[]
}

// ---------- API ----------

export const workstudyApi = {
  list(params: { page?: number; size?: number; status?: string; department?: string; title?: string } = {}) {
    const search = new URLSearchParams()
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== '') search.set(k, String(v))
    })
    return request<JobListResponse>(`/api/v1/workstudy/jobs?${search.toString()}`)
  },
  create(payload: Partial<WorkStudyJob>) {
    return request<WorkStudyJob>('/api/v1/workstudy/jobs', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  detail(id: number) {
    return request<WorkStudyJob>(`/api/v1/workstudy/jobs/${id}`)
  },
  update(id: number, payload: Partial<WorkStudyJob>) {
    return request<WorkStudyJob>(`/api/v1/workstudy/jobs/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    })
  },
  remove(id: number) {
    return request<{ id: number }>(`/api/v1/workstudy/jobs/${id}`, { method: 'DELETE' })
  },
  publish(id: number) {
    return request<WorkStudyJob>(`/api/v1/workstudy/jobs/${id}/publish`, { method: 'POST' })
  },
  close(id: number) {
    return request<WorkStudyJob>(`/api/v1/workstudy/jobs/${id}/close`, { method: 'POST' })
  },
  apply(id: number) {
    return request<JobApplication>(`/api/v1/workstudy/jobs/${id}/apply`, { method: 'POST' })
  },
  cancelApplication(id: number) {
    return request<{ jobId: number }>(`/api/v1/workstudy/jobs/${id}/cancel-application`, { method: 'POST' })
  },
  listApplications(id: number, status?: string) {
    const params = status ? `?status=${status}` : ''
    return request<JobApplication[]>(`/api/v1/workstudy/jobs/${id}/applications${params}`)
  },
  acceptApplication(jobId: number, appId: number, remark?: string) {
    return request<JobApplication>(`/api/v1/workstudy/jobs/${jobId}/applications/${appId}/accept`, {
      method: 'POST',
      body: JSON.stringify({ remark }),
    })
  },
  rejectApplication(jobId: number, appId: number, remark?: string) {
    return request<JobApplication>(`/api/v1/workstudy/jobs/${jobId}/applications/${appId}/reject`, {
      method: 'POST',
      body: JSON.stringify({ remark }),
    })
  },
  createAttendance(jobId: number, studentId: number, date: string, method?: string) {
    return request<WorkAttendance>(`/api/v1/workstudy/jobs/${jobId}/attendances`, {
      method: 'POST',
      body: JSON.stringify({ studentId, date, method: method || 'manual' }),
    })
  },
  listAttendances(jobId: number, params?: { studentId?: number; date?: string }) {
    const search = new URLSearchParams()
    if (params) {
      Object.entries(params).forEach(([k, v]) => {
        if (v !== undefined && v !== '') search.set(k, String(v))
      })
    }
    const qs = search.toString()
    return request<WorkAttendance[]>(`/api/v1/workstudy/jobs/${jobId}/attendances${qs ? '?' + qs : ''}`)
  },
  checkoutAttendance(attId: number) {
    return request<WorkAttendance>(`/api/v1/workstudy/attendances/${attId}/checkout`, {
      method: 'PUT',
    })
  },
  calculateSalary(jobId: number, month: string) {
    return request<SalaryRecord[]>(`/api/v1/workstudy/jobs/${jobId}/salary/calculate`, {
      method: 'POST',
      body: JSON.stringify({ month }),
    })
  },
  listSalaries(jobId: number, params?: { month?: string; status?: string; studentId?: number }) {
    const search = new URLSearchParams()
    if (params) {
      Object.entries(params).forEach(([k, v]) => {
        if (v !== undefined && v !== '') search.set(k, String(v))
      })
    }
    const qs = search.toString()
    return request<SalaryRecord[]>(`/api/v1/workstudy/jobs/${jobId}/salary${qs ? '?' + qs : ''}`)
  },
  paySalary(salaryId: number) {
    return request<SalaryRecord>(`/api/v1/workstudy/salary/${salaryId}/pay`, {
      method: 'PUT',
    })
  },
  uploadFile(jobId: number, file: File, fileType: string) {
    const form = new FormData()
    form.append('file', file)
    form.append('fileType', fileType)
    return request<WorkStudyFile>(`/api/v1/workstudy/jobs/${jobId}/files`, {
      method: 'POST',
      body: form,
    })
  },
  statistics() {
    return request<WorkStudyStatistics>('/api/v1/workstudy/statistics')
  },
}

// ---------- 常量 ----------

export const JOB_STATUS_OPTIONS = [
  { value: 'draft', label: '草稿', type: 'info' },
  { value: 'published', label: '已发布', type: 'success' },
  { value: 'closed', label: '已关闭', type: 'warning' },
  { value: 'completed', label: '已完成', type: '' },
]

export const APPLICATION_STATUS_OPTIONS = [
  { value: 'applied', label: '已报名', type: 'warning' },
  { value: 'accepted', label: '已录用', type: 'success' },
  { value: 'rejected', label: '已拒绝', type: 'danger' },
  { value: 'cancelled', label: '已取消', type: 'info' },
]

export const SALARY_STATUS_OPTIONS = [
  { value: 'pending', label: '待发放', type: 'warning' },
  { value: 'paid', label: '已发放', type: 'success' },
  { value: 'cancelled', label: '已取消', type: 'info' },
]

export function jobStatusLabel(value: string): string {
  return JOB_STATUS_OPTIONS.find((item) => item.value === value)?.label || value
}

export function jobStatusType(value: string): string {
  return (JOB_STATUS_OPTIONS.find((item) => item.value === value)?.type as any) || 'info'
}

export function applicationStatusLabel(value: string): string {
  return APPLICATION_STATUS_OPTIONS.find((item) => item.value === value)?.label || value
}

export function applicationStatusType(value: string): string {
  return (APPLICATION_STATUS_OPTIONS.find((item) => item.value === value)?.type as any) || 'info'
}

export function salaryStatusLabel(value: string): string {
  return SALARY_STATUS_OPTIONS.find((item) => item.value === value)?.label || value
}

export function salaryStatusType(value: string): string {
  return (SALARY_STATUS_OPTIONS.find((item) => item.value === value)?.type as any) || 'info'
}
