// BFF: браузер ходит только в Nuxt, а Nitro проксирует /api/* в Go API.
// Адрес бэкенда не попадает в клиент, а CORS не нужен.
export default defineEventHandler((event) => {
  const { apiBase } = useRuntimeConfig(event)
  const path = getRouterParam(event, 'path') ?? ''
  const query = getRequestURL(event).search
  return proxyRequest(event, `${apiBase}/api/${path}${query}`)
})
