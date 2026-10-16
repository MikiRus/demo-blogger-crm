<script setup lang="ts">
const props = defineProps<{ drr: number | null }>()
const styles = {
  good: 'bg-good-soft text-good',
  warn: 'bg-warn-soft text-warn',
  bad: 'bg-bad-soft text-bad',
  none: 'bg-canvas text-muted',
}
const cls = computed(() => styles[drrHealth(props.drr)])
</script>

<template>
  <span class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold tabular-nums" :class="cls" :title="DRR_HINT">
    {{ formatPercent(drr) }}
  </span>
</template>
