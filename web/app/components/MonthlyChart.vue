<script setup lang="ts">
import { Chart as ChartJS, BarElement, CategoryScale, Legend, LinearScale, LineElement, PointElement, Tooltip, BarController, LineController } from 'chart.js'
import { Chart } from 'vue-chartjs'
import type { Summary } from '~/types/api'

ChartJS.register(BarElement, BarController, LineElement, LineController, PointElement, CategoryScale, LinearScale, Tooltip, Legend)

const props = defineProps<{ monthly: Summary['monthly'] }>()

const data = computed(() => ({
  labels: props.monthly.map((m) => formatMonth(m.month)),
  datasets: [
    { type: 'bar' as const, label: 'Расходы', data: props.monthly.map((m) => m.costRub), backgroundColor: '#c9bff5', borderRadius: 6, yAxisID: 'y' },
    { type: 'bar' as const, label: 'Выручка', data: props.monthly.map((m) => m.revenueRub), backgroundColor: '#5b3fd9', borderRadius: 6, yAxisID: 'y' },
    {
      type: 'line' as const, label: 'ДРР', data: props.monthly.map((m) => m.drr), borderColor: '#e8860c', backgroundColor: '#e8860c',
      tension: 0.35, pointRadius: 3, yAxisID: 'drr',
    },
  ],
}))

const options = {
  responsive: true,
  maintainAspectRatio: false,
  interaction: { mode: 'index' as const, intersect: false },
  plugins: {
    legend: { position: 'bottom' as const, labels: { usePointStyle: true, boxWidth: 8, font: { family: 'Onest' } } },
    tooltip: {
      callbacks: {
        label: (ctx: { dataset: { label?: string }; parsed: { y: number | null } }) =>
          ctx.dataset.label === 'ДРР'
            ? ` ДРР: ${formatPercent(ctx.parsed.y)}`
            : ` ${ctx.dataset.label}: ${formatRub(ctx.parsed.y ?? 0)}`,
      },
    },
  },
  scales: {
    x: { grid: { display: false } },
    y: { ticks: { callback: (v: string | number) => formatCompactRub(Number(v)) }, grid: { color: '#eef0f3' } },
    drr: { position: 'right' as const, min: 0, grid: { display: false }, ticks: { callback: (v: string | number) => `${v}%` } },
  },
}
</script>

<template>
  <div class="h-72">
    <Chart type="bar" :data="data" :options="options" />
  </div>
</template>
