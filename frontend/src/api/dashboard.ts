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

export interface DashboardOverview {
  members: {
    total: number
    stageBreakdown: { stage: string; count: number }[]
  }
  activities: {
    total: number
    statusBreakdown: { status: string; count: number }[]
    totalRegistrations: number
    totalCheckins: number
  }
  services: {
    totalServiceHours: number
    totalDutyHours: number
    totalHours: number
    totalVolunteerCount: number
    totalDutyRecordCount: number
  }
  workstudy: {
    totalJobs: number
    statusBreakdown: { status: string; count: number }[]
    totalApplications: number
    totalAccepted: number
    totalSalaryPaid: number
    totalWorkHours: number
  }
}

export interface TrendItem {
  month: string
  count: number
}

export interface HoursTrendItem {
  month: string
  hours: number
}

export interface ServiceTrendData {
  volunteer: HoursTrendItem[]
  duty: HoursTrendItem[]
}

// ---------- API ----------

export const dashboardApi = {
  overview() {
    return request<DashboardOverview>('/api/v1/dashboard/overview')
  },
  memberTrend() {
    return request<{ trend: TrendItem[] }>('/api/v1/dashboard/member-trend')
  },
  activityTrend() {
    return request<{ trend: TrendItem[] }>('/api/v1/dashboard/activity-trend')
  },
  serviceTrend() {
    return request<ServiceTrendData>('/api/v1/dashboard/service-trend')
  },
}
