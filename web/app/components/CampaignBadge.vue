<script setup lang="ts">
import type { CampaignStatus } from '~/types/api'

defineProps<{ status: CampaignStatus }>()
const styles: Record<CampaignStatus, string> = {
  planned: 'bg-canvas text-muted',
  active: 'bg-brand-soft text-brand',
  finished: 'bg-good-soft text-good',
}
</script>

<template>
  <span class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold" :class="styles[status]">
    {{ campaignStatusLabels[status] }}
  </span>
</template>
