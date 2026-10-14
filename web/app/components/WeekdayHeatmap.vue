<script setup lang="ts">
import type { Platform, Summary } from '~/types/api'

const props = defineProps<{ cells: Summary['heatmap']; platforms: Platform[] }>()
const weekdays = ['Пн', 'Вт', 'Ср', 'Чт', 'Пт', 'Сб', 'Вс']

// Нормируем по строке: сравниваем дни внутри площадки, а не площадки между собой,
// иначе крупная площадка «засветит» всю карту.
const rows = computed(() =>
  props.platforms
    .map((p) => {
      const values = weekdays.map((_, i) => props.cells.find((c) => c.platform === p.code && c.weekday === i + 1)?.revenueRub ?? 0)
      const max = Math.max(...values, 1)
      return { platform: p, values, shares: values.map((v) => v / max) }
    })
    .filter((r) => r.values.some((v) => v > 0)),
)
</script>

<template>
  <div class="overflow-x-auto">
    <table class="w-full min-w-[480px] border-separate border-spacing-1 text-xs">
      <thead>
        <tr>
          <th class="w-28" />
          <th v-for="d in weekdays" :key="d" class="pb-1 font-medium text-muted">{{ d }}</th>
        </tr>
      </thead>
      <tbody>
        <tr v-for="r in rows" :key="r.platform.code">
          <th class="pr-2 text-left font-medium">{{ r.platform.name }}</th>
          <td
            v-for="(v, i) in r.values"
            :key="i"
            class="h-9 rounded-md text-center tabular-nums"
            :style="{ background: `rgba(91, 63, 217, ${0.08 + r.shares[i]! * 0.82})`, color: r.shares[i]! > 0.55 ? '#fff' : '#16181d' }"
            :title="`${r.platform.name}, ${weekdays[i]}: ${formatRub(v)}`"
          >
            {{ formatCompact(v) }}
          </td>
        </tr>
      </tbody>
    </table>
  </div>
</template>
