<script setup lang="ts">
import { useAppStore } from '@/stores/app'
import { useI18n } from 'vue-i18n'
import { 
  PhArticle, 
  PhImage, 
  PhCheck, 
  PhStar, 
  PhClock,
  PhArrowDown
} from '@phosphor-icons/vue'
import { formatDistanceToNow } from '@/utils/date'

const store = useAppStore()
const { t } = useI18n()

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
  <div class="flex flex-col h-full bg-[var(--bg-primary)]">
    <!-- Header -->
    <div class="p-3 border-b border-[var(--border-color)]">
      <div class="flex items-center justify-between">
        <h2 class="text-lg font-semibold text-[var(--text-primary)]">
          {{ store.currentFeedId 
            ? store.feedMap[store.currentFeedId]?.title 
            : t('filter.' + (store.currentFilter || 'all')) 
          }}
        </h2>
        <button
          @click="store.markAllAsRead(store.currentFeedId || undefined)"
          class="text-sm text-primary-500 hover:text-primary-600"
        >
          {{ t('article.markAllAsRead') }}
        </button>
      </div>
    </div>

    <!-- Article List -->
    <div 
      class="flex-1 overflow-y-auto"
      @scroll="(e) => {
        const target = e.target as HTMLElement
        if (target.scrollHeight - target.scrollTop === target.clientHeight) {
          store.loadMore()
        }
      }"
    >
      <div
        v-for="article in store.filteredArticles"
        :key="article.id"
        @click="selectArticle(article.id)"
        :class="[
          'p-3 border-b border-[var(--border-color)] cursor-pointer transition-colors',
          store.currentArticleId === article.id
            ? 'bg-primary-50 dark:bg-primary-900/20'
            : 'hover:bg-[var(--bg-secondary)]',
          article.is_read ? 'opacity-60' : ''
        ]"
      >
        <div class="flex items-start gap-3">
          <!-- Thumbnail -->
          <div v-if="article.image_url" class="flex-shrink-0">
            <img 
              :src="article.image_url" 
              class="w-16 h-16 object-cover rounded"
              @error="($event.target as HTMLImageElement).style.display = 'none'"
            />
          </div>

          <!-- Content -->
          <div class="flex-1 min-w-0">
            <div class="flex items-center gap-2 mb-1">
              <span 
                v-if="!article.is_read"
                class="w-2 h-2 rounded-full bg-primary-500 flex-shrink-0"
              />
              <h3 class="text-sm font-medium text-[var(--text-primary)] line-clamp-2">
                {{ article.title }}
              </h3>
            </div>

            <div class="flex items-center gap-2 text-xs text-[var(--text-tertiary)]">
              <span v-if="article.feed_title">{{ article.feed_title }}</span>
              <span v-if="article.author">{{ article.author }}</span>
              <span>{{ formatDate(article.published_at) }}</span>
            </div>

            <!-- Summary -->
            <p v-if="article.summary" class="mt-1 text-xs text-[var(--text-secondary)] line-clamp-2">
              {{ article.summary }}
            </p>

            <!-- Actions -->
            <div class="flex items-center gap-2 mt-2">
              <button
                @click.stop="store.toggleFavorite(article.id)"
                :class="[
                  'p-1 rounded transition-colors',
                  article.is_favorite 
                    ? 'text-yellow-500' 
                    : 'text-[var(--text-tertiary)] hover:text-yellow-500'
                ]"
              >
                <PhStar :size="14" :weight="article.is_favorite ? 'fill' : 'regular'" />
              </button>
              <button
                @click.stop="store.toggleReadLater(article.id)"
                :class="[
                  'p-1 rounded transition-colors',
                  article.is_read_later 
                    ? 'text-blue-500' 
                    : 'text-[var(--text-tertiary)] hover:text-blue-500'
                ]"
              >
                <PhClock :size="14" :weight="article.is_read_later ? 'fill' : 'regular'" />
              </button>
              <button
                @click.stop="store.toggleHide(article.id)"
                :class="[
                  'p-1 rounded transition-colors',
                  article.is_hidden 
                    ? 'text-red-500' 
                    : 'text-[var(--text-tertiary)] hover:text-red-500'
                ]"
              >
                <PhCheck :size="14" />
              </button>
            </div>
          </div>
        </div>
      </div>

      <!-- Loading -->
      <div v-if="store.isLoading" class="p-4 text-center text-[var(--text-tertiary)]">
        {{ t('common.loading') }}
      </div>

      <!-- Load More -->
      <div 
        v-if="store.hasMore && !store.isLoading" 
        class="p-4 text-center"
      >
        <button 
          @click="store.loadMore()"
          class="text-sm text-primary-500 hover:text-primary-600"
        >
          {{ t('article.loadingMore') }}
        </button>
      </div>

      <!-- Empty -->
      <div 
        v-if="!store.isLoading && store.filteredArticles.length === 0"
        class="p-8 text-center text-[var(--text-tertiary)]"
      >
        <PhArticle :size="48" class="mx-auto mb-4 opacity-50" />
        <p>{{ t('article.noArticles') }}</p>
      </div>
    </div>
  </div>
</template>
