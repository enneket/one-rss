<script setup lang="ts">
import { ref } from 'vue'
import { useAppStore, type Filter, type Feed } from '@/stores/app'
import { useI18n } from 'vue-i18n'
import { 
  PhHouse, 
  PhEnvelopeSimple, 
  PhStar, 
  PhClock, 
  PhGear, 
  PhPlus,
  PhMagnifyingGlass,
  PhRss,
  PhPencilSimple,
  PhArrowsClockwise,
  PhExport,
  PhDownloadSimple
} from '@phosphor-icons/vue'
import AppLogo from '@/components/AppLogo.vue'
import AddFeedModal from '@/components/modals/AddFeedModal.vue'
import EditFeedModal from '@/components/modals/EditFeedModal.vue'
import SettingsModal from '@/components/modals/SettingsModal.vue'

const store = useAppStore()
const { t } = useI18n()

const showAddFeed = ref(false)
const showSettings = ref(false)
const editingFeed = ref<Feed | null>(null)
const isRefreshing = ref(false)
const importInput = ref<HTMLInputElement | null>(null)

async function handleRefresh() {
  isRefreshing.value = true
  await store.refreshAll()
  setTimeout(() => { isRefreshing.value = false }, 500)
}

async function handleExport() {
  try {
    await store.exportFeeds()
  } catch (error) {
    console.error('Export failed:', error)
  }
}

function handleImportClick() {
  importInput.value?.click()
}

async function handleImport(event: Event) {
  const input = event.target as HTMLInputElement
  const file = input.files?.[0]
  if (!file) return
  
  try {
    const result = await store.importFeeds(file)
    alert(t('sidebar.importSuccess', { imported: result.imported, skipped: result.skipped }))
  } catch (error) {
    console.error('Import failed:', error)
    alert(t('sidebar.importFailed'))
  } finally {
    input.value = ''
  }
}

function editFeed(feed: Feed, e: MouseEvent) {
  e.stopPropagation()
  editingFeed.value = feed
}

const filters: { key: Filter; icon: any; label: string }[] = [
  { key: 'all', icon: PhHouse, label: 'sidebar.all' },
  { key: 'unread', icon: PhEnvelopeSimple, label: 'sidebar.unread' },
  { key: 'favorites', icon: PhStar, label: 'sidebar.favorites' },
  { key: 'readLater', icon: PhClock, label: 'sidebar.readLater' }
]

function setFilter(filter: Filter) {
  store.setFilter(filter)
}

function setFeed(feedId: number) {
  store.setFeed(feedId)
}

function getUnreadCount(feedId: number): number {
  return store.unreadCounts.feedCounts[feedId] || 0
}
</script>

<template>
  <div class="flex flex-col h-full bg-[var(--bg-secondary)]">
    <!-- Header -->
    <div class="p-4 border-b border-[var(--border-color)]">
      <div class="flex items-center justify-between mb-4">
        <h1 class="text-xl font-bold text-[var(--text-primary)]">
          {{ t('common.app_name') }}
        </h1>
        <div class="flex items-center gap-1">
          <button 
            @click="handleRefresh"
            class="p-2 rounded-lg hover:bg-[var(--bg-tertiary)] transition-colors"
            :class="{ 'animate-spin': isRefreshing }"
            :title="t('sidebar.refreshAll')"
          >
            <PhArrowsClockwise :size="18" />
          </button>
          <button 
            @click="showAddFeed = true"
            class="p-2 rounded-lg hover:bg-[var(--bg-tertiary)] transition-colors"
            :title="t('sidebar.addFeed')"
          >
            <PhPlus :size="20" />
          </button>
        </div>
      </div>
      
      <!-- Search -->
      <div class="relative flex items-center">
        <PhMagnifyingGlass 
          :size="16" 
          class="absolute left-3 text-[var(--text-tertiary)] pointer-events-none"
        />
        <input
          v-model="store.searchQuery"
          type="text"
          :placeholder="t('common.search')"
          class="w-full pl-10 pr-3 py-1.5 border border-gray-300 dark:border-gray-600 rounded-lg bg-white dark:bg-gray-800 text-sm"
        />
      </div>
    </div>

    <!-- Filters -->
    <div class="p-2 border-b border-[var(--border-color)]">
      <button
        v-for="filter in filters"
        :key="filter.key"
        @click="setFilter(filter.key)"
        :class="[
          'flex items-center w-full px-3 py-2 rounded-lg transition-colors text-left',
          store.currentFilter === filter.key
            ? 'bg-primary-500 text-white'
            : 'hover:bg-[var(--bg-tertiary)] text-[var(--text-primary)]'
        ]"
      >
        <component :is="filter.icon" :size="18" class="mr-3" />
        <span>{{ t(filter.label) }}</span>
        <span 
          v-if="filter.key === 'unread' && store.unreadCounts.total > 0"
          class="ml-auto px-2 py-0.5 text-xs rounded-full bg-red-500 text-white"
        >
          {{ store.unreadCounts.total }}
        </span>
      </button>
    </div>

    <!-- Feed List -->
    <div class="flex-1 overflow-y-auto p-2">
      <div class="mb-2 px-3 text-xs font-semibold text-[var(--text-tertiary)] uppercase tracking-wider">
        {{ t('sidebar.categories') }}
      </div>
      
      <div v-for="feed in store.feeds" :key="feed.id" class="group relative">
        <button
          @click="setFeed(feed.id)"
          :class="[
            'flex items-center w-full px-3 py-2 rounded-lg transition-colors text-sm text-left',
            store.currentFeedId === feed.id
              ? 'bg-primary-100 text-primary-700 dark:bg-primary-900 dark:text-primary-300'
              : 'hover:bg-[var(--bg-tertiary)] text-[var(--text-primary)]'
          ]"
        >
          <PhRss v-if="!feed.image_url" :size="18" class="mr-3 flex-shrink-0 text-[var(--text-tertiary)]" />
          <img 
            v-else
            :src="feed.image_url" 
            class="w-[18px] h-[18px] rounded mr-3 flex-shrink-0 object-cover"
            @error="($event.target as HTMLImageElement).style.display = 'none'"
          />
          
          <span class="flex-1 min-w-0 truncate">{{ feed.title }}</span>
          
          <span 
            v-if="getUnreadCount(feed.id) > 0"
            class="ml-auto px-1.5 py-0.5 text-xs leading-none rounded-full bg-red-500 text-white"
          >
            {{ getUnreadCount(feed.id) }}
          </span>
        </button>
        <button
          @click="editFeed(feed, $event)"
          class="absolute right-2 top-1/2 -translate-y-1/2 p-1 rounded opacity-0 group-hover:opacity-100 hover:bg-[var(--bg-tertiary)] transition-opacity text-[var(--text-tertiary)] hover:text-[var(--text-primary)]"
          :title="t('common.edit')"
        >
          <PhPencilSimple :size="14" />
        </button>
      </div>

      <div v-if="store.feeds.length === 0" class="px-3 py-4 text-sm text-[var(--text-tertiary)]">
        {{ t('common.noData') }}
      </div>
    </div>

    <!-- Footer -->
    <div class="p-2 border-t border-[var(--border-color)] space-y-1">
      <div class="flex gap-1">
        <button 
          @click="handleExport"
          class="flex items-center flex-1 px-3 py-2 rounded-lg hover:bg-[var(--bg-tertiary)] transition-colors text-[var(--text-primary)] text-sm"
          :title="t('sidebar.exportFeeds')"
        >
          <PhExport :size="16" class="mr-2" />
          <span>{{ t('sidebar.exportFeeds') }}</span>
        </button>
        <button 
          @click="handleImportClick"
          class="flex items-center flex-1 px-3 py-2 rounded-lg hover:bg-[var(--bg-tertiary)] transition-colors text-[var(--text-primary)] text-sm"
          :title="t('sidebar.importFeeds')"
        >
          <PhDownloadSimple :size="16" class="mr-2" />
          <span>{{ t('sidebar.importFeeds') }}</span>
        </button>
      </div>
      <button 
        @click="showSettings = true"
        class="flex items-center w-full px-3 py-2 rounded-lg hover:bg-[var(--bg-tertiary)] transition-colors text-[var(--text-primary)]"
      >
        <PhGear :size="18" class="mr-3" />
        <span>{{ t('common.settings') }}</span>
      </button>
    </div>
    
    <!-- Hidden file input for import -->
    <input 
      ref="importInput"
      type="file"
      accept=".xml,.opml"
      class="hidden"
      @change="handleImport"
    />

    <!-- Add Feed Modal -->
    <AddFeedModal v-if="showAddFeed" @close="showAddFeed = false" />
    
    <!-- Edit Feed Modal -->
    <EditFeedModal v-if="editingFeed" :feed="editingFeed" @close="editingFeed = null" />
    
    <!-- Settings Modal -->
    <SettingsModal v-if="showSettings" @close="showSettings = false" />
  </div>
</template>
