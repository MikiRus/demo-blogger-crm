<script setup lang="ts">
import { ArcElement, Chart as ChartJS, DoughnutController, Tooltip } from 'chart.js'
import { Doughnut } from 'vue-chartjs'
import type { Summary } from '~/types/api'

ChartJS.register(ArcElement, DoughnutController, Tooltip)

const props = defineProps<{ rows: Summary['bySource'] }>()
const total = computed(() => props.rows.reduce((s, r) => s + r.revenueRub, 0))

const data = computed(() => ({
  labels: props.rows.map((r) => sourceLabels[r.source] ?? r.source),
  datasets: [{
    data: props.rows.map((r) => r.revenueRub),
    backgroundColor: props.rows.map((r) => sourceColors[r.source] ?? '#999'),
    borderWidth: 2,
    borderColor: '#fff',
  }],
}))

const options = {
  responsive: true,
  maintainAspectRatio: false,
  cutout: '68%',
  plugins: {
    legend: { display: false },
    tooltip: { callbacks: { label: (ctx: { parsed: number }) => ` ${formatRub(ctx.parsed)}` } },
  },
}
</script>

<template>
  <div class="flex flex-col items-center gap-4 sm:flex-row lg:flex-col xl:flex-row">
    <div class="relative size-40 shrink-0">
      <Doughnut :data="data" :options="options" />
      <div class="pointer-events-none absolute inset-0 grid place-items-center text-center">
        <div>
          <p class="text-xs text-muted">Выручка</p>
          <p class="text-sm font-semibold">{{ formatCompactRub(total) }}</p>
        </div>
      </div>
    </div>
    <ul class="w-full space-y-3 text-sm">
      <li v-for="r in rows" :key="r.source" class="flex items-center gap-3">
        <span class="size-2.5 rounded-full" :style="{ background: sourceColors[r.source] }" />
        <span class="flex-1">{{ sourceLabels[r.source] ?? r.source }}</span>
        <span class="tabular-nums text-muted">{{ total ? Math.round((r.revenueRub / total) * 100) : 0 }}%</span>
      </li>
    </ul>
  </div>
</template>
