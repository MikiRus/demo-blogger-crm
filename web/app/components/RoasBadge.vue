<script setup lang="ts">
const props = defineProps<{ roas: number | null }>()
const styles = {
  good: 'bg-good-soft text-good',
  warn: 'bg-warn-soft text-warn',
  bad: 'bg-bad-soft text-bad',
  none: 'bg-canvas text-muted',
}
const cls = computed(() => styles[roasHealth(props.roas)])
</script>

<template>
  <span class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold tabular-nums" :class="cls">
    {{ formatRoas(roas) }}
  </span>
</template>
