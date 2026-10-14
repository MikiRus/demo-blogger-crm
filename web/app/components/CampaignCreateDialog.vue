<script setup lang="ts">
const emit = defineEmits<{ close: []; created: [id: number] }>()

const iso = (d: Date) => d.toISOString().slice(0, 10)
const today = new Date()
const form = reactive({
  title: '',
  product: '',
  startDate: iso(today),
  endDate: iso(new Date(today.getTime() + 28 * 86_400_000)),
  budgetRub: 500_000,
})
const error = ref('')
const saving = ref(false)

async function submit() {
  saving.value = true
  error.value = ''
  try {
    const { id } = await $fetch<{ id: number }>('/api/campaigns', { method: 'POST', body: form })
    emit('created', id)
  } catch (e) {
    error.value = apiError(e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <AppDialog title="Новая кампания" @close="emit('close')">
    <form class="grid gap-4 sm:grid-cols-2" @submit.prevent="submit">
      <label class="grid gap-1 text-sm sm:col-span-2">Название
        <input v-model="form.title" class="input" required maxlength="160" placeholder="Весенний запуск">
      </label>
      <label class="grid gap-1 text-sm sm:col-span-2">Товар
        <input v-model="form.product" class="input" required>
      </label>
      <label class="grid gap-1 text-sm">Начало
        <input v-model="form.startDate" class="input" type="date" required>
      </label>
      <label class="grid gap-1 text-sm">Окончание
        <input v-model="form.endDate" class="input" type="date" required>
      </label>
      <label class="grid gap-1 text-sm sm:col-span-2">Бюджет, ₽
        <input v-model.number="form.budgetRub" class="input" type="number" min="0" step="1000" required>
      </label>
      <p v-if="error" class="text-sm text-bad sm:col-span-2" role="alert">{{ error }}</p>
      <div class="flex justify-end gap-2 sm:col-span-2">
        <button type="button" class="btn btn-ghost" @click="emit('close')">Отмена</button>
        <button class="btn btn-primary" :disabled="saving">Создать</button>
      </div>
    </form>
  </AppDialog>
</template>
