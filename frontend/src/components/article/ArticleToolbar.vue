<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { 
  PhStar, 
  PhClock, 
  PhEye, 
  PhEyeSlash, 
  PhTranslate, 
  PhArticle,
  PhArrowSquareOut,
  PhExport,
  PhChatCircle
} from '@phosphor-icons/vue'

const { t } = useI18n()
const store = useAppStore()

const props = defineProps<{
  articleId: number
  isFavorite: boolean
  isReadLater: boolean
  isHidden: boolean
}>()

const emit = defineEmits<{
  (e: 'translate'): void
  (e: 'summarize'): void
  (e: 'chat'): void
  (e: 'export'): void
}>()
</script>

<template>
  <div class="flex items-center gap-1">
    <button
      @click="store.toggleFavorite(articleId)"
      :class="[
        'p-2 rounded-lg transition-colors',
        isFavorite 
          ? 'text-yellow-500 bg-yellow-50 dark:bg-yellow-900/20' 
          : 'text-gray-500 hover:text-yellow-500 hover:bg-gray-100 dark:hover:bg-gray-700'
      ]"
      :title="isFavorite ? t('article.unfavorite') : t('article.favorite')"
    >
      <PhStar :size="18" :weight="isFavorite ? 'fill' : 'regular'" />
    </button>

    <button
      @click="store.toggleReadLater(articleId)"
      :class="[
        'p-2 rounded-lg transition-colors',
        isReadLater 
          ? 'text-blue-500 bg-blue-50 dark:bg-blue-900/20' 
          : 'text-gray-500 hover:text-blue-500 hover:bg-gray-100 dark:hover:bg-gray-700'
      ]"
      :title="t('article.readLater')"
    >
      <PhClock :size="18" :weight="isReadLater ? 'fill' : 'regular'" />
    </button>

    <button
      @click="store.toggleHide(articleId)"
      :class="[
        'p-2 rounded-lg transition-colors',
        isHidden 
          ? 'text-red-500 bg-red-50 dark:bg-red-900/20' 
          : 'text-gray-500 hover:text-red-500 hover:bg-gray-100 dark:hover:bg-gray-700'
      ]"
      :title="isHidden ? t('article.unhide') : t('article.hide')"
    >
      <component :is="isHidden ? PhEye : PhEyeSlash" :size="18" />
    </button>

    <div class="w-px h-6 bg-gray-200 dark:bg-gray-700 mx-1" />

    <button
      @click="emit('translate')"
      class="p-2 rounded-lg text-gray-500 hover:text-primary-500 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
      :title="t('article.translate')"
    >
      <PhTranslate :size="18" />
    </button>

    <button
      @click="emit('summarize')"
      class="p-2 rounded-lg text-gray-500 hover:text-primary-500 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
      :title="t('article.summarize')"
    >
      <PhArticle :size="18" />
    </button>

    <button
      @click="emit('chat')"
      class="p-2 rounded-lg text-gray-500 hover:text-primary-500 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
      :title="t('article.chat')"
    >
      <PhChatCircle :size="18" />
    </button>

    <div class="w-px h-6 bg-gray-200 dark:bg-gray-700 mx-1" />

    <button
      @click="emit('export')"
      class="p-2 rounded-lg text-gray-500 hover:text-primary-500 hover:bg-gray-100 dark:hover:bg-gray-700 transition-colors"
      :title="t('article.export')"
    >
      <PhExport :size="18" />
    </button>
  </div>
</template>
