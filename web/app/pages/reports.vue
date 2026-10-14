<script setup lang="ts">
import type { Preview, ReportRow, RowStatus } from '~/types/api'

const statusMeta: Record<RowStatus, { label: string; cls: string; hint: string }> = {
  new: { label: 'Новая', cls: 'bg-good-soft text-good', hint: 'будет записана' },
  duplicate: { label: 'Уже загружено', cls: 'bg-canvas text-muted', hint: 'день не перезаписываем' },
  unknown_promo: { label: 'Промокод не найден', cls: 'bg-warn-soft text-warn', hint: 'нет такой интеграции' },
  invalid: { label: 'Ошибка', cls: 'bg-bad-soft text-bad', hint: 'строку не прочитать' },
}
const order: RowStatus[] = ['new', 'duplicate', 'unknown_promo', 'invalid']

const fileName = ref('')
const preview = ref<Preview | null>(null)
const filter = ref<RowStatus | ''>('')
const error = ref('')
const busy = ref(false)
const result = ref<{ inserted: number; skipped: number } | null>(null)
const dragging = ref(false)

const rows = computed(() => (preview.value?.rows ?? []).filter((r) => !filter.value || r.status === filter.value))

async function upload(file: File | undefined) {
  if (!file) return
  fileName.value = file.name
  error.value = ''
  result.value = null
  preview.value = null
  busy.value = true
  try {
    const body = new FormData()
    body.append('file', file)
    preview.value = await $fetch<Preview>('/api/reports/preview', { method: 'POST', body })
  } catch (e) {
    error.value = apiError(e)
  } finally {
    busy.value = false
  }
}

async function apply() {
  if (!preview.value) return
  busy.value = true
  error.value = ''
  try {
    // Отправляем исходные строки: сервер заново проверит дубли и промокоды,
    // статусы из превью ему не нужны.
    const payload: ReportRow[] = preview.value.rows.map(({ line, day, promoCode, source, orders, revenueRub, error }) =>
      ({ line, day, promoCode, source, orders, revenueRub, error }))
    result.value = await $fetch('/api/reports/apply', { method: 'POST', body: { rows: payload } })
    preview.value = null
  } catch (e) {
    error.value = apiError(e)
  } finally {
    busy.value = false
  }
}

function onDrop(e: DragEvent) {
  dragging.value = false
  upload(e.dataTransfer?.files[0])
}
</script>

<template>
  <div>
    <PageHeader title="Загрузка отчёта Wildberries" subtitle="Продажи по промокодам: сначала превью, потом запись. Уже загруженные дни не перезаписываются.">
      <a href="/api/reports/sample.xlsx" class="btn btn-ghost" download>
        <svg class="size-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" aria-hidden="true"><path d="M12 4v11m0 0-4-4m4 4 4-4M5 19h14" /></svg>
        Скачать пример отчёта
      </a>
    </PageHeader>

    <label
      class="card mb-6 flex cursor-pointer flex-col items-center gap-2 border-2 border-dashed p-10 text-center transition-colors"
      :class="dragging ? 'border-brand bg-brand-soft' : 'hover:border-brand'"
      @dragover.prevent="dragging = true"
      @dragleave="dragging = false"
      @drop.prevent="onDrop"
    >
      <svg class="size-10 text-brand" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
        <path d="M12 15V4m0 0L8 8m4-4 4 4M5 15v4h14v-4" />
      </svg>
      <span class="font-medium">{{ busy ? 'Обрабатываем…' : fileName || 'Перетащите XLSX или нажмите, чтобы выбрать' }}</span>
      <span class="text-xs text-muted">Колонки: Дата, Промокод, Заказы, Выручка (и необязательный Источник)</span>
      <input type="file" accept=".xlsx" class="sr-only" @change="upload(($event.target as HTMLInputElement).files?.[0])">
    </label>

    <p v-if="error" class="mb-4 rounded-xl bg-bad-soft px-4 py-3 text-sm text-bad" role="alert">{{ error }}</p>

    <p v-if="result" class="mb-4 rounded-xl bg-good-soft px-4 py-3 text-sm text-good" role="status">
      Записано строк: <b>{{ result.inserted }}</b>, пропущено: {{ result.skipped }}. Цифры уже на
      <NuxtLink to="/" class="font-semibold underline">дашборде</NuxtLink>.
    </p>

    <template v-if="preview">
      <div class="mb-4 flex flex-wrap items-center gap-2">
        <button
          class="rounded-full border px-3 py-1.5 text-sm"
          :class="filter === '' ? 'border-ink bg-ink text-white' : 'border-line bg-white'"
          @click="filter = ''"
        >
          Все · {{ preview.rows.length }}
        </button>
        <button
          v-for="s in order"
          :key="s"
          class="rounded-full border px-3 py-1.5 text-sm"
          :class="filter === s ? 'border-ink bg-ink text-white' : 'border-line bg-white'"
          @click="filter = s"
        >
          {{ statusMeta[s].label }} · {{ preview.summary[s] }}
        </button>
        <span class="flex-1" />
        <button class="btn btn-primary" :disabled="busy || !preview.summary.new" @click="apply">
          Записать {{ preview.summary.new }} {{ preview.summary.new === 1 ? 'строку' : 'строк' }}
        </button>
      </div>

      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="table-base min-w-[760px]">
            <thead>
              <tr><th>Строка</th><th>Дата</th><th>Промокод</th><th>Блогер</th><th class="text-right">Заказы</th><th class="text-right">Выручка</th><th>Статус</th></tr>
            </thead>
            <tbody>
              <tr v-for="r in rows" :key="r.line">
                <td class="text-muted tabular-nums">{{ r.line }}</td>
                <td class="whitespace-nowrap">{{ r.day ? formatDate(r.day) : '—' }}</td>
                <td><code class="rounded bg-canvas px-1.5 py-0.5 text-xs">{{ r.promoCode || '—' }}</code></td>
                <td>{{ r.blogger ?? '—' }}</td>
                <td class="text-right tabular-nums">{{ r.error ? '—' : r.orders }}</td>
                <td class="text-right tabular-nums">{{ r.error ? '—' : formatRub(r.revenueRub) }}</td>
                <td>
                  <span class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold" :class="statusMeta[r.status].cls">{{ statusMeta[r.status].label }}</span>
                  <p class="mt-1 text-xs text-muted">{{ r.error ?? statusMeta[r.status].hint }}</p>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>
    </template>
  </div>
</template>
