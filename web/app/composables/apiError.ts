// Достаёт текст ошибки из ответа Go API ({"error": "..."}).
export function apiError(e: unknown): string {
  const data = (e as { data?: { error?: string } })?.data
  return data?.error ?? 'Не удалось выполнить запрос'
}
