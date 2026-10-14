<script setup lang="ts">
import { CategoryScale, Chart as ChartJS, Filler, LinearScale, LineController, LineElement, PointElement, Tooltip } from 'chart.js'
import { Line } from 'vue-chartjs'
import type { CampaignDetails } from '~/types/api'

ChartJS.register(LineElement, LineController, PointElement, CategoryScale, LinearScale, Tooltip, Filler)

const props = defineProps<{ daily: CampaignDetails['daily'] }>()

const data = computed(() => ({
  labels: props.daily.map((d) => new Date(`${d.day}T00:00:00`).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short' })),
  datasets: [{
    label: 'Выручка',
    data: props.daily.map((d) => d.revenueRub),
    borderColor: '#5b3fd9',
    backgroundColor: 'rgba(91, 63, 217, 0.12)',
    fill: true,
    tension: 0.3,
    pointRadius: 0,
    pointHitRadius: 8,
  }],
}))

const options = {
  responsive: true,
  maintainAspectRatio: false,
  plugins: {
    legend: { display: false },
    tooltip: { callbacks: { label: (ctx: { parsed: { y: number | null } }) => ` ${formatRub(ctx.parsed.y ?? 0)}` } },
  },
  scales: {
    x: { grid: { display: false }, ticks: { maxTicksLimit: 8 } },
    y: { ticks: { callback: (v: string | number) => formatCompactRub(Number(v)) }, grid: { color: '#eef0f3' } },
  },
}
</script>

<template>
  <div class="h-56">
    <Line :data="data" :options="options" />
  </div>
</template>
