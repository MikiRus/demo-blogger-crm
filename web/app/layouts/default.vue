<script setup lang="ts">
const nav = [
  { to: '/', label: 'Дашборд', icon: 'M4 13h6V4H4zm0 7h6v-5H4zm10 0h6v-9h-6zm0-16v5h6V4z' },
  { to: '/campaigns', label: 'Кампании', icon: 'M4 10v4h3l5 4V6L7 10zm12-2a5 5 0 0 1 0 8M18.5 5.5a8.500 8.500 0 0 1 0 13' },
  { to: '/bloggers', label: 'Блогеры', icon: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8m-7 8a7 7 0 0 1 14 0' },
  { to: '/reports', label: 'Загрузка отчёта', icon: 'M12 15V4m0 0L8 8m4-4 4 4M5 15v4h14v-4' },
]
const route = useRoute()
const isActive = (to: string) => (to === '/' ? route.path === '/' : route.path.startsWith(to))
</script>

<template>
  <div class="min-h-screen md:flex">
    <aside class="border-b border-line bg-white md:sticky md:top-0 md:h-screen md:w-60 md:shrink-0 md:border-b-0 md:border-r">
      <div class="flex items-center gap-3 px-5 py-4 md:py-6">
        <span class="grid size-9 place-items-center rounded-xl bg-brand text-sm font-bold text-white">BC</span>
        <div class="leading-tight">
          <p class="font-semibold">Blogger CRM</p>
          <p class="text-xs text-muted">демо-версия</p>
        </div>
      </div>
      <nav class="flex gap-1 overflow-x-auto px-3 pb-3 md:flex-col md:pb-0">
        <NuxtLink
          v-for="item in nav"
          :key="item.to"
          :to="item.to"
          class="flex shrink-0 items-center gap-3 rounded-xl px-3 py-2.5 text-sm font-medium transition-colors"
          :class="isActive(item.to) ? 'bg-brand-soft text-brand' : 'text-muted hover:bg-canvas hover:text-ink'"
        >
          <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
            <path :d="item.icon" />
          </svg>
          {{ item.label }}
        </NuxtLink>
      </nav>
      <p class="hidden px-5 pt-8 text-xs leading-relaxed text-muted md:block">
        Все блогеры, товары и цифры выдуманы. Данные пересоздаются при каждом запуске.
      </p>
    </aside>
    <main class="min-w-0 flex-1 px-4 py-6 md:px-8 md:py-8">
      <slot />
    </main>
  </div>
</template>
