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

export interface CommunityTeam {
  id: number
  name: string
  teamType: string
  description: string
  quota: number
  location: string
  contactInfo: string
  status: string
  createdBy: number
  createdAt: string
  updatedAt: string
  members?: TeamMember[]
}

export interface TeamMember {
  id: number
  teamId: number
  userId: number
  name: string
  studentNo: string
  role: string
  status: string
  joinDate: string
  leaveDate?: string | null
  termStart: string
  termEnd: string
  remark: string
  createdAt: string
  updatedAt: string
}

export interface DutySchedule {
  id: number
  teamId: number
  date: string
  startTime: string
  endTime: string
  location: string
  status: string
  createdBy: number
  createdAt: string
  updatedAt: string
  records?: DutyRecord[]
}

export interface DutyRecord {
  id: number
  scheduleId: number
  teamId: number
  userId: number
  name: string
  checkinTime?: string | null
  checkoutTime?: string | null
  duration?: number | null
  status: string
  remark: string
  createdAt: string
  updatedAt: string
}

export interface VolunteerService {
  id: number
  teamId: number
  userId: number
  name: string
  studentNo: string
  title: string
  date: string
  hours: number
  category: string
  description: string
  verified: boolean
  verifiedBy?: number | null
  verifiedAt?: string | null
  createdBy: number
  createdAt: string
  updatedAt: string
}

export interface TeamListResponse {
  items: CommunityTeam[]
  total: number
  page: number
  size: number
}

export interface DutyListResponse {
  items: DutySchedule[]
  total: number
  page: number
  size: number
}

export interface ServiceListResponse {
  items: VolunteerService[]
  total: number
  page: number
  size: number
}

export interface TeamStatistics {
  totalTeams: number
  typeBreakdown: { teamType: string; count: number }[]
  totalMembers: number
  totalServiceHours: number
  totalDutyHours: number
}

export interface ServiceProfile {
  userId: number
  totalServiceHours: number
  totalDutyHours: number
  totalHours: number
  services: VolunteerService[]
  dutyRecords: DutyRecord[]
  teamMemberships: TeamMember[]
}

// ---------- API ----------

export const communityApi = {
  // 队伍管理
  listTeams(params: { page?: number; size?: number; teamType?: string; name?: string; status?: string } = {}) {
    const search = new URLSearchParams()
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== '') search.set(k, String(v))
    })
    return request<TeamListResponse>(`/api/v1/community/teams?${search.toString()}`)
  },
  createTeam(payload: Partial<CommunityTeam>) {
    return request<CommunityTeam>('/api/v1/community/teams', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  getTeam(id: number) {
    return request<CommunityTeam>(`/api/v1/community/teams/${id}`)
  },
  updateTeam(id: number, payload: Partial<CommunityTeam>) {
    return request<CommunityTeam>(`/api/v1/community/teams/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    })
  },
  deleteTeam(id: number) {
    return request<{ id: number }>(`/api/v1/community/teams/${id}`, { method: 'DELETE' })
  },

  // 成员管理
  listMembers(teamId: number, params: { status?: string; role?: string } = {}) {
    const search = new URLSearchParams()
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== '') search.set(k, String(v))
    })
    return request<TeamMember[]>(`/api/v1/community/teams/${teamId}/members?${search.toString()}`)
  },
  addMember(teamId: number, payload: Partial<TeamMember>) {
    return request<TeamMember>(`/api/v1/community/teams/${teamId}/members`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  updateMember(teamId: number, memberId: number, payload: Partial<TeamMember>) {
    return request<TeamMember>(`/api/v1/community/teams/${teamId}/members/${memberId}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    })
  },
  removeMember(teamId: number, memberId: number) {
    return request<{ id: number }>(`/api/v1/community/teams/${teamId}/members/${memberId}`, { method: 'DELETE' })
  },

  // 值班管理
  listDuties(teamId: number, params: { page?: number; size?: number; date?: string; status?: string } = {}) {
    const search = new URLSearchParams()
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== '') search.set(k, String(v))
    })
    return request<DutyListResponse>(`/api/v1/community/teams/${teamId}/duties?${search.toString()}`)
  },
  createDuty(teamId: number, payload: { date: string; startTime: string; endTime: string; location?: string; memberIds: number[] }) {
    return request<DutySchedule>(`/api/v1/community/teams/${teamId}/duties`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  dutyCheckin(teamId: number, scheduleId: number, userId: number) {
    return request<DutyRecord>(`/api/v1/community/teams/${teamId}/duties/${scheduleId}/checkin`, {
      method: 'POST',
      body: JSON.stringify({ userId }),
    })
  },
  dutyCheckout(teamId: number, scheduleId: number, userId: number) {
    return request<DutyRecord>(`/api/v1/community/teams/${teamId}/duties/${scheduleId}/checkout`, {
      method: 'POST',
      body: JSON.stringify({ userId }),
    })
  },

  // 志愿服务
  listServices(teamId: number, params: { page?: number; size?: number; category?: string; verified?: string } = {}) {
    const search = new URLSearchParams()
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== '') search.set(k, String(v))
    })
    return request<ServiceListResponse>(`/api/v1/community/teams/${teamId}/services?${search.toString()}`)
  },
  createService(teamId: number, payload: Partial<VolunteerService>) {
    return request<VolunteerService>(`/api/v1/community/teams/${teamId}/services`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  verifyService(teamId: number, serviceId: number, verified: boolean) {
    return request<VolunteerService>(`/api/v1/community/teams/${teamId}/services/${serviceId}/verify`, {
      method: 'PUT',
      body: JSON.stringify({ verified }),
    })
  },

  // 统计
  statistics() {
    return request<TeamStatistics>('/api/v1/community/teams/statistics')
  },
  serviceProfile(userId?: number) {
    const params = userId ? `?userId=${userId}` : ''
    return request<ServiceProfile>(`/api/v1/community/service-profile${params}`)
  },
}

// ---------- 常量 ----------

export const TEAM_TYPE_OPTIONS = [
  { value: 'autonomy', label: '自治队伍', color: '#409EFF' },
  { value: 'volunteer', label: '志愿服务队', color: '#67C23A' },
  { value: 'duty', label: '值班队伍', color: '#E6A23C' },
]

export const MEMBER_ROLE_OPTIONS = [
  { value: 'leader', label: '负责人' },
  { value: 'vice', label: '副负责人' },
  { value: 'member', label: '成员' },
  { value: 'trainee', label: '实习成员' },
]

export const MEMBER_STATUS_OPTIONS = [
  { value: 'active', label: '在职', type: 'success' },
  { value: 'pending', label: '待审核', type: 'warning' },
  { value: 'left', label: '已退出', type: 'info' },
]

export const DUTY_STATUS_OPTIONS = [
  { value: 'scheduled', label: '已排班', type: 'info' },
  { value: 'active', label: '值班中', type: 'primary' },
  { value: 'completed', label: '已完成', type: 'success' },
  { value: 'absent', label: '缺勤', type: 'danger' },
]

export const SERVICE_CATEGORY_OPTIONS = [
  { value: 'community', label: '社区服务' },
  { value: 'campus', label: '校园服务' },
  { value: 'charity', label: '公益慈善' },
  { value: 'culture', label: '文化传播' },
  { value: 'other', label: '其他' },
]

export function teamTypeLabel(value: string): string {
  return TEAM_TYPE_OPTIONS.find((item) => item.value === value)?.label || value
}

export function memberRoleLabel(value: string): string {
  return MEMBER_ROLE_OPTIONS.find((item) => item.value === value)?.label || value
}

export function memberStatusType(value: string): string {
  return (MEMBER_STATUS_OPTIONS.find((item) => item.value === value)?.type as any) || 'info'
}

export function memberStatusLabel(value: string): string {
  return MEMBER_STATUS_OPTIONS.find((item) => item.value === value)?.label || value
}

export function dutyStatusLabel(value: string): string {
  return DUTY_STATUS_OPTIONS.find((item) => item.value === value)?.label || value
}

export function dutyStatusType(value: string): string {
  return (DUTY_STATUS_OPTIONS.find((item) => item.value === value)?.type as any) || 'info'
}

export function serviceCategoryLabel(value: string): string {
  return SERVICE_CATEGORY_OPTIONS.find((item) => item.value === value)?.label || value
}
