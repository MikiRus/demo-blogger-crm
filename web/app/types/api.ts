export interface Totals {
  costRub: number
  revenueRub: number
  orders: number
  drr: number | null
  roas: number | null
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

export type RowStatus = 'new' | 'duplicate' | 'unknown_promo' | 'invalid'

export interface ReportRow {
  line: number
  day: string
  promoCode: string
  source: string
  orders: number
  revenueRub: number
  error?: string
}

export interface PreviewRow extends ReportRow {
  status: RowStatus
  integrationId?: number
  blogger?: string
}

export interface Preview {
  rows: PreviewRow[]
  summary: Record<RowStatus, number>
}
