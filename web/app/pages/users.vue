<script setup lang="ts">
import type { User, UserRole } from '~/types/api'

const { data: users, refresh } = await useFetch<User[]>('/api/users', { default: () => [] })

const statusMeta: Record<User['status'], { label: string; cls: string }> = {
  active: { label: 'Активен', cls: 'bg-good-soft text-good' },
  invited: { label: 'Приглашён', cls: 'bg-warn-soft text-warn' },
  blocked: { label: 'Заблокирован', cls: 'bg-canvas text-muted' },
}

const roleRights: { role: UserRole; rights: string }[] = [
  { role: 'admin', rights: 'всё, включая пользователей и подключения' },
  { role: 'marketer', rights: 'кампании, бюджеты, аналитика' },
  { role: 'manager', rights: 'блогеры и интеграции, промокоды' },
  { role: 'analyst', rights: 'дашборд и отчёты, только чтение' },
  { role: 'finance', rights: 'расходы и закрывающие документы' },
]

const initials = (name: string) => name.split(' ').map((p) => p[0]).join('').slice(0, 2)

const showInvite = ref(false)
const form = reactive({ name: '', email: '', role: 'manager' as UserRole })
const error = ref('')
const saving = ref(false)
async function invite() {
  saving.value = true
  error.value = ''
  try {
    await $fetch('/api/users', { method: 'POST', body: form })
    showInvite.value = false
    Object.assign(form, { name: '', email: '', role: 'manager' })
    await refresh()
  } catch (e) {
    error.value = apiError(e)
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div>
    <PageHeader title="Пользователи" :subtitle="`${users.length} сотрудников с доступом к CRM`">
      <button class="btn btn-primary" @click="showInvite = true">Пригласить</button>
    </PageHeader>

    <div class="grid gap-6 xl:grid-cols-[1fr_20rem]">
      <div class="card overflow-hidden">
        <div class="overflow-x-auto">
          <table class="table-base min-w-[640px]">
            <thead>
              <tr><th>Сотрудник</th><th>Роль</th><th>Статус</th><th>Был в CRM</th></tr>
            </thead>
            <tbody>
              <tr v-for="u in users" :key="u.id" class="hover:bg-canvas" :class="{ 'text-muted': u.status === 'blocked' }">
                <td>
                  <div class="flex items-center gap-3">
                    <span class="grid size-9 shrink-0 place-items-center rounded-full bg-brand-soft text-xs font-semibold text-brand">{{ initials(u.name) }}</span>
                    <div class="min-w-0">
                      <p class="font-medium">{{ u.name }}</p>
                      <p class="truncate text-xs text-muted">{{ u.email }}</p>
                    </div>
                  </div>
                </td>
                <td>{{ roleLabels[u.role] }}</td>
                <td><span class="inline-flex rounded-full px-2.5 py-0.5 text-xs font-semibold" :class="statusMeta[u.status].cls">{{ statusMeta[u.status].label }}</span></td>
                <td class="whitespace-nowrap">
                  <span v-if="formatAgo(u.lastSeenAt) === 'в сети'" class="inline-flex items-center gap-1.5 text-good"><span class="size-1.5 rounded-full bg-good" />в сети</span>
                  <template v-else>{{ formatAgo(u.lastSeenAt) }}</template>
                </td>
              </tr>
            </tbody>
          </table>
        </div>
      </div>

      <aside class="card h-fit p-5">
        <h2 class="mb-3 font-semibold">Роли</h2>
        <dl class="space-y-3 text-sm">
          <div v-for="r in roleRights" :key="r.role">
            <dt class="font-medium">{{ roleLabels[r.role] }}</dt>
            <dd class="text-xs text-muted">{{ r.rights }}</dd>
          </div>
        </dl>
      </aside>
    </div>

    <AppDialog v-if="showInvite" title="Пригласить сотрудника" @close="showInvite = false">
      <form class="grid gap-4" @submit.prevent="invite">
        <label class="grid gap-1 text-sm">Имя и фамилия
          <input v-model="form.name" class="input" required maxlength="120">
        </label>
        <label class="grid gap-1 text-sm">Email
          <input v-model="form.email" class="input" type="email" required>
        </label>
        <label class="grid gap-1 text-sm">Роль
          <select v-model="form.role" class="input">
            <option v-for="r in roleRights" :key="r.role" :value="r.role">{{ roleLabels[r.role] }}</option>
          </select>
        </label>
        <p class="text-xs text-muted">В демо письмо не отправляется — сотрудник появится в списке со статусом «Приглашён».</p>
        <p v-if="error" class="text-sm text-bad" role="alert">{{ error }}</p>
        <div class="flex justify-end gap-2">
          <button type="button" class="btn btn-ghost" @click="showInvite = false">Отмена</button>
          <button class="btn btn-primary" :disabled="saving">Пригласить</button>
        </div>
      </form>
    </AppDialog>
  </div>
</template>
