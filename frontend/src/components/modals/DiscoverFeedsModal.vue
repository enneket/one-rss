<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { PhX, PhMagnifyingGlass, PhPlus, PhSpinner } from '@phosphor-icons/vue'
import axios from 'axios'

const { t } = useI18n()
const store = useAppStore()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const mode = ref<'single' | 'batch'>('single')
const url = ref('')
const isLoading = ref(false)
const discoveredFeeds = ref<{ url: string; title: string }[]>([])

async function discoverFeeds() {
  if (!url.value) return

  isLoading.value = true
  discoveredFeeds.value = []

  try {
    const response = await axios.post('/api/feeds/discover', { url: url.value })
    discoveredFeeds.value = response.data.feeds || []
  } catch (error) {
    console.error('Failed to discover feeds:', error)
  } finally {
    isLoading.value = false
  }
}

async function addFeed(feedUrl: string) {
  try {
    await store.addFeed(feedUrl)
    discoveredFeeds.value = discoveredFeeds.value.filter(f => f.url !== feedUrl)
  } catch (error) {
    console.error('Failed to add feed:', error)
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-[600px] max-h-[80vh] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('sidebar.discover') }}
        </h2>
        <button 
          @click="emit('close')"
          class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
        >
          <PhX :size="20" />
        </button>
      </div>

      <!-- Mode Tabs -->
      <div class="flex border-b border-gray-200 dark:border-gray-700">
        <button
          @click="mode = 'single'"
          :class="[
            'flex-1 py-2 text-sm font-medium transition-colors',
            mode === 'single'
              ? 'text-primary-600 border-b-2 border-primary-600'
              : 'text-gray-500 hover:text-gray-700'
          ]"
        >
          Single Discovery
        </button>
        <button
          @click="mode = 'batch'"
          :class="[
            'flex-1 py-2 text-sm font-medium transition-colors',
            mode === 'batch'
              ? 'text-primary-600 border-b-2 border-primary-600'
              : 'text-gray-500 hover:text-gray-700'
          ]"
        >
          Batch Discovery
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-4">
        <!-- Single Mode -->
        <div v-if="mode === 'single'" class="space-y-4">
          <div class="flex gap-2">
            <input 
              v-model="url" 
              class="input flex-1" 
              placeholder="https://example.com"
              @keyup.enter="discoverFeeds"
            />
            <button 
              @click="discoverFeeds"
              :disabled="isLoading || !url"
              class="btn btn-primary"
            >
              <PhSpinner v-if="isLoading" :size="16" class="animate-spin" />
              <PhMagnifyingGlass v-else :size="16" />
            </button>
          </div>

          <!-- Discovered Feeds -->
          <div v-if="discoveredFeeds.length > 0" class="space-y-2">
            <h3 class="text-sm font-medium text-gray-700 dark:text-gray-300">
              Discovered Feeds
            </h3>
            <div 
              v-for="feed in discoveredFeeds" 
              :key="feed.url"
              class="flex items-center justify-between p-3 bg-gray-50 dark:bg-gray-700 rounded-lg"
            >
              <div>
                <p class="text-sm font-medium">{{ feed.title }}</p>
                <p class="text-xs text-gray-500">{{ feed.url }}</p>
              </div>
              <button 
                @click="addFeed(feed.url)"
                class="p-2 rounded-lg text-primary-500 hover:bg-primary-50 dark:hover:bg-primary-900/20"
              >
                <PhPlus :size="16" />
              </button>
            </div>
          </div>

          <!-- Empty State -->
          <div 
            v-else-if="!isLoading"
            class="text-center py-8 text-gray-500"
          >
            <PhMagnifyingGlass :size="48" class="mx-auto mb-4 opacity-50" />
            <p>Enter a website URL to discover RSS feeds</p>
          </div>
        </div>

        <!-- Batch Mode -->
        <div v-if="mode === 'batch'" class="space-y-4">
          <p class="text-sm text-gray-500">
            Discover feeds from your existing subscriptions' websites.
          </p>
          <button 
            @click="discoverFeeds"
            :disabled="isLoading"
            class="btn btn-primary w-full"
          >
            {{ isLoading ? 'Discovering...' : 'Start Batch Discovery' }}
          </button>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex justify-end p-4 border-t border-gray-200 dark:border-gray-700">
        <button @click="emit('close')" class="btn btn-secondary">
          {{ t('common.close') }}
        </button>
      </div>
    </div>
  </div>
</template>
