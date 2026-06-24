import { createApp } from 'vue'
import { createPinia } from 'pinia'
import { createI18n } from 'vue-i18n'
import App from './App.vue'
import router from './router'
import en from './i18n/locales/en'
import zh from './i18n/locales/zh'
import { vLazyload } from './directives/lazyload'
import './style.css'

// 初始化主题
const theme = localStorage.getItem('theme') || 'auto'
if (theme === 'dark' || (theme === 'auto' && window.matchMedia('(prefers-color-scheme: dark)').matches)) {
  document.documentElement.classList.add('dark')
}

const i18n = createI18n({
  legacy: false,
  locale: localStorage.getItem('language') || 'en-US',
  fallbackLocale: 'en-US',
  messages: {
    'en-US': en,
    'zh-CN': zh
  }
})

const app = createApp(App)
app.directive('lazyload', vLazyload)
app.use(createPinia())
app.use(router)
app.use(i18n)
app.mount('#app')
