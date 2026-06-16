<script setup lang="ts">
import { ref } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore } from '@/stores/app'
import { PhX, PhMagnifyingGlass, PhPlus } from '@phosphor-icons/vue'

const { t } = useI18n()
const store = useAppStore()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const url = ref('')
const category = ref('')
const isLoading = ref(false)
const error = ref('')
const discoveredFeeds = ref<string[]>([])

async function addFeed() {
  if (!url.value) {
    error.value = 'URL is required'
    return
  }

  isLoading.value = true
  error.value = ''

  try {
    await store.addFeed(url.value, category.value || undefined)
    emit('close')
  } catch (err: any) {
    error.value = err.message || 'Failed to add feed'
  } finally {
    isLoading.value = false
  }
}

async function discoverFeeds() {
  if (!url.value) {
    error.value = 'URL is required'
    return
  }

  isLoading.value = true
  error.value = ''

  try {
    // TODO: Implement feed discovery
    discoveredFeeds.value = []
  } catch (err: any) {
    error.value = err.message || 'Failed to discover feeds'
  } finally {
    isLoading.value = false
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-[500px]">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('sidebar.addFeed') }}
        </h2>
        <button 
          @click="emit('close')"
          class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
        >
          <PhX :size="20" />
        </button>
      </div>

      <!-- Content -->
      <div class="p-4 space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('feed.url') }}</label>
          <div class="flex gap-2">
            <input 
              v-model="url" 
              class="input flex-1" 
              placeholder="https://example.com/feed.xml"
              @keyup.enter="addFeed"
            />
            <button 
              @click="discoverFeeds"
              :disabled="isLoading"
              class="btn btn-secondary"
            >
              <PhMagnifyingGlass :size="16" />
            </button>
          </div>
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">{{ t('feed.category') }}</label>
          <input 
            v-model="category" 
            class="input" 
            placeholder="Optional category"
          />
        </div>

        <!-- Discovered Feeds -->
        <div v-if="discoveredFeeds.length > 0">
          <label class="block text-sm font-medium mb-2">Discovered Feeds</label>
          <div class="space-y-2">
            <button
              v-for="feed in discoveredFeeds"
              :key="feed"
              @click="url = feed"
              class="flex items-center w-full p-2 text-sm text-left rounded-lg hover:bg-gray-100 dark:hover:bg-gray-700"
            >
              <PhPlus :size="14" class="mr-2 text-primary-500" />
              {{ feed }}
            </button>
          </div>
        </div>

        <!-- Error -->
        <div v-if="error" class="text-sm text-red-500">
          {{ error }}
        </div>
      </div>

      <!-- Footer -->
      <div class="flex justify-end gap-2 p-4 border-t border-gray-200 dark:border-gray-700">
        <button @click="emit('close')" class="btn btn-secondary">
          {{ t('common.cancel') }}
        </button>
        <button 
          @click="addFeed" 
          :disabled="isLoading || !url"
          class="btn btn-primary"
        >
          {{ isLoading ? t('common.loading') : t('common.add') }}
        </button>
      </div>
    </div>
  </div>
</template>
