<script setup lang="ts">
import type { BloggerItem, CampaignDetails } from '~/types/api'

const route = useRoute()
const { data: c, error, refresh } = await useFetch<CampaignDetails>(() => `/api/campaigns/${route.params.id}`)

const showAdd = ref(false)
// Список блогеров нужен только в диалоге — грузим при первом открытии.
const bloggers = ref<BloggerItem[]>([])
async function openAdd() {
  showAdd.value = true
  if (!bloggers.value.length) {
    const res = await $fetch<{ items: BloggerItem[] }>('/api/bloggers', { query: { limit: 100, sort: 'name' } })
    bloggers.value = res.items
  }
}

const form = reactive({ bloggerId: 0, publishDate: new Date().toISOString().slice(0, 10), costRub: 50_000 })
const addError = ref('')
const createdPromo = ref('')
async function addIntegration() {
  addError.value = ''
  try {
    const res = await $fetch<{ promoCode: string }>(`/api/campaigns/${route.params.id}/integrations`, { method: 'POST', body: form })
    createdPromo.value = res.promoCode
    await refresh()
  } catch (e) {
    addError.value = apiError(e)
  }
}
function closeAdd() {
  showAdd.value = false
  createdPromo.value = ''
}
</script>

<template>
  <div>
    <NuxtLink to="/campaigns" class="mb-4 inline-block text-sm text-muted hover:text-brand">← Все кампании</NuxtLink>
    <p v-if="error" class="card p-8 text-center text-muted">Кампания не найдена</p>

    <template v-else-if="c">
      <PageHeader :title="c.title" :subtitle="`${c.product} · ${formatDate(c.startDate)} — ${formatDate(c.endDate)}`">
        <div class="flex items-center gap-3">
          <CampaignBadge :status="c.status" />
          <button class="btn btn-primary" @click="openAdd">Добавить интеграцию</button>
        </div>
      </PageHeader>

      <section class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
        <KpiCard label="Бюджет" :value="formatRub(c.budgetRub)" />
        <KpiCard label="Потрачено" :value="formatRub(c.costRub)" />
        <KpiCard label="Выручка" :value="formatRub(c.revenueRub)" :hint="`${formatNumber(c.orders)} заказов`" />
        <KpiCard label="ДРР" :value="formatPercent(c.drr)" hint="доля рекламных расходов в выручке" />
      </section>

      <div v-if="c.daily.length" class="card mb-6 p-5">
        <h2 class="mb-4 font-semibold">Выручка по дням</h2>
        <DailyChart :daily="c.daily" />
      </div>

      <div class="card overflow-hidden">
        <h2 class="px-5 pt-5 font-semibold">Интеграции</h2>
        <div class="overflow-x-auto">
          <table class="table-base mt-2 min-w-[760px]">
            <thead>
              <tr><th>Блогер</th><th>Публикация</th><th>Промокод / UTM</th><th class="text-right">Стоимость</th><th class="text-right">Выручка</th><th class="cursor-help text-right" :title="DRR_HINT">ДРР</th></tr>
            </thead>
            <tbody>
              <tr v-for="i in c.integrationList" :key="i.id" class="hover:bg-canvas">
                <td>
                  <NuxtLink :to="`/bloggers/${i.bloggerId}`" class="font-medium hover:text-brand">{{ i.blogger }}</NuxtLink>
                  <p class="flex items-center gap-1.5 text-xs text-muted">
                    <span class="size-1.5 rounded-full" :style="{ background: platformColors[i.platform] }" />@{{ i.handle }}
                  </p>
                </td>
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
            </tbody>
          </table>
        </div>
      </div>
    </template>

    <AppDialog v-if="showAdd" title="Новая интеграция" @close="closeAdd">
      <div v-if="createdPromo" class="space-y-4">
        <p class="text-sm">Интеграция добавлена. Промокод для блогера:</p>
        <p class="rounded-xl bg-brand-soft px-4 py-3 text-center font-mono text-lg font-semibold text-brand">{{ createdPromo }}</p>
        <div class="flex justify-end"><button class="btn btn-primary" @click="closeAdd">Готово</button></div>
      </div>
      <form v-else class="grid gap-4" @submit.prevent="addIntegration">
        <label class="grid gap-1 text-sm">Блогер
          <select v-model.number="form.bloggerId" class="input" required>
            <option :value="0" disabled>Выберите блогера</option>
            <option v-for="b in bloggers" :key="b.id" :value="b.id">{{ b.name }} (@{{ b.handle }}, {{ b.platform.name }})</option>
          </select>
        </label>
        <div class="grid gap-4 sm:grid-cols-2">
          <label class="grid gap-1 text-sm">Дата публикации
            <input v-model="form.publishDate" class="input" type="date" required>
          </label>
          <label class="grid gap-1 text-sm">Стоимость, ₽
            <input v-model.number="form.costRub" class="input" type="number" min="0" step="1000" required>
          </label>
        </div>
        <p class="text-xs text-muted">Промокод и UTM-метка сгенерируются автоматически.</p>
        <p v-if="addError" class="text-sm text-bad" role="alert">{{ addError }}</p>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-ghost" @click="closeAdd">Отмена</button>
          <button class="btn btn-primary" :disabled="!form.bloggerId">Добавить</button>
        </div>
      </form>
    </AppDialog>
  </div>
</template>
