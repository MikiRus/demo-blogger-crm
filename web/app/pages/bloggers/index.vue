<script setup lang="ts">
import type { BloggerItem, Platform } from '~/types/api'

const PAGE = 15
const platform = ref('')
const sort = ref('revenue')
const search = ref('')
const q = refDebounced(search, 300)
const page = ref(0)
watch([platform, sort, q], () => (page.value = 0))

const query = computed(() => ({ platform: platform.value, sort: sort.value, q: q.value, limit: PAGE, offset: page.value * PAGE }))
const { data, status } = await useFetch<{ items: BloggerItem[]; total: number }>('/api/bloggers', { query })
const { data: platforms } = await useFetch<Platform[]>('/api/platforms', { default: () => [] })
const pages = computed(() => Math.max(1, Math.ceil((data.value?.total ?? 0) / PAGE)))

const showCreate = ref(false)
</script>

<template>
  <div>
    <PageHeader title="Блогеры" :subtitle="`${data?.total ?? 0} в базе`">
      <button class="btn btn-primary" @click="showCreate = true">Добавить блогера</button>
    </PageHeader>

    <div class="mb-4 grid gap-3 sm:grid-cols-[1fr_auto_auto]">
      <label class="sr-only" for="q">Поиск</label>
      <input id="q" v-model="search" class="input" placeholder="Имя или ник">
      <label class="sr-only" for="platform">Площадка</label>
      <select id="platform" v-model="platform" class="input sm:w-44">
        <option value="">Все площадки</option>
        <option v-for="p in platforms" :key="p.code" :value="p.code">{{ p.name }}</option>
      </select>
      <label class="sr-only" for="sort">Сортировка</label>
      <select id="sort" v-model="sort" class="input sm:w-48">
        <option value="revenue">По выручке</option>
        <option value="roas">По ROAS</option>
        <option value="followers">По подписчикам</option>
        <option value="name">По имени</option>
      </select>
    </div>

    <div class="card overflow-hidden" :class="{ 'opacity-60': status === 'pending' }">
      <div class="overflow-x-auto">
        <table class="table-base min-w-[760px]">
          <thead>
            <tr>
              <th>Блогер</th><th>Площадка</th><th class="text-right">Подписчики</th><th class="text-right">Интеграций</th>
              <th class="text-right">Расходы</th><th class="text-right">Выручка</th><th class="text-right">ROAS</th>
            </tr>
          </thead>
          <tbody>
            <tr v-for="b in data?.items" :key="b.id" class="hover:bg-canvas">
              <td>
                <NuxtLink :to="`/bloggers/${b.id}`" class="font-medium hover:text-brand">{{ b.name }}</NuxtLink>
                <p class="text-xs text-muted">@{{ b.handle }} · {{ b.category }}</p>
              </td>
              <td>
                <span class="inline-flex items-center gap-2">
                  <span class="size-2 rounded-full" :style="{ background: platformColors[b.platform.code] }" />{{ b.platform.name }}
                </span>
              </td>
              <td class="text-right tabular-nums">{{ formatCompact(b.followers) }}</td>
              <td class="text-right tabular-nums">{{ b.integrations }}</td>
              <td class="text-right tabular-nums">{{ formatRub(b.costRub) }}</td>
              <td class="text-right tabular-nums">{{ formatRub(b.revenueRub) }}</td>
              <td class="text-right"><RoasBadge :roas="b.roas" /></td>
            </tr>
            <tr v-if="data && !data.items.length">
              <td colspan="7" class="py-10 text-center text-muted">Никого не нашли</td>
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

    <BloggerCreateDialog v-if="showCreate" :platforms="platforms ?? []" @close="showCreate = false" @created="(id) => navigateTo(`/bloggers/${id}`)" />
  </div>
</template>
