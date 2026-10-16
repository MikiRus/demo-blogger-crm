<script setup lang="ts">
import type { CampaignItem, CampaignStatus } from '~/types/api'

const tabs: { value: '' | CampaignStatus; label: string }[] = [
  { value: '', label: 'Все' },
  { value: 'active', label: 'Идут' },
  { value: 'planned', label: 'Запланированы' },
  { value: 'finished', label: 'Завершены' },
]
const status = ref<'' | CampaignStatus>('')
const { data: campaigns, refresh } = await useFetch<CampaignItem[]>('/api/campaigns', { query: { status }, default: () => [] })

const showCreate = ref(false)
const budgetUsed = (c: CampaignItem) => (c.budgetRub ? Math.min(100, Math.round((c.costRub / c.budgetRub) * 100)) : 0)
</script>

<template>
  <div>
    <PageHeader title="Рекламные кампании" subtitle="Продвижение товара через пул блогеров">
      <button class="btn btn-primary" @click="showCreate = true">Новая кампания</button>
    </PageHeader>

    <div class="mb-4 flex flex-wrap gap-2" role="group" aria-label="Статус">
      <button
        v-for="t in tabs"
        :key="t.value"
        class="rounded-full border px-4 py-1.5 text-sm font-medium"
        :class="status === t.value ? 'border-brand bg-brand text-white' : 'border-line bg-white text-muted hover:text-ink'"
        @click="status = t.value"
      >
        {{ t.label }}
      </button>
    </div>

    <div class="grid gap-4 md:grid-cols-2 2xl:grid-cols-3">
      <NuxtLink v-for="c in campaigns" :key="c.id" :to="`/campaigns/${c.id}`" class="card block p-5 transition-shadow hover:shadow-md">
        <div class="mb-3 flex items-start justify-between gap-3">
          <div>
            <h2 class="font-semibold">{{ c.title }}</h2>
            <p class="mt-0.5 text-xs text-muted">{{ formatDate(c.startDate) }} — {{ formatDate(c.endDate) }}</p>
          </div>
          <CampaignBadge :status="c.status" />
        </div>
        <dl class="grid grid-cols-3 gap-3 text-sm">
          <div><dt class="text-xs text-muted">Расходы</dt><dd class="font-medium tabular-nums">{{ formatCompactRub(c.costRub) }}</dd></div>
          <div><dt class="text-xs text-muted">Выручка</dt><dd class="font-medium tabular-nums">{{ formatCompactRub(c.revenueRub) }}</dd></div>
          <div><dt class="cursor-help text-xs text-muted" :title="DRR_HINT">ДРР</dt><dd><DrrBadge :drr="c.drr" /></dd></div>
        </dl>
        <div class="mt-4">
          <div class="mb-1 flex justify-between text-xs text-muted">
            <span>{{ c.integrations }} интеграций</span>
            <span>бюджет {{ budgetUsed(c) }}%</span>
          </div>
          <div class="h-1.5 overflow-hidden rounded-full bg-canvas">
            <div class="h-full rounded-full bg-brand" :style="{ width: `${budgetUsed(c)}%` }" />
          </div>
        </div>
      </NuxtLink>
    </div>
    <p v-if="!campaigns.length" class="card p-10 text-center text-muted">Кампаний с таким статусом нет</p>

    <CampaignCreateDialog v-if="showCreate" @close="showCreate = false" @created="(id) => { refresh(); navigateTo(`/campaigns/${id}`) }" />
  </div>
</template>
