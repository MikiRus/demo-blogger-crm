import tailwindcss from '@tailwindcss/vite'

export default defineNuxtConfig({
  compatibilityDate: '2026-10-01',
  // Как в оригинале: SPA, а Nitro служит BFF-прослойкой перед Go API.
  ssr: false,
  devtools: { enabled: false },
  css: ['~/assets/css/main.css'],
  vite: { plugins: [tailwindcss()] },
  runtimeConfig: {
    // NUXT_API_BASE — адрес Go API, виден только серверу Nitro.
    apiBase: 'http://localhost:8080',
  },
  app: {
    head: {
      htmlAttrs: { lang: 'ru' },
      title: 'Blogger CRM — демо',
      meta: [{ name: 'viewport', content: 'width=device-width, initial-scale=1' }],
      link: [
        { rel: 'preconnect', href: 'https://fonts.googleapis.com' },
        { rel: 'stylesheet', href: 'https://fonts.googleapis.com/css2?family=Onest:wght@400;500;600;700&display=swap' },
      ],
    },
  },
})
