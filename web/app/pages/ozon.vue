<script setup lang="ts">
import type { OzonLine, OzonReport, OzonStatus } from '~/types/api'

const statusMeta: Record<OzonStatus, { label: string; cls: string }> = {
  success: { label: 'Применён', cls: 'bg-good-soft text-good' },
  processing: { label: 'Загружается', cls: 'bg-brand-soft text-brand' },
  queued: { label: 'В очереди', cls: 'bg-canvas text-muted' },
  error: { label: 'Ошибка', cls: 'bg-bad-soft text-bad' },
}

const PAGE = 30
const page = ref(0)
const query = computed(() => ({ limit: PAGE, offset: page.value * PAGE }))
const { data, refresh } = await useFetch<{ items: OzonReport[]; total: number }>('/api/ozon/reports', { query })
const items = computed(() => data.value?.items ?? [])
const pages = computed(() => Math.max(1, Math.ceil((data.value?.total ?? 0) / PAGE)))

const todayIso = new Date().toISOString().slice(0, 10)
const busy = computed(() => items.value.some((r) => r.status === 'queued' || r.status === 'processing'))
const lastDone = computed(() => items.value.find((r) => r.finishedAt))
const month = computed(() => {
  const done = items.value.filter((r) => r.status === 'success')
  return {
    orders: done.reduce((s, r) => s + r.orders, 0),
    revenue: done.reduce((s, r) => s + r.revenueRub, 0),
    retries: done.filter((r) => r.error).length,
  }
})

// Журнал живой: пока что-то грузится — опрашиваем чаще.
let timer: ReturnType<typeof setTimeout> | undefined
function schedule() {
  timer = setTimeout(async () => {
    await refresh()
    if (opened.value) loadLines(opened.value, true)
    schedule()
  }, busy.value ? 2000 : 10000)
}
onMounted(schedule)
onBeforeUnmount(() => clearTimeout(timer))

const runError = ref('')
async function runNow() {
  runError.value = ''
  try {
    await $fetch('/api/ozon/reports/today', { method: 'POST' })
    await refresh()
  } catch (e) {
    runError.value = apiError(e)
  }
}

const opened = ref('')
const lines = ref<Record<string, OzonLine[]>>({})
async function loadLines(day: string, force = false) {
  if (lines.value[day] && !force) return
  lines.value[day] = await $fetch<OzonLine[]>(`/api/ozon/reports/${day}`)
}
function toggle(day: string) {
  opened.value = opened.value === day ? '' : day
  if (opened.value) loadLines(day)
}
</script>

<template>
  <div>
    <PageHeader title="Отчёты Ozon" subtitle="Продажи по промокодам блогеров. Отчёт за каждый день скачивается автоматически и сразу раскладывается по интеграциям.">
      <button class="btn btn-primary" :disabled="busy" @click="runNow">
        <svg class="size-4" :class="{ 'animate-spin': busy }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true">
          <path d="M20 12a8 8 0 1 1-2.34-5.66M20 4v4h-4" />
        </svg>
        {{ busy ? 'Загружается…' : 'Обновить за сегодня' }}
      </button>
    </PageHeader>

    <p v-if="runError" class="mb-4 rounded-xl bg-bad-soft px-4 py-3 text-sm text-bad" role="alert">{{ runError }}</p>

    <section class="mb-6 grid gap-4 lg:grid-cols-3">
      <div class="card p-5">
        <p class="text-sm text-muted">Подключение</p>
        <p class="mt-2 flex items-center gap-2 font-semibold">
          <span class="size-2 rounded-full bg-good" />Ozon Seller API
        </p>
        <p class="mt-1 text-xs text-muted">Client-Id ••••4821 · отчёт «Заказы с промокодами»</p>
        <p class="mt-1 text-xs text-muted">Расписание: ежедневно в 06:00 МСК за прошедший день; в демо сегодняшний отчёт докачивается каждые 30 секунд</p>
      </div>
      <div class="card p-5">
        <p class="text-sm text-muted">Последняя загрузка</p>
        <template v-if="lastDone?.finishedAt">
          <p class="mt-2 text-2xl font-semibold">{{ formatDateTime(lastDone.finishedAt) }}</p>
          <p class="mt-1 text-xs text-muted">отчёт за {{ formatDate(lastDone.day) }} · {{ statusMeta[lastDone.status].label.toLowerCase() }}</p>
        </template>
        <p v-else class="mt-2 text-2xl font-semibold">—</p>
      </div>
      <div class="card p-5">
        <p class="text-sm text-muted">{{ page === 0 ? 'За последние 30 дней' : 'За дни на этой странице' }}</p>
        <p class="mt-2 text-2xl font-semibold tabular-nums">{{ formatRub(month.revenue) }}</p>
        <p class="mt-1 text-xs text-muted">{{ formatNumber(month.orders) }} заказов по промокодам · повторов после ошибки: {{ month.retries }}</p>
      </div>
    </section>

    <div class="card overflow-hidden">
      <div class="overflow-x-auto">
        <table class="table-base min-w-[820px]">
          <thead>
            <tr>
              <th>Отчёт за</th><th>Статус</th><th>Загружен</th>
              <th class="text-right">Заказов в отчёте</th><th class="text-right">С промокодом</th><th class="text-right">Выручка по промокодам</th><th class="w-10" />
            </tr>
          </thead>
          <tbody>
            <template v-for="r in items" :key="r.day">
              <tr class="cursor-pointer hover:bg-canvas" :aria-expanded="opened === r.day" @click="toggle(r.day)">
                <td class="whitespace-nowrap">
                  <p class="font-medium">{{ formatDate(r.day) }}</p>
                  <p class="text-xs text-muted">{{ r.day === todayIso ? 'сегодня, обновляется в течение дня' : r.reportId }}</p>
                </td>
                <td>
                  <span class="inline-flex items-center gap-1.5 rounded-full px-2.5 py-0.5 text-xs font-semibold" :class="statusMeta[r.status].cls">
                    <span v-if="r.status === 'processing'" class="size-1.5 animate-pulse rounded-full bg-current" />
                    {{ statusMeta[r.status].label }}
                  </span>
                  <p v-if="r.status === 'error'" class="mt-1 text-xs text-bad">{{ r.error }}, повторим по расписанию</p>
                  <p v-else-if="r.error && r.status === 'success'" class="mt-1 text-xs text-muted" :title="r.error">со {{ r.attempts }}-й попытки</p>
                </td>
                <td class="whitespace-nowrap">
                  {{ formatDateTime(r.finishedAt ?? r.requestedAt) }}
                  <p class="text-xs text-muted">
                    {{ r.trigger === 'manual' ? 'вручную' : 'по расписанию' }}<template v-if="r.day === todayIso"> · загрузок: {{ r.attempts }}</template>
                  </p>
                </td>
                <td class="text-right tabular-nums">{{ formatNumber(r.rowsTotal) }}</td>
                <td class="text-right tabular-nums">{{ formatNumber(r.orders) }}</td>
                <td class="text-right tabular-nums">{{ formatRub(r.revenueRub) }}</td>
                <td class="text-muted">
                  <svg class="size-4 transition-transform" :class="{ 'rotate-180': opened === r.day }" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="m6 9 6 6 6-6" /></svg>
                </td>
              </tr>
              <tr v-if="opened === r.day">
                <td colspan="7" class="bg-canvas/60 px-4 py-4">
                  <p v-if="!lines[r.day]" class="text-sm text-muted">Загружаем строки…</p>
                  <p v-else-if="!lines[r.day]!.length" class="text-sm text-muted">В этот день заказов по промокодам не было.</p>
                  <table v-else class="w-full text-sm">
                    <thead>
                      <tr class="text-left text-xs text-muted">
                        <th class="pb-2 font-medium">Промокод</th><th class="pb-2 font-medium">Блогер</th><th class="pb-2 font-medium">Кампания</th>
                        <th class="pb-2 text-right font-medium">Заказы</th><th class="pb-2 text-right font-medium">Выручка</th>
                      </tr>
                    </thead>
                    <tbody>
                      <tr v-for="l in lines[r.day]" :key="l.integrationId">
                        <td class="py-1"><code class="rounded bg-white px-1.5 py-0.5 text-xs">{{ l.promoCode }}</code></td>
                        <td class="py-1"><NuxtLink :to="`/bloggers/${l.bloggerId}`" class="hover:text-brand" @click.stop>{{ l.blogger }}</NuxtLink></td>
                        <td class="py-1"><NuxtLink :to="`/campaigns/${l.campaignId}`" class="hover:text-brand" @click.stop>{{ l.campaign }}</NuxtLink></td>
                        <td class="py-1 text-right tabular-nums">{{ l.orders }}</td>
                        <td class="py-1 text-right tabular-nums">{{ formatRub(l.revenueRub) }}</td>
                      </tr>
                    </tbody>
                  </table>
                </td>
              </tr>
            </template>
          </tbody>
        </table>
      </div>
    </div>

    <nav v-if="pages > 1" class="mt-4 flex items-center justify-end gap-3 text-sm" aria-label="Страницы">
      <button class="btn btn-ghost" :disabled="page === 0" @click="page--">Новее</button>
      <span class="text-muted">{{ page + 1 }} из {{ pages }}</span>
      <button class="btn btn-ghost" :disabled="page + 1 >= pages" @click="page++">Старше</button>
    </nav>
  </div>
</template>
