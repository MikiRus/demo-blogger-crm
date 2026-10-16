const rub = new Intl.NumberFormat('ru-RU', { maximumFractionDigits: 0 })
const compact = new Intl.NumberFormat('ru-RU', { notation: 'compact', maximumFractionDigits: 1 })

export const formatRub = (v: number) => `${rub.format(v)} ₽`
export const formatCompactRub = (v: number) => `${compact.format(v)} ₽`
export const formatNumber = (v: number) => rub.format(v)
export const formatCompact = (v: number) => compact.format(v)

export const formatPercent = (v: number | null) => (v === null ? '—' : `${v.toLocaleString('ru-RU')}%`)

export const formatDate = (iso: string) =>
  new Date(`${iso}T00:00:00`).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })

const monthFmt = new Intl.DateTimeFormat('ru-RU', { month: 'short' })
export const formatMonth = (yyyyMM: string) => monthFmt.format(new Date(`${yyyyMM}-01T00:00:00`)).replace('.', '')

const dateTimeFmt = new Intl.DateTimeFormat('ru-RU', { day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit' })
export const formatDateTime = (iso: string) => dateTimeFmt.format(new Date(iso))

const relFmt = new Intl.RelativeTimeFormat('ru-RU', { numeric: 'auto' })
export function formatAgo(iso: string | null, now = Date.now()) {
  if (!iso) return 'ещё не заходил'
  const min = Math.round((now - new Date(iso).getTime()) / 60_000)
  if (min < 5) return 'в сети'
  if (min < 60) return relFmt.format(-min, 'minute')
  if (min < 24 * 60) return relFmt.format(-Math.round(min / 60), 'hour')
  return relFmt.format(-Math.round(min / 1440), 'day')
}

export const DRR_HINT = 'ДРР — доля рекламных расходов: расходы на блогеров ÷ выручка × 100%. Чем ниже, тем выгоднее.'

// ДРР до 25% — интеграция окупается с запасом, до 40% — на грани, выше — в минус.
export type Health = 'good' | 'warn' | 'bad' | 'none'
export const drrHealth = (drr: number | null): Health =>
  drr === null ? 'none' : drr <= 25 ? 'good' : drr <= 40 ? 'warn' : 'bad'

export const platformColors: Record<string, string> = {
  telegram: '#2a9fd8',
  youtube: '#e5342b',
  vk: '#2f6fe0',
  dzen: '#1c1c1e',
  rutube: '#14b8a6',
}

export const sourceLabels: Record<string, string> = { ozon: 'Ozon', wb: 'Wildberries', site: 'Сайт (CRM)' }
export const sourceColors: Record<string, string> = { ozon: '#2f6fe0', wb: '#a23bd6', site: '#e8860c' }

export const roleLabels: Record<string, string> = {
  admin: 'Администратор',
  marketer: 'Маркетолог',
  manager: 'Менеджер по блогерам',
  analyst: 'Аналитик',
  finance: 'Финансы',
}

export const campaignStatusLabels = { planned: 'Запланирована', active: 'Идёт', finished: 'Завершена' } as const
