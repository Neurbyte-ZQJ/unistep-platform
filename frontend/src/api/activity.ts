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

export interface ClubActivity {
  id: number
  clubName: string
  title: string
  startTime: string
  endTime: string
  location: string
  capacity: number
  description: string
  budget?: number | null
  status: string
  approvalOpinion?: string
  summary?: string
  createdBy: number
  approvedBy?: number | null
  approvedAt?: string | null
  createdAt: string
  updatedAt: string
  registrations?: ActivityRegistration[]
  checkins?: ActivityCheckin[]
  files?: ActivityFile[]
}

export interface ActivityRegistration {
  id: number
  activityId: number
  studentId: number
  status: string
  registeredAt: string
  cancelledAt?: string | null
}

export interface ActivityCheckin {
  id: number
  activityId: number
  studentId: number
  checkinTime: string
  checkinMethod: string
}

export interface ActivityFile {
  id: number
  activityId: number
  fileName: string
  objectKey: string
  url: string
  fileType: string
  size: number
  uploadedBy: number
  createdAt: string
}

export interface ActivityListResponse {
  items: ClubActivity[]
  total: number
  page: number
  size: number
}

export interface ActivityStatistics {
  totalActivities: number
  statusBreakdown: { status: string; count: number }[]
  totalRegistrations: number
  totalCheckins: number
  recentActivities: ClubActivity[]
}

// ---------- API ----------

export const activityApi = {
  list(params: { page?: number; size?: number; status?: string; clubName?: string; title?: string } = {}) {
    const search = new URLSearchParams()
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== '') search.set(k, String(v))
    })
    return request<ActivityListResponse>(`/api/v1/activities?${search.toString()}`)
  },
  create(payload: Partial<ClubActivity>) {
    return request<ClubActivity>('/api/v1/activities', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  detail(id: number) {
    return request<ClubActivity>(`/api/v1/activities/${id}`)
  },
  update(id: number, payload: Partial<ClubActivity>) {
    return request<ClubActivity>(`/api/v1/activities/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    })
  },
  remove(id: number) {
    return request<{ id: number }>(`/api/v1/activities/${id}`, { method: 'DELETE' })
  },
  submitForApproval(id: number) {
    return request<ClubActivity>(`/api/v1/activities/${id}/submit`, { method: 'POST' })
  },
  approve(id: number, opinion: string, approve: boolean) {
    return request<ClubActivity>(`/api/v1/activities/${id}/approve`, {
      method: 'POST',
      body: JSON.stringify({ opinion, approve }),
    })
  },
  register(id: number) {
    return request<ActivityRegistration>(`/api/v1/activities/${id}/register`, { method: 'POST' })
  },
  cancelRegistration(id: number) {
    return request<{ activityId: number }>(`/api/v1/activities/${id}/cancel-registration`, { method: 'POST' })
  },
  checkin(id: number, studentId: number) {
    return request<ActivityCheckin>(`/api/v1/activities/${id}/checkin`, {
      method: 'POST',
      body: JSON.stringify({ studentId }),
    })
  },
  uploadFile(id: number, file: File, fileType: string) {
    const form = new FormData()
    form.append('file', file)
    form.append('fileType', fileType)
    return request<ActivityFile>(`/api/v1/activities/${id}/files`, {
      method: 'POST',
      body: form,
    })
  },
  submitSummary(id: number, summary: string) {
    return request<ClubActivity>(`/api/v1/activities/${id}/summary`, {
      method: 'POST',
      body: JSON.stringify({ summary }),
    })
  },
  updateStatus(id: number, status: string) {
    return request<ClubActivity>(`/api/v1/activities/${id}/status`, {
      method: 'PUT',
      body: JSON.stringify({ status }),
    })
  },
  statistics() {
    return request<ActivityStatistics>('/api/v1/activities/statistics')
  },
}

// ---------- 常量 ----------

export const ACTIVITY_STATUS_OPTIONS = [
  { value: 'draft', label: '草稿', type: 'info' },
  { value: 'pending', label: '待审批', type: 'warning' },
  { value: 'rejected', label: '已驳回', type: 'danger' },
  { value: 'reg_open', label: '报名开放', type: 'success' },
  { value: 'reg_closed', label: '报名截止', type: '' },
  { value: 'in_progress', label: '进行中', type: 'primary' },
  { value: 'completed', label: '已完成', type: 'success' },
  { value: 'archived', label: '已归档', type: 'info' },
]

export function activityStatusLabel(value: string): string {
  return ACTIVITY_STATUS_OPTIONS.find((item) => item.value === value)?.label || value
}

export function activityStatusType(value: string): string {
  return (ACTIVITY_STATUS_OPTIONS.find((item) => item.value === value)?.type as any) || 'info'
}
