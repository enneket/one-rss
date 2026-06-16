<script setup lang="ts">
import { computed } from 'vue'
import { useAppStore, type Article } from '@/stores/app'
import { useI18n } from 'vue-i18n'
import { PhImage, PhStar, PhClock } from '@phosphor-icons/vue'
import { formatDistanceToNow } from '@/utils/date'

const store = useAppStore()
const { t } = useI18n()

const imageArticles = computed(() => {
  return store.filteredArticles.filter(a => a.image_url)
})

function selectArticle(id: number) {
  store.selectArticle(id)
}

function formatDate(dateStr: string): string {
  try {
    return formatDistanceToNow(new Date(dateStr))
  } catch {
    return ''
  }
}
</script>

<template>
  <div class="h-full overflow-y-auto p-4">
    <!-- Header -->
    <div class="flex items-center justify-between mb-4">
      <h2 class="text-lg font-semibold text-[var(--text-primary)]">
        {{ t('filter.imageGallery') }}
      </h2>
    </div>

    <!-- Grid -->
    <div class="grid grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-4">
      <div
        v-for="article in imageArticles"
        :key="article.id"
        @click="selectArticle(article.id)"
        class="group relative aspect-square rounded-lg overflow-hidden cursor-pointer"
      >
        <!-- Image -->
        <img 
          :src="article.image_url" 
          :alt="article.title"
          class="w-full h-full object-cover transition-transform group-hover:scale-105"
          loading="lazy"
        />

        <!-- Overlay -->
        <div class="absolute inset-0 bg-gradient-to-t from-black/60 to-transparent opacity-0 group-hover:opacity-100 transition-opacity">
          <div class="absolute bottom-0 left-0 right-0 p-3">
            <h3 class="text-sm font-medium text-white line-clamp-2">
              {{ article.title }}
            </h3>
            <div class="flex items-center gap-2 mt-1">
              <span class="text-xs text-white/80">{{ article.feed_title }}</span>
              <span class="text-xs text-white/80">{{ formatDate(article.published_at) }}</span>
            </div>
          </div>

          <!-- Status Icons -->
          <div class="absolute top-2 right-2 flex items-center gap-1">
            <span 
              v-if="article.is_favorite"
              class="p-1 rounded-full bg-yellow-500/80 text-white"
            >
              <PhStar :size="12" weight="fill" />
            </span>
            <span 
              v-if="article.is_read_later"
              class="p-1 rounded-full bg-blue-500/80 text-white"
            >
              <PhClock :size="12" weight="fill" />
            </span>
          </div>
        </div>

        <!-- Unread Indicator -->
        <div 
          v-if="!article.is_read"
          class="absolute top-2 left-2 w-2 h-2 rounded-full bg-primary-500"
        />
      </div>
    </div>

    <!-- Empty State -->
    <div 
      v-if="imageArticles.length === 0"
      class="flex flex-col items-center justify-center h-64 text-[var(--text-tertiary)]"
    >
      <PhImage :size="48" class="mb-4 opacity-50" />
      <p>{{ t('article.noArticles') }}</p>
    </div>
  </div>
</template>
