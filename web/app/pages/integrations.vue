<script setup lang="ts">
import type { IntegrationListItem } from '~/types/api'

const tabs = [
  { value: '', label: 'Все' },
  { value: 'published', label: 'Вышли' },
  { value: 'planned', label: 'Запланированы' },
]
const PAGE = 20
const status = ref('')
const search = ref('')
const q = refDebounced(search, 300)
const page = ref(0)
watch([status, q], () => (page.value = 0))

const query = computed(() => ({ status: status.value, q: q.value, limit: PAGE, offset: page.value * PAGE }))
const { data, status: loading } = await useFetch<{ items: IntegrationListItem[]; total: number }>('/api/integrations', { query })
const pages = computed(() => Math.max(1, Math.ceil((data.value?.total ?? 0) / PAGE)))
</script>

<template>
  <div>
    <PageHeader title="Интеграции" :subtitle="`${data?.total ?? 0} публикаций во всех кампаниях`" />

    <div class="mb-4 flex flex-wrap items-center gap-3">
      <div class="flex flex-wrap gap-2" role="group" aria-label="Статус">
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
      <label class="sr-only" for="q">Поиск</label>
      <input id="q" v-model="search" class="input sm:ml-auto sm:w-72" placeholder="Блогер, ник или промокод">
    </div>

    <div class="card overflow-hidden" :class="{ 'opacity-60': loading === 'pending' }">
      <div class="overflow-x-auto">
        <table class="table-base min-w-[900px]">
          <thead>
            <tr>
              <th>Блогер</th><th>Кампания</th><th>Публикация</th><th>Промокод</th>
              <th class="text-right">Стоимость</th><th class="text-right">Выручка</th><th class="cursor-help text-right" :title="DRR_HINT">ДРР</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="i in data?.items" :key="i.id" class="hover:bg-canvas">
              <td>
                <NuxtLink :to="`/bloggers/${i.bloggerId}`" class="font-medium hover:text-brand">{{ i.blogger }}</NuxtLink>
                <p class="flex items-center gap-1.5 text-xs text-muted">
                  <span class="size-1.5 rounded-full" :style="{ background: platformColors[i.platform] }" />@{{ i.handle }}
                </p>
              </td>
              <td class="max-w-64"><NuxtLink :to="`/campaigns/${i.campaignId}`" class="hover:text-brand">{{ i.campaign }}</NuxtLink></td>
              <td class="whitespace-nowrap">
                {{ formatDate(i.publishDate) }}
                <span v-if="i.status === 'planned'" class="ml-1 rounded-full bg-canvas px-2 py-0.5 text-xs text-muted">план</span>
              </td>
              <td>
                <code class="rounded bg-canvas px-1.5 py-0.5 text-xs">{{ i.promoCode }}</code>
                <p class="mt-1 text-xs text-muted">{{ i.utmCampaign }}</p>
              </td>
              <td class="text-right tabular-nums">{{ formatRub(i.costRub) }}</td>
              <td class="text-right tabular-nums">{{ formatRub(i.revenueRub) }}</td>
              <td class="text-right"><DrrBadge :drr="i.drr" /></td>
            </tr>
            <tr v-if="data && !data.items.length">
              <td colspan="7" class="py-10 text-center text-muted">Ничего не нашли</td>
            </tr>
          </tbody>
        </table>
      </div>
    </div>

    <nav v-if="pages > 1" class="mt-4 flex items-center justify-end gap-3 text-sm" aria-label="Страницы">
      <button class="btn btn-ghost" :disabled="page === 0" @click="page--">Назад</button>
      <span class="text-muted">{{ page + 1 }} из {{ pages }}</span>
      <button class="btn btn-ghost" :disabled="page + 1 >= pages" @click="page++">Вперёд</button>
    </nav>
  </div>
</template>
