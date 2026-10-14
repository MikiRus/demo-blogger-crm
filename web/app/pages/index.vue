<script setup lang="ts">
import type { Platform, Summary } from '~/types/api'

const periods = [
  { days: 30, label: '30 дней' },
  { days: 90, label: '90 дней' },
  { days: 180, label: 'Полгода' },
]
const days = ref(180)
const range = computed(() => {
  const to = new Date()
  const from = new Date(to.getTime() - days.value * 86_400_000)
  const iso = (d: Date) => d.toISOString().slice(0, 10)
  return { from: iso(from), to: iso(to) }
})

const { data: summary, status } = await useFetch<Summary>('/api/analytics/summary', { query: range })
const { data: platforms } = await useFetch<Platform[]>('/api/platforms', { default: () => [] })
</script>

<template>
  <div>
    <PageHeader title="Окупаемость интеграций" subtitle="Расходы — по дате публикации, выручка — по дню продажи">
      <div class="flex rounded-xl border border-line bg-white p-1 text-sm" role="group" aria-label="Период">
        <button
          v-for="p in periods"
          :key="p.days"
          class="rounded-lg px-3 py-1.5 font-medium"
          :class="days === p.days ? 'bg-brand text-white' : 'text-muted hover:text-ink'"
          @click="days = p.days"
        >
          {{ p.label }}
        </button>
      </div>
    </PageHeader>

    <div v-if="summary" class="space-y-6" :class="{ 'opacity-60 transition-opacity': status === 'pending' }">
      <section class="grid grid-cols-2 gap-4 lg:grid-cols-5">
        <KpiCard label="Расходы на блогеров" :value="formatRub(summary.total.costRub)" />
        <KpiCard label="Выручка" :value="formatRub(summary.total.revenueRub)" />
        <KpiCard label="Заказы" :value="formatNumber(summary.total.orders)" />
        <KpiCard label="ДРР" :value="formatPercent(summary.total.drr)" hint="доля расходов в выручке" />
        <KpiCard label="ROAS" :value="formatRoas(summary.total.roas)" hint="выручка на 1 ₽ рекламы" />
      </section>

      <section class="grid gap-6 lg:grid-cols-3">
        <div class="card p-5 lg:col-span-2">
          <h2 class="mb-4 font-semibold">По месяцам</h2>
          <MonthlyChart :monthly="summary.monthly" />
        </div>
        <div class="card p-5">
          <h2 class="mb-4 font-semibold">Откуда продажи</h2>
          <SourceChart :rows="summary.bySource" />
        </div>
      </section>

      <section class="grid gap-6 lg:grid-cols-2">
        <div class="card overflow-hidden">
          <h2 class="px-5 pt-5 font-semibold">Топ блогеров по выручке</h2>
          <div class="overflow-x-auto">
            <table class="table-base mt-2">
              <thead>
                <tr><th>Блогер</th><th class="text-right">Расходы</th><th class="text-right">Выручка</th><th class="text-right">ROAS</th></tr>
              </thead>
              <tbody>
                <tr v-for="b in summary.topBloggers" :key="b.id" class="hover:bg-canvas">
                  <td>
                    <NuxtLink :to="`/bloggers/${b.id}`" class="font-medium hover:text-brand">{{ b.name }}</NuxtLink>
                    <p class="text-xs text-muted">@{{ b.handle }}</p>
                  </td>
                  <td class="text-right tabular-nums">{{ formatRub(b.costRub) }}</td>
                  <td class="text-right tabular-nums">{{ formatRub(b.revenueRub) }}</td>
                  <td class="text-right"><RoasBadge :roas="b.roas" /></td>
                </tr>
              </tbody>
            </table>
          </div>
        </div>
        <div class="card p-5">
          <h2 class="font-semibold">Когда покупают</h2>
          <p class="mb-4 text-xs text-muted">Выручка по дням недели, внутри каждой площадки</p>
          <WeekdayHeatmap :cells="summary.heatmap" :platforms="platforms ?? []" />
        </div>
      </section>
    </div>
  </div>
</template>
