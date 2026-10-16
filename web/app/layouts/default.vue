<script setup lang="ts">
const groups = [
  {
    title: '',
    items: [{ to: '/', label: 'Дашборд', icon: 'M4 13h6V4H4zm0 7h6v-5H4zm10 0h6v-9h-6zm0-16v5h6V4z' }],
  },
  {
    title: 'Работа с блогерами',
    items: [
      { to: '/campaigns', label: 'Кампании', icon: 'M4 10v4h3l5 4V6L7 10zm12-2a5 5 0 0 1 0 8M18.5 5.5a8.500 8.500 0 0 1 0 13' },
      { to: '/integrations', label: 'Интеграции', icon: 'M10 14a4 4 0 0 0 5.66 0l3-3a4 4 0 0 0-5.66-5.66l-1 1M14 10a4 4 0 0 0-5.66 0l-3 3a4 4 0 0 0 5.66 5.66l1-1' },
      { to: '/bloggers', label: 'Блогеры', icon: 'M12 12a4 4 0 1 0 0-8 4 4 0 0 0 0 8m-7 8a7 7 0 0 1 14 0' },
    ],
  },
  {
    title: 'Данные',
    items: [{ to: '/ozon', label: 'Отчёты Ozon', icon: 'M7 3h7l5 5v13H7zm7 0v5h5M10 13h6m-6 4h6' }],
  },
  {
    title: 'Администрирование',
    items: [
      { to: '/users', label: 'Пользователи', icon: 'M9 11a4 4 0 1 0 0-8 4 4 0 0 0 0 8m-7 9a7 7 0 0 1 14 0M16 3.5a4 4 0 0 1 0 7.5M22 20a7 7 0 0 0-4-6.3' },
      { to: '/settings', label: 'Настройки', icon: 'M4 6h9m4 0h3M4 12h3m4 0h9M4 18h11m4 0h1M15 4v4M9 10v4m8 2v4' },
    ],
  },
]
const route = useRoute()
const isActive = (to: string) => (to === '/' ? route.path === '/' : route.path.startsWith(to))
</script>

<template>
  <div class="min-h-screen md:flex">
    <aside class="border-b border-line bg-white md:sticky md:top-0 md:flex md:h-screen md:w-64 md:shrink-0 md:flex-col md:border-b-0 md:border-r">
      <div class="flex items-center gap-3 px-5 py-4 md:py-6">
        <span class="grid size-9 place-items-center rounded-xl bg-brand text-sm font-bold text-white">BC</span>
        <div class="leading-tight">
          <p class="font-semibold">Blogger CRM</p>
          <p class="text-xs text-muted">демо-версия</p>
        </div>
      </div>

      <nav class="flex gap-1 overflow-x-auto px-3 pb-3 md:flex-1 md:flex-col md:gap-5 md:overflow-y-auto md:pb-4">
        <div v-for="g in groups" :key="g.title" class="flex gap-1 md:flex-col">
          <p v-if="g.title" class="hidden px-3 pb-1 text-[11px] font-semibold uppercase tracking-wider text-muted/80 md:block">{{ g.title }}</p>
          <NuxtLink
            v-for="item in g.items"
            :key="item.to"
            :to="item.to"
            class="flex shrink-0 items-center gap-3 rounded-xl px-3 py-2 text-sm font-medium transition-colors"
            :class="isActive(item.to) ? 'bg-brand-soft text-brand' : 'text-muted hover:bg-canvas hover:text-ink'"
          >
            <svg class="size-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
              <path :d="item.icon" />
            </svg>
            {{ item.label }}
          </NuxtLink>
        </div>
      </nav>

      <div class="hidden border-t border-line p-4 md:block">
        <div class="flex items-center gap-3">
          <span class="grid size-9 shrink-0 place-items-center rounded-full bg-brand-soft text-sm font-semibold text-brand">АС</span>
          <div class="min-w-0 leading-tight">
            <p class="truncate text-sm font-medium">Анна Смирнова</p>
            <p class="truncate text-xs text-muted">Администратор</p>
          </div>
        </div>
        <p class="mt-3 text-xs leading-relaxed text-muted">
          Все блогеры, товары и цифры выдуманы. Данные пересоздаются при каждом запуске.
        </p>
      </div>
    </aside>
    <main class="min-w-0 flex-1 px-4 py-6 md:px-8 md:py-8">
      <slot />
    </main>
  </div>
</template>
