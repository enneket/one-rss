<script setup lang="ts">
import { onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import Sidebar from '@/components/sidebar/Sidebar.vue'
import ArticleList from '@/components/article/ArticleList.vue'
import ArticleDetail from '@/components/article/ArticleDetail.vue'
import axios from 'axios'

const store = useAppStore()
const { locale } = useI18n()

onMounted(async () => {
  // 从 API 加载设置
  try {
    const response = await axios.get('/api/settings')
    const settings = response.data
    
    // 应用语言设置
    if (settings.language) {
      locale.value = settings.language
      localStorage.setItem('language', settings.language)
    }
    
    // 应用主题设置
    if (settings.theme) {
      const html = document.documentElement
      if (settings.theme === 'dark') {
        html.classList.add('dark')
      } else if (settings.theme === 'light') {
        html.classList.remove('dark')
      } else {
        if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
          html.classList.add('dark')
        } else {
          html.classList.remove('dark')
        }
      }
      localStorage.setItem('theme', settings.theme)
    }
  } catch (error) {
    console.error('Failed to load settings:', error)
  }
  
  await store.fetchFeeds()
  await store.fetchArticles()
  await store.fetchUnreadCounts()
})
</script>

<template>
  <div class="flex h-screen overflow-hidden bg-[var(--bg-primary)]">
    <!-- Sidebar -->
    <Sidebar class="w-64 flex-shrink-0 border-r border-[var(--border-color)]" />
    
    <!-- Article List -->
    <ArticleList class="w-96 flex-shrink-0 border-r border-[var(--border-color)]" />
    
    <!-- Article Detail -->
    <ArticleDetail class="flex-1" />
  </div>
</template>
