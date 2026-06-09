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

export interface MemberProfile {
  id: number
  userId: number
  name: string
  studentNo: string
  gender?: string
  birthday?: string
  idCard?: string
  nation?: string
  phone?: string
  college?: string
  major?: string
  className?: string
  stage: string
  joinDate?: string
  remark?: string
  applications?: LeagueApplication[]
  activistRecords?: ActivistRecord[]
  developRecords?: DevelopTargetRecord[]
  politicalRecords?: PoliticalReview[]
  attachments?: MemberAttachment[]
}

export interface LeagueApplication {
  id?: number
  applyDate: string
  motivation?: string
  introducer?: string
  status?: string
  reviewNote?: string
}

export interface ActivistRecord {
  id?: number
  startDate: string
  trainer?: string
  trainPlan?: string
  evaluation?: string
  score?: number
}

export interface DevelopTargetRecord {
  id?: number
  confirmedDate: string
  mentor?: string
  publicityNote?: string
  conclusion?: string
}

export interface PoliticalReview {
  id?: number
  reviewDate: string
  reviewer?: string
  familyMembers?: string
  conclusion?: string
  status?: string
}

export interface MemberAttachment {
  id: number
  category: string
  fileName: string
  url: string
  size: number
  createdAt: string
}

export interface MemberListResponse {
  items: MemberProfile[]
  total: number
  page: number
  size: number
}

export interface MemberArchive {
  profile: MemberProfile
  timeline: { date: string; stage: string; text: string }[]
  summary: Record<string, number | string>
}

export const memberApi = {
  list(params: { page?: number; size?: number; stage?: string; name?: string } = {}) {
    const search = new URLSearchParams()
    Object.entries(params).forEach(([k, v]) => {
      if (v !== undefined && v !== '') search.set(k, String(v))
    })
    return request<MemberListResponse>(`/api/v1/members?${search.toString()}`)
  },
  create(payload: Partial<MemberProfile>) {
    return request<MemberProfile>('/api/v1/members', {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  update(id: number, payload: Partial<MemberProfile>) {
    return request<MemberProfile>(`/api/v1/members/${id}`, {
      method: 'PUT',
      body: JSON.stringify(payload),
    })
  },
  detail(id: number) {
    return request<MemberProfile>(`/api/v1/members/${id}`)
  },
  remove(id: number) {
    return request<{ id: number }>(`/api/v1/members/${id}`, { method: 'DELETE' })
  },
  archive(id: number) {
    return request<MemberArchive>(`/api/v1/members/${id}/archive`)
  },
  addApplication(id: number, payload: LeagueApplication) {
    return request<LeagueApplication>(`/api/v1/members/${id}/applications`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  addActivist(id: number, payload: ActivistRecord) {
    return request<ActivistRecord>(`/api/v1/members/${id}/activists`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  addDevelop(id: number, payload: DevelopTargetRecord) {
    return request<DevelopTargetRecord>(`/api/v1/members/${id}/develop-targets`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  addPoliticalReview(id: number, payload: PoliticalReview) {
    return request<PoliticalReview>(`/api/v1/members/${id}/political-reviews`, {
      method: 'POST',
      body: JSON.stringify(payload),
    })
  },
  uploadAttachment(id: number, file: File, category: string) {
    const form = new FormData()
    form.append('file', file)
    form.append('category', category)
    return request<MemberAttachment>(`/api/v1/members/${id}/attachments`, {
      method: 'POST',
      body: form,
    })
  },
}

export const STAGE_OPTIONS = [
  { value: 'applicant', label: '入团申请人' },
  { value: 'activist', label: '积极分子' },
  { value: 'develop_target', label: '发展对象' },
  { value: 'political_review', label: '政审备案' },
  { value: 'league_member', label: '正式团员' },
]

export function stageLabel(value: string): string {
  return STAGE_OPTIONS.find((item) => item.value === value)?.label || value
}
