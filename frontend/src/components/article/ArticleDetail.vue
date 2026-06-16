<script setup lang="ts">
import { ref, computed, watch } from 'vue'
import { useAppStore } from '@/stores/app'
import { useI18n } from 'vue-i18n'
import { 
  PhStar, 
  PhClock, 
  PhEye, 
  PhEyeSlash, 
  PhTranslate, 
  PhArticle,
  PhArrowSquareOut,
  PhExport
} from '@phosphor-icons/vue'
import axios from 'axios'

const store = useAppStore()
const { t } = useI18n()
const articleContent = ref('')
const isLoadingContent = ref(false)

const currentArticle = computed(() => {
  if (!store.currentArticleId) return null
  return store.articles.find(a => a.id === store.currentArticleId) || null
})

// 当文章 ID 变化时获取内容
watch(() => store.currentArticleId, async (newId) => {
  if (!newId) {
    articleContent.value = ''
    return
  }
  
  isLoadingContent.value = true
  try {
    const response = await axios.get(`/api/articles/content?id=${newId}`)
    articleContent.value = response.data.content || ''
  } catch (error) {
    console.error('Failed to fetch article content:', error)
    articleContent.value = ''
  } finally {
    isLoadingContent.value = false
  }
})

function formatDate(dateStr: string): string {
  try {
    return new Date(dateStr).toLocaleString()
  } catch {
    return ''
  }
}

function openOriginal() {
  if (currentArticle.value?.url) {
    window.open(currentArticle.value.url, '_blank')
  }
}
</script>

<template>
  <div class="flex flex-col h-full bg-[var(--bg-primary)]">
    <!-- Empty State -->
    <div v-if="!currentArticle" class="flex-1 flex items-center justify-center">
      <div class="text-center text-[var(--text-tertiary)]">
        <PhArticle :size="64" class="mx-auto mb-4 opacity-50" />
        <p>{{ t('article.noArticles') }}</p>
      </div>
    </div>

    <!-- Article Detail -->
    <template v-else>
      <!-- Header -->
      <div class="p-4 border-b border-[var(--border-color)]">
        <div class="flex items-center justify-between mb-3">
          <div class="flex items-center gap-2">
            <span class="text-sm text-[var(--text-tertiary)]">
              {{ currentArticle.feed_title }}
            </span>
            <span v-if="currentArticle.author" class="text-sm text-[var(--text-tertiary)]">
              · {{ currentArticle.author }}
            </span>
          </div>
          <span class="text-sm text-[var(--text-tertiary)]">
            {{ formatDate(currentArticle.published_at) }}
          </span>
        </div>

        <h1 class="text-xl font-bold text-[var(--text-primary)] mb-2">
          {{ currentArticle.title }}
        </h1>

        <p v-if="currentArticle.translated_title" class="text-lg text-[var(--text-secondary)] mb-3">
          {{ currentArticle.translated_title }}
        </p>

        <!-- Actions -->
        <div class="flex items-center gap-2">
          <button
            @click="store.toggleFavorite(currentArticle!.id)"
            :class="[
              'flex items-center gap-1 px-3 py-1.5 rounded-lg transition-colors',
              currentArticle.is_favorite 
                ? 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900/30 dark:text-yellow-400' 
                : 'bg-[var(--bg-secondary)] text-[var(--text-secondary)] hover:bg-[var(--bg-tertiary)]'
            ]"
          >
            <PhStar :size="16" :weight="currentArticle.is_favorite ? 'fill' : 'regular'" />
            <span class="text-sm">{{ currentArticle.is_favorite ? t('article.unfavorite') : t('article.favorite') }}</span>
          </button>

          <button
            @click="store.toggleReadLater(currentArticle!.id)"
            :class="[
              'flex items-center gap-1 px-3 py-1.5 rounded-lg transition-colors',
              currentArticle.is_read_later 
                ? 'bg-blue-100 text-blue-700 dark:bg-blue-900/30 dark:text-blue-400' 
                : 'bg-[var(--bg-secondary)] text-[var(--text-secondary)] hover:bg-[var(--bg-tertiary)]'
            ]"
          >
            <PhClock :size="16" :weight="currentArticle.is_read_later ? 'fill' : 'regular'" />
            <span class="text-sm">{{ t('article.readLater') }}</span>
          </button>

          <button
            @click="store.toggleHide(currentArticle!.id)"
            :class="[
              'flex items-center gap-1 px-3 py-1.5 rounded-lg transition-colors',
              currentArticle.is_hidden 
                ? 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400' 
                : 'bg-[var(--bg-secondary)] text-[var(--text-secondary)] hover:bg-[var(--bg-tertiary)]'
            ]"
          >
            <component :is="currentArticle.is_hidden ? PhEye : PhEyeSlash" :size="16" />
            <span class="text-sm">{{ currentArticle.is_hidden ? t('article.unhide') : t('article.hide') }}</span>
          </button>

          <button
            class="flex items-center gap-1 px-3 py-1.5 rounded-lg bg-[var(--bg-secondary)] text-[var(--text-secondary)] hover:bg-[var(--bg-tertiary)] transition-colors"
          >
            <PhTranslate :size="16" />
            <span class="text-sm">{{ t('article.translate') }}</span>
          </button>

          <button
            class="flex items-center gap-1 px-3 py-1.5 rounded-lg bg-[var(--bg-secondary)] text-[var(--text-secondary)] hover:bg-[var(--bg-tertiary)] transition-colors"
          >
            <PhArticle :size="16" />
            <span class="text-sm">{{ t('article.summarize') }}</span>
          </button>

          <button
            @click="openOriginal"
            class="flex items-center gap-1 px-3 py-1.5 rounded-lg bg-[var(--bg-secondary)] text-[var(--text-secondary)] hover:bg-[var(--bg-tertiary)] transition-colors"
          >
            <PhArrowSquareOut :size="16" />
            <span class="text-sm">{{ t('article.openOriginal') }}</span>
          </button>
        </div>
      </div>

      <!-- Summary -->
      <div v-if="currentArticle.summary" class="p-4 bg-[var(--bg-secondary)] border-b border-[var(--border-color)]">
        <h3 class="text-sm font-semibold text-[var(--text-primary)] mb-2">
          {{ t('article.summarize') }}
        </h3>
        <p class="text-sm text-[var(--text-secondary)]">
          {{ currentArticle.summary }}
        </p>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-6">
        <div v-if="currentArticle.image_url" class="mb-6">
          <img 
            :src="currentArticle.image_url" 
            class="w-full h-auto rounded-lg"
            @error="($event.target as HTMLImageElement).style.display = 'none'"
          />
        </div>

        <div class="prose prose-lg max-w-none dark:prose-invert">
          <div v-if="isLoadingContent" class="text-center py-8">
            <p class="text-[var(--text-secondary)]">{{ t('common.loading') }}</p>
          </div>
          <div v-else-if="articleContent" v-html="articleContent"></div>
          <div v-else class="text-center py-8">
            <p class="text-[var(--text-secondary)] mb-4">{{ t('article.noContent') }}</p>
            <a 
              v-if="currentArticle.url" 
              :href="currentArticle.url" 
              target="_blank"
              class="inline-flex items-center gap-2 px-4 py-2 bg-primary-500 text-white rounded-lg hover:bg-primary-600 transition-colors"
            >
              <PhArrowSquareOut :size="16" />
              {{ t('article.openOriginal') }}
            </a>
          </div>
        </div>
      </div>
    </template>
  </div>
</template>
