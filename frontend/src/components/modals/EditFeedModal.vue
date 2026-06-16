<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { useAppStore, type Feed } from '@/stores/app'
import { PhX } from '@phosphor-icons/vue'
import axios from 'axios'

const { t } = useI18n()
const store = useAppStore()

const props = defineProps<{
  feed: Feed
}>()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const form = ref({
  title: '',
  url: '',
  category: '',
  hide_from_timeline: false,
  proxy_enabled: false,
  proxy_url: '',
  refresh_interval: 0,
  is_image_mode: false,
  article_view_mode: 'global',
  auto_expand_content: 'global'
})

onMounted(() => {
  form.value = {
    title: props.feed.title,
    url: props.feed.url,
    category: props.feed.category,
    hide_from_timeline: props.feed.hide_from_timeline,
    proxy_enabled: props.feed.proxy_enabled,
    proxy_url: props.feed.proxy_url,
    refresh_interval: props.feed.refresh_interval,
    is_image_mode: props.feed.is_image_mode,
    article_view_mode: props.feed.article_view_mode,
    auto_expand_content: props.feed.auto_expand_content
  }
})

async function saveFeed() {
  try {
    await axios.post('/api/feeds/update', {
      id: props.feed.id,
      ...form.value
    })
    await store.fetchFeeds()
    emit('close')
  } catch (error) {
    console.error('Failed to update feed:', error)
  }
}

async function deleteFeed() {
  if (confirm(t('feed.deleteConfirm'))) {
    try {
      await store.deleteFeed(props.feed.id)
      emit('close')
    } catch (error) {
      console.error('Failed to delete feed:', error)
    }
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-[500px] max-h-[80vh] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('common.edit') }} - {{ feed.title }}
        </h2>
        <button 
          @click="emit('close')"
          class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
        >
          <PhX :size="20" />
        </button>
      </div>

      <!-- Content -->
      <div class="flex-1 overflow-y-auto p-4 space-y-4">
        <div>
          <label class="block text-sm font-medium mb-1">{{ t('feed.title') }}</label>
          <input v-model="form.title" class="input" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">{{ t('feed.url') }}</label>
          <input v-model="form.url" class="input" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">{{ t('feed.category') }}</label>
          <input v-model="form.category" class="input" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">{{ t('feed.refreshInterval') }}</label>
          <input v-model="form.refresh_interval" type="number" class="input" :placeholder="t('feed.refreshIntervalPlaceholder')" />
        </div>

        <div class="space-y-2">
          <label class="flex items-center gap-2">
            <input v-model="form.hide_from_timeline" type="checkbox" class="rounded" />
            <span class="text-sm">{{ t('feed.hideFromTimeline') }}</span>
          </label>
          <label class="flex items-center gap-2">
            <input v-model="form.is_image_mode" type="checkbox" class="rounded" />
            <span class="text-sm">{{ t('feed.imageMode') }}</span>
          </label>
          <label class="flex items-center gap-2">
            <input v-model="form.proxy_enabled" type="checkbox" class="rounded" />
            <span class="text-sm">{{ t('feed.proxy') }}</span>
          </label>
        </div>

        <div v-if="form.proxy_enabled">
          <label class="block text-sm font-medium mb-1">{{ t('feed.proxyUrl') }}</label>
          <input v-model="form.proxy_url" class="input" placeholder="socks5://localhost:1080" />
        </div>

        <div>
          <label class="block text-sm font-medium mb-1">{{ t('feed.articleViewMode') }}</label>
          <select v-model="form.article_view_mode" class="input">
            <option value="global">{{ t('viewMode.global') }}</option>
            <option value="normal">{{ t('viewMode.normal') }}</option>
            <option value="compact">{{ t('viewMode.compact') }}</option>
            <option value="card">{{ t('viewMode.card') }}</option>
            <option value="gallery">{{ t('viewMode.gallery') }}</option>
          </select>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex justify-between p-4 border-t border-gray-200 dark:border-gray-700">
        <button @click="deleteFeed" class="btn btn-danger">
          {{ t('common.delete') }}
        </button>
        <div class="flex gap-2">
          <button @click="emit('close')" class="btn btn-secondary">
            {{ t('common.cancel') }}
          </button>
          <button @click="saveFeed" class="btn btn-primary">
            {{ t('common.save') }}
          </button>
        </div>
      </div>
    </div>
  </div>
</template>
