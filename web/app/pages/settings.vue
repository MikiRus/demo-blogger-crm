<script setup lang="ts">
const connections = [
  {
    name: 'Ozon Seller API',
    about: 'Заказы с промокодами — ежедневный отчёт за прошедший день',
    detail: 'Client-Id ••••4821 · ежедневно в 06:00 МСК',
    connected: true,
    to: '/ozon',
  },
  {
    name: 'RetailCRM',
    about: 'Заказы сайта, сопоставление по UTM-метке',
    detail: 'shop.retailcrm.ru · каждые 15 минут',
    connected: true,
  },
  {
    name: 'Wildberries',
    about: 'Продажи по промокодам из статистики продавца',
    detail: 'Не подключено',
    connected: false,
  },
]

const thresholds = [
  { label: 'Окупается', range: 'до 25%', cls: 'bg-good-soft text-good' },
  { label: 'На грани', range: '25–40%', cls: 'bg-warn-soft text-warn' },
  { label: 'В минус', range: 'выше 40%', cls: 'bg-bad-soft text-bad' },
]
</script>

<template>
  <div>
    <PageHeader title="Настройки" subtitle="Подключения к площадкам и правила подсчёта" />

    <section class="mb-6">
      <h2 class="mb-3 font-semibold">Подключения</h2>
      <div class="grid gap-4 md:grid-cols-3">
        <div v-for="c in connections" :key="c.name" class="card flex flex-col p-5">
          <div class="mb-2 flex items-center justify-between gap-3">
            <p class="font-semibold">{{ c.name }}</p>
            <span
              class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold"
              :class="c.connected ? 'bg-good-soft text-good' : 'bg-canvas text-muted'"
            >{{ c.connected ? 'Подключено' : 'Выключено' }}</span>
          </div>
          <p class="text-sm">{{ c.about }}</p>
          <p class="mt-1 text-xs text-muted">{{ c.detail }}</p>
          <div class="mt-4 flex gap-2 pt-1">
            <NuxtLink v-if="c.to" :to="c.to" class="btn btn-ghost">Журнал</NuxtLink>
            <button class="btn btn-ghost" disabled title="В демо недоступно">{{ c.connected ? 'Изменить ключи' : 'Подключить' }}</button>
          </div>
        </div>
      </div>
    </section>

    <section class="grid gap-6 lg:grid-cols-2">
      <div class="card p-5">
        <h2 class="font-semibold">ДРР — доля рекламных расходов</h2>
        <p class="mt-1 text-sm text-muted">
          Расходы на интеграции ÷ выручка по их промокодам и UTM × 100%. Расходы относятся к дню публикации,
          выручка — к дню продажи. Чем ниже ДРР, тем выгоднее блогер.
        </p>
        <div class="mt-4 flex flex-wrap gap-2">
          <span v-for="t in thresholds" :key="t.label" class="inline-flex rounded-full px-3 py-1 text-xs font-semibold" :class="t.cls">
            {{ t.label }} · {{ t.range }}
          </span>
        </div>
      </div>
      <div class="card p-5">
        <h2 class="font-semibold">Атрибуция продаж</h2>
        <dl class="mt-3 grid gap-3 text-sm">
          <div class="flex justify-between gap-4"><dt class="text-muted">Окно продаж после публикации</dt><dd class="font-medium">14 дней</dd></div>
          <div class="flex justify-between gap-4"><dt class="text-muted">Формат промокода</dt><dd><code class="rounded bg-canvas px-1.5 py-0.5 text-xs">НИК-XXXX</code></dd></div>
          <div class="flex justify-between gap-4"><dt class="text-muted">UTM-метка</dt><dd><code class="rounded bg-canvas px-1.5 py-0.5 text-xs">crm_&lt;кампания&gt;_&lt;ник&gt;</code></dd></div>
          <div class="flex justify-between gap-4"><dt class="text-muted">Валюта</dt><dd class="font-medium">₽, целые рубли</dd></div>
        </dl>
      </div>
    </section>
  </div>
</template>
