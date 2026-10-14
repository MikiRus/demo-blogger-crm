<script setup lang="ts">
import type { BloggerDetails } from '~/types/api'

const route = useRoute()
const { data: b, error } = await useFetch<BloggerDetails>(() => `/api/bloggers/${route.params.id}`)
</script>

<template>
  <div>
    <NuxtLink to="/bloggers" class="mb-4 inline-block text-sm text-muted hover:text-brand">← Все блогеры</NuxtLink>
    <p v-if="error" class="card p-8 text-center text-muted">Блогер не найден</p>

    <template v-else-if="b">
      <PageHeader :title="b.name" :subtitle="`@${b.handle} · ${b.platform.name} · ${b.category}`">
        <RoasBadge :roas="b.roas" />
      </PageHeader>

      <section class="mb-6 grid grid-cols-2 gap-4 lg:grid-cols-4">
        <KpiCard label="Подписчики" :value="formatNumber(b.followers)" :hint="`охват поста ≈ ${formatCompact(b.avgReach)}`" />
        <KpiCard label="Потрачено" :value="formatRub(b.costRub)" :hint="`${b.integrations} интеграций`" />
        <KpiCard label="Выручка" :value="formatRub(b.revenueRub)" :hint="`${formatNumber(b.orders)} заказов`" />
        <KpiCard label="ДРР" :value="formatPercent(b.drr)" />
      </section>

      <div class="card overflow-hidden">
        <h2 class="px-5 pt-5 font-semibold">История интеграций</h2>
        <div class="overflow-x-auto">
          <table class="table-base mt-2 min-w-[680px]">
            <thead>
              <tr><th>Кампания</th><th>Дата</th><th>Промокод</th><th class="text-right">Расходы</th><th class="text-right">Выручка</th><th class="text-right">ROAS</th></tr>
            </thead>
            <tbody>
              <tr v-for="h in b.history" :key="h.id" class="hover:bg-canvas">
                <td><NuxtLink :to="`/campaigns/${h.campaignId}`" class="font-medium hover:text-brand">{{ h.campaign }}</NuxtLink></td>
                <td class="whitespace-nowrap">
                  {{ formatDate(h.publishDate) }}
                  <span v-if="h.status === 'planned'" class="ml-1 rounded-full bg-canvas px-2 py-0.5 text-xs text-muted">план</span>
                </td>
                <td><code class="rounded bg-canvas px-1.5 py-0.5 text-xs">{{ h.promoCode }}</code></td>
                <td class="text-right tabular-nums">{{ formatRub(h.costRub) }}</td>
                <td class="text-right tabular-nums">{{ formatRub(h.revenueRub) }}</td>
                <td class="text-right"><RoasBadge :roas="h.roas" /></td>
              </tr>
              <tr v-if="!b.history.length"><td colspan="6" class="py-10 text-center text-muted">Интеграций пока не было</td></tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>
