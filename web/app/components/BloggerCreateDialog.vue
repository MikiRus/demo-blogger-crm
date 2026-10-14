<script setup lang="ts">
import type { Platform } from '~/types/api'

const props = defineProps<{ platforms: Platform[] }>()
const emit = defineEmits<{ close: []; created: [id: number] }>()

const form = reactive({
  name: '',
  handle: '',
  platform: props.platforms[0]?.code ?? '',
  followers: 50_000,
  avgReach: 10_000,
  category: 'Кулинария',
})
const error = ref('')
const saving = ref(false)

async function submit() {
  saving.value = true
  error.value = ''
  try {
    const { id } = await $fetch<{ id: number }>('/api/bloggers', { method: 'POST', body: form })
    emit('created', id)
  } catch (e) {
    error.value = apiError(e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <AppDialog title="Новый блогер" @close="emit('close')">
    <form class="grid gap-4 sm:grid-cols-2" @submit.prevent="submit">
      <label class="grid gap-1 text-sm sm:col-span-2">Имя
        <input v-model="form.name" class="input" required maxlength="120">
      </label>
      <label class="grid gap-1 text-sm">Ник
        <input v-model="form.handle" class="input" required maxlength="64" placeholder="@nickname">
      </label>
      <label class="grid gap-1 text-sm">Площадка
        <select v-model="form.platform" class="input">
          <option v-for="p in platforms" :key="p.code" :value="p.code">{{ p.name }}</option>
        </select>
      </label>
      <label class="grid gap-1 text-sm">Подписчики
        <input v-model.number="form.followers" class="input" type="number" min="0" required>
      </label>
      <label class="grid gap-1 text-sm">Средний охват
        <input v-model.number="form.avgReach" class="input" type="number" min="0" required>
      </label>
      <label class="grid gap-1 text-sm sm:col-span-2">Тематика
        <input v-model="form.category" class="input" required>
      </label>
      <p v-if="error" class="text-sm text-bad sm:col-span-2" role="alert">{{ error }}</p>
      <div class="flex justify-end gap-2 sm:col-span-2">
        <button type="button" class="btn btn-ghost" @click="emit('close')">Отмена</button>
        <button class="btn btn-primary" :disabled="saving">Сохранить</button>
      </div>
    </form>
  </AppDialog>
</template>
