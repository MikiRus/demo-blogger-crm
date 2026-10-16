export interface Totals {
  costRub: number
  revenueRub: number
  orders: number
  drr: number | null
}

export interface Platform {
  id: number
  code: string
  name: string
}

export interface BloggerItem extends Totals {
  id: number
  name: string
  handle: string
  platform: Platform
  followers: number
  avgReach: number
  category: string
  integrations: number
}

export interface BloggerHistory extends Totals {
  id: number
  campaignId: number
  campaign: string
  publishDate: string
  promoCode: string
  status: 'planned' | 'published'
}

export interface BloggerDetails extends BloggerItem {
  history: BloggerHistory[]
}

export type CampaignStatus = 'planned' | 'active' | 'finished'

export interface CampaignItem extends Totals {
  id: number
  title: string
  product: string
  status: CampaignStatus
  startDate: string
  endDate: string
  budgetRub: number
  integrations: number
}

export interface CampaignIntegration extends Totals {
  id: number
  bloggerId: number
  blogger: string
  handle: string
  platform: string
  publishDate: string
  promoCode: string
  utmCampaign: string
  status: 'planned' | 'published'
}

export interface CampaignDetails extends CampaignItem {
  integrationList: CampaignIntegration[]
  daily: { day: string; orders: number; revenueRub: number }[]
}

export interface Summary {
  from: string
  to: string
  total: Totals
  monthly: (Totals & { month: string })[]
  topBloggers: (Totals & { id: number; name: string; handle: string; platform: string })[]
  heatmap: { weekday: number; platform: string; revenueRub: number }[]
  bySource: { source: string; orders: number; revenueRub: number }[]
}

export interface IntegrationListItem extends CampaignIntegration {
  campaignId: number
  campaign: string
}

export type OzonStatus = 'queued' | 'processing' | 'success' | 'error'

export interface OzonReport {
  day: string
  reportId: string
  status: OzonStatus
  trigger: 'schedule' | 'manual'
  attempts: number
  rowsTotal: number
  orders: number
  revenueRub: number
  error: string | null
  requestedAt: string
  finishedAt: string | null
}

export interface OzonLine {
  integrationId: number
  promoCode: string
  bloggerId: number
  blogger: string
  campaignId: number
  campaign: string
  orders: number
  revenueRub: number
}

export type UserRole = 'admin' | 'marketer' | 'manager' | 'analyst' | 'finance'

export interface User {
  id: number
  name: string
  email: string
  role: UserRole
  status: 'active' | 'invited' | 'blocked'
  lastSeenAt: string | null
}
