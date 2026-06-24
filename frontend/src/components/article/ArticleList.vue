<script setup lang="ts">
import { ref, computed, onMounted, onUnmounted } from 'vue'
import { useAppStore } from '@/stores/app'
import { useI18n } from 'vue-i18n'
import { 
  PhArticle, 
  PhImage, 
  PhCheck, 
  PhStar, 
  PhClock,
  PhArrowDown,
  PhList,
  PhGridFour,
  PhRows
} from '@phosphor-icons/vue'
import { formatDistanceToNow } from '@/utils/date'

const store = useAppStore()
const { t } = useI18n()
const scrollContainer = ref<HTMLElement | null>(null)
const selectedArticles = ref<number[]>([])
const isSelectionMode = ref(false)

function selectArticle(id: number) {
  if (isSelectionMode.value) {
    toggleSelection(id)
  } else {
    store.selectArticle(id)
  }
}

function toggleSelection(id: number) {
  const index = selectedArticles.value.indexOf(id)
  if (index === -1) {
    selectedArticles.value.push(id)
  } else {
    selectedArticles.value.splice(index, 1)
  }
}

function toggleSelectionMode() {
  isSelectionMode.value = !isSelectionMode.value
  if (!isSelectionMode.value) {
    selectedArticles.value = []
  }
}

function selectAll() {
  selectedArticles.value = store.filteredArticles.map(a => a.id)
}

function formatDate(dateStr: string): string {
  try {
    return formatDistanceToNow(new Date(dateStr))
  } catch {
    return ''
  }
}

function getDateGroup(dateStr: string): string {
  try {
    const date = new Date(dateStr)
    const now = new Date()
    const today = new Date(now.getFullYear(), now.getMonth(), now.getDate())
    const yesterday = new Date(today)
    yesterday.setDate(yesterday.getDate() - 1)
    const articleDate = new Date(date.getFullYear(), date.getMonth(), date.getDate())
    
    if (articleDate.getTime() === today.getTime()) {
      return t('date.today')
    } else if (articleDate.getTime() === yesterday.getTime()) {
      return t('date.yesterday')
    } else {
      return date.toLocaleDateString(undefined, { month: 'long', day: 'numeric' })
    }
  } catch {
    return ''
  }
}

const groupedArticles = computed(() => {
  const groups: { label: string; articles: typeof store.filteredArticles }[] = []
  let currentGroup = ''
  
  store.filteredArticles.forEach(article => {
    const group = getDateGroup(article.published_at)
    if (group !== currentGroup) {
      currentGroup = group
      groups.push({ label: group, articles: [] })
    }
    groups[groups.length - 1].articles.push(article)
  })
  
  return groups
})

const viewModes = [
  { key: 'list' as const, icon: PhList, label: 'view.list' },
  { key: 'card' as const, icon: PhGridFour, label: 'view.card' },
  { key: 'compact' as const, icon: PhRows, label: 'view.compact' }
]

async function batchMarkRead() {
  if (selectedArticles.value.length === 0) return
  await store.batchMarkAsRead(selectedArticles.value)
  selectedArticles.value = []
  isSelectionMode.value = false
}

async function batchFavorite() {
  if (selectedArticles.value.length === 0) return
  await store.batchToggleFavorite(selectedArticles.value)
  selectedArticles.value = []
  isSelectionMode.value = false
}

async function batchHide() {
  if (selectedArticles.value.length === 0) return
  await store.batchToggleHide(selectedArticles.value)
  selectedArticles.value = []
  isSelectionMode.value = false
}

// Auto load more on scroll
function handleScroll(e: Event) {
  const target = e.target as HTMLElement
  const threshold = 200
  if (target.scrollHeight - target.scrollTop - target.clientHeight < threshold) {
    store.loadMore()
  }
}

onMounted(() => {
  scrollContainer.value?.addEventListener('scroll', handleScroll)
})

onUnmounted(() => {
  scrollContainer.value?.removeEventListener('scroll', handleScroll)
})
</script>

<template>
  <div class="flex flex-col h-full bg-[var(--bg-primary)]">
    <!-- Header -->
    <div class="p-3 border-b border-[var(--border-color)]">
      <div class="flex items-center justify-between mb-2">
        <h2 class="text-lg font-semibold text-[var(--text-primary)]">
          {{ store.currentFeedId 
            ? store.feedMap[store.currentFeedId]?.title 
            : t('filter.' + (store.currentFilter || 'all')) 
          }}
        </h2>
        <div class="flex items-center gap-1">
          <!-- View Mode Toggle -->
          <div class="flex items-center bg-[var(--bg-secondary)] rounded-lg p-0.5">
            <button
              v-for="mode in viewModes"
              :key="mode.key"
              @click="store.setViewMode(mode.key)"
              :class="[
                'p-1.5 rounded-md transition-colors',
                store.viewMode === mode.key
                  ? 'bg-[var(--bg-primary)] text-[var(--text-primary)] shadow-sm'
                  : 'text-[var(--text-tertiary)] hover:text-[var(--text-secondary)]'
              ]"
              :title="t(mode.label)"
            >
              <component :is="mode.icon" :size="14" />
            </button>
          </div>
        </div>
      </div>
      
      <!-- Batch Actions -->
      <div class="flex items-center justify-between">
        <div class="flex items-center gap-2">
          <button
            @click="toggleSelectionMode"
            :class="[
              'text-xs px-2 py-1 rounded transition-colors',
              isSelectionMode 
                ? 'bg-primary-500 text-white' 
                : 'text-[var(--text-tertiary)] hover:text-[var(--text-secondary)]'
            ]"
          >
            {{ isSelectionMode ? t('action.cancel') : t('action.select') }}
          </button>
          <template v-if="isSelectionMode">
            <button
              @click="selectAll"
              class="text-xs text-[var(--text-tertiary)] hover:text-[var(--text-secondary)]"
            >
              {{ t('action.selectAll') }}
            </button>
            <span class="text-xs text-[var(--text-tertiary)]">
              {{ selectedArticles.length }} {{ t('article.selected') }}
            </span>
          </template>
        </div>
        
        <template v-if="isSelectionMode && selectedArticles.length > 0">
          <div class="flex items-center gap-1">
            <button
              @click="batchMarkRead"
              class="text-xs px-2 py-1 rounded bg-[var(--bg-secondary)] hover:bg-[var(--bg-tertiary)] text-[var(--text-secondary)]"
            >
              {{ t('action.markRead') }}
            </button>
            <button
              @click="batchFavorite"
              class="text-xs px-2 py-1 rounded bg-[var(--bg-secondary)] hover:bg-[var(--bg-tertiary)] text-[var(--text-secondary)]"
            >
              {{ t('action.favorite') }}
            </button>
            <button
              @click="batchHide"
              class="text-xs px-2 py-1 rounded bg-[var(--bg-secondary)] hover:bg-[var(--bg-tertiary)] text-[var(--text-secondary)]"
            >
              {{ t('action.hide') }}
            </button>
          </div>
        </template>
        
        <button
          v-else
          @click="store.markAllAsRead(store.currentFeedId || undefined)"
          class="text-sm text-primary-500 hover:text-primary-600"
        >
          {{ t('article.markAllAsRead') }}
        </button>
      </div>
    </div>

    <!-- Article List -->
    <div ref="scrollContainer" class="flex-1 overflow-y-auto">
      <!-- List View -->
      <template v-if="store.viewMode === 'list'">
        <div v-for="group in groupedArticles" :key="group.label">
          <!-- Date Group Header -->
          <div class="sticky top-0 z-10 px-3 py-2 bg-[var(--bg-secondary)] border-b border-[var(--border-color)]">
            <span class="text-xs font-semibold text-[var(--text-tertiary)] uppercase tracking-wider">
              {{ group.label }}
            </span>
          </div>
          
          <div
            v-for="article in group.articles"
            :key="article.id"
            @click="selectArticle(article.id)"
            :class="[
              'p-3 border-b border-[var(--border-color)] cursor-pointer transition-colors',
              store.currentArticleId === article.id
                ? 'bg-primary-50 dark:bg-primary-900/20'
                : 'hover:bg-[var(--bg-secondary)]',
              article.is_read ? 'opacity-60' : '',
              selectedArticles.includes(article.id) ? 'bg-primary-50 dark:bg-primary-900/10' : ''
            ]"
          >
            <div class="flex items-start gap-3">
              <!-- Selection Checkbox -->
              <div v-if="isSelectionMode" class="flex-shrink-0 pt-1">
                <input
                  type="checkbox"
                  :checked="selectedArticles.includes(article.id)"
                  @click.stop="toggleSelection(article.id)"
                  class="w-4 h-4 rounded border-gray-300 text-primary-500 focus:ring-primary-500"
                />
              </div>
              
              <!-- Thumbnail -->
              <div v-if="article.image_url" class="flex-shrink-0">
                <img 
                  v-lazyload="article.image_url"
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
        </div>
      </template>

      <!-- Card View -->
      <template v-else-if="store.viewMode === 'card'">
        <div class="grid grid-cols-2 gap-3 p-3">
          <div
            v-for="article in store.filteredArticles"
            :key="article.id"
            @click="selectArticle(article.id)"
            :class="[
              'border border-[var(--border-color)] rounded-lg overflow-hidden cursor-pointer transition-all hover:shadow-md',
              store.currentArticleId === article.id
                ? 'ring-2 ring-primary-500'
                : '',
              article.is_read ? 'opacity-60' : ''
            ]"
          >
            <!-- Card Image -->
            <div v-if="article.image_url" class="aspect-video bg-[var(--bg-secondary)]">
              <img 
                v-lazyload="article.image_url"
                class="w-full h-full object-cover"
                @error="($event.target as HTMLImageElement).style.display = 'none'"
              />
            </div>
            
            <!-- Card Content -->
            <div class="p-3">
              <div class="flex items-center gap-1 mb-2">
                <span 
                  v-if="!article.is_read"
                  class="w-2 h-2 rounded-full bg-primary-500 flex-shrink-0"
                />
                <span class="text-xs text-[var(--text-tertiary)] truncate">
                  {{ article.feed_title }}
                </span>
              </div>
              
              <h3 class="text-sm font-medium text-[var(--text-primary)] line-clamp-3 mb-2">
                {{ article.title }}
              </h3>
              
              <div class="flex items-center justify-between">
                <span class="text-xs text-[var(--text-tertiary)]">
                  {{ formatDate(article.published_at) }}
                </span>
                <div class="flex items-center gap-1">
                  <button
                    @click.stop="store.toggleFavorite(article.id)"
                    :class="[
                      'p-1 rounded transition-colors',
                      article.is_favorite 
                        ? 'text-yellow-500' 
                        : 'text-[var(--text-tertiary)]'
                    ]"
                  >
                    <PhStar :size="12" :weight="article.is_favorite ? 'fill' : 'regular'" />
                  </button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </template>

      <!-- Compact View -->
      <template v-else>
        <div
          v-for="article in store.filteredArticles"
          :key="article.id"
          @click="selectArticle(article.id)"
          :class="[
            'px-3 py-2 border-b border-[var(--border-color)] cursor-pointer transition-colors group',
            store.currentArticleId === article.id
              ? 'bg-primary-50 dark:bg-primary-900/20'
              : 'hover:bg-[var(--bg-secondary)]',
            article.is_read ? 'opacity-60' : ''
          ]"
        >
          <div class="flex items-center gap-2">
            <span 
              v-if="!article.is_read"
              class="w-2 h-2 rounded-full bg-primary-500 flex-shrink-0"
            />
            <span class="text-xs text-[var(--text-tertiary)] flex-shrink-0 w-20 truncate">
              {{ article.feed_title }}
            </span>
            <h3 class="text-sm text-[var(--text-primary)] flex-1 truncate">
              {{ article.title }}
            </h3>
            <button
              @click.stop="store.toggleFavorite(article.id)"
              :class="[
                'p-1 rounded transition-colors flex-shrink-0',
                article.is_favorite 
                  ? 'text-yellow-500' 
                  : 'text-[var(--text-tertiary)] opacity-0 group-hover:opacity-100'
              ]"
            >
              <PhStar :size="12" :weight="article.is_favorite ? 'fill' : 'regular'" />
            </button>
            <span class="text-xs text-[var(--text-tertiary)] flex-shrink-0">
              {{ formatDate(article.published_at) }}
            </span>
          </div>
        </div>
      </template>

      <!-- Loading -->
      <div v-if="store.isLoading" class="p-4 text-center text-[var(--text-tertiary)]">
        <div class="inline-flex items-center gap-2">
          <div class="w-4 h-4 border-2 border-primary-500 border-t-transparent rounded-full animate-spin"></div>
          {{ t('common.loading') }}
        </div>
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
