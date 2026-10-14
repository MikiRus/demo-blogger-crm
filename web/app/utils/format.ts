const rub = new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0 })
const compact = new Intl.NumberFormat('ru-RU', { notation: 'compact', maximumFractionDigits: 1 })

export const formatRub = (v: number) => `${rub.format(v)} ₽`
export const formatCompactRub = (v: number) => `${compact.format(v)} ₽`
export const formatNumber = (v: number) => rub.format(v)
export const formatCompact = (v: number) => compact.format(v)

export const formatPercent = (v: number | null) => (v === null ? '—' : `${v.toLocaleString('ru-RU')}%`)
export const formatRoas = (v: number | null) => (v === null ? '—' : `×${v.toLocaleString('ru-RU')}`)

export const formatDate = (iso: string) =>
  new Date(`${iso}T00:00:00`).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })

const monthFmt = new Intl.DateTimeFormat('ru-RU', { month: 'short' })
export const formatMonth = (yyyyMM: string) => monthFmt.format(new Date(`${yyyyMM}-01T00:00:00`)).replace('.', '')

// ROAS ниже 1 — реклама не окупилась, 1–3 — на грани, выше 3 — хорошо.
export type Health = 'good' | 'warn' | 'bad' | 'none'
export const roasHealth = (roas: number | null): Health =>
  roas === null ? 'none' : roas >= 3 ? 'good' : roas >= 1 ? 'warn' : 'bad'

export const platformColors: Record<string, string> = {
  telegram: '#2a9fd8',
  youtube: '#e5342b',
  vk: '#2f6fe0',
  dzen: '#1c1c1e',
  rutube: '#14b8a6',
}

export const sourceLabels: Record<string, string> = { ozon: 'Ozon', wb: 'Wildberries', site: 'Сайт (CRM)' }
export const sourceColors: Record<string, string> = { ozon: '#2f6fe0', wb: '#a23bd6', site: '#e8860c' }

export const campaignStatusLabels = { planned: 'Запланирована', active: 'Идёт', finished: 'Завершена' } as const
