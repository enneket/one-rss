<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { useI18n } from 'vue-i18n'
import { 
  PhX, 
  PhGear, 
  PhPaintBrush, 
  PhTranslate, 
  PhRobot, 
  PhArticle,
  PhNetwork,
  PhHardDrives,
  PhPlug,
  PhListChecks,
  PhKeyboard,
  PhTag,
  PhChartLine,
  PhInfo
} from '@phosphor-icons/vue'
import axios from 'axios'

const { t, locale } = useI18n()

const emit = defineEmits<{
  (e: 'close'): void
}>()

const activeTab = ref('general')
const settings = ref<Record<string, any>>({})

const tabs = [
  { id: 'general', icon: PhGear, label: 'settings.general' },
  { id: 'reading', icon: PhArticle, label: 'settings.reading' },
  { id: 'translation', icon: PhTranslate, label: 'settings.translation' },
  { id: 'ai', icon: PhRobot, label: 'settings.ai' },
  { id: 'summary', icon: PhArticle, label: 'settings.summary' },
  { id: 'network', icon: PhNetwork, label: 'settings.network' },
  { id: 'storage', icon: PhHardDrives, label: 'settings.storage' },
  { id: 'integrations', icon: PhPlug, label: 'settings.integrations' },
  { id: 'rules', icon: PhListChecks, label: 'settings.rules' },
  { id: 'shortcuts', icon: PhKeyboard, label: 'settings.shortcuts' },
  { id: 'tags', icon: PhTag, label: 'settings.tags' },
  { id: 'statistics', icon: PhChartLine, label: 'settings.statistics' },
  { id: 'about', icon: PhInfo, label: 'settings.about' }
]

onMounted(async () => {
  try {
    const response = await axios.get('/api/settings')
    settings.value = response.data
  } catch (error) {
    console.error('Failed to load settings:', error)
  }
})

const saveSuccess = ref(false)

async function saveSettings() {
  try {
    await axios.post('/api/settings', settings.value)
    
    // 更新语言设置
    if (settings.value.language) {
      locale.value = settings.value.language
      localStorage.setItem('language', settings.value.language)
    }
    
    // 更新主题设置
    if (settings.value.theme) {
      const html = document.documentElement
      if (settings.value.theme === 'dark') {
        html.classList.add('dark')
      } else if (settings.value.theme === 'light') {
        html.classList.remove('dark')
      } else {
        // auto 模式
        if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
          html.classList.add('dark')
        } else {
          html.classList.remove('dark')
        }
      }
      localStorage.setItem('theme', settings.value.theme)
    }
    
    saveSuccess.value = true
    setTimeout(() => {
      saveSuccess.value = false
      // 刷新页面以应用所有设置
      window.location.reload()
    }, 1000)
  } catch (error) {
    console.error('Failed to save settings:', error)
  }
}
</script>

<template>
  <div class="fixed inset-0 z-50 flex items-center justify-center bg-black/50">
    <div class="bg-white dark:bg-gray-800 rounded-lg shadow-xl w-[800px] h-[600px] flex flex-col">
      <!-- Header -->
      <div class="flex items-center justify-between p-4 border-b border-gray-200 dark:border-gray-700">
        <h2 class="text-lg font-semibold text-gray-900 dark:text-white">
          {{ t('common.settings') }}
        </h2>
        <button 
          @click="emit('close')"
          class="p-1 rounded hover:bg-gray-100 dark:hover:bg-gray-700"
        >
          <PhX :size="20" />
        </button>
      </div>

      <!-- Content -->
      <div class="flex flex-1 overflow-hidden">
        <!-- Sidebar -->
        <div class="w-48 border-r border-gray-200 dark:border-gray-700 overflow-y-auto">
          <button
            v-for="tab in tabs"
            :key="tab.id"
            @click="activeTab = tab.id"
            :class="[
              'flex items-center w-full px-3 py-2 text-sm transition-colors',
              activeTab === tab.id
                ? 'bg-primary-50 text-primary-600 dark:bg-primary-900/20 dark:text-primary-400'
                : 'hover:bg-gray-100 dark:hover:bg-gray-700 text-gray-700 dark:text-gray-300'
            ]"
          >
            <component :is="tab.icon" :size="16" class="mr-2" />
            {{ t(tab.label) }}
          </button>
        </div>

        <!-- Panel -->
        <div class="flex-1 overflow-y-auto p-4">
          <!-- General -->
          <div v-if="activeTab === 'general'" class="space-y-4">
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.language') }}</label>
              <select v-model="settings.language" class="input">
                <option value="en-US">English</option>
                <option value="zh-CN">中文</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.theme') }}</label>
              <select v-model="settings.theme" class="input">
                <option value="light">{{ t('settings.themeLight') }}</option>
                <option value="dark">{{ t('settings.themeDark') }}</option>
                <option value="auto">{{ t('settings.themeAuto') }}</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.updateInterval') }}</label>
              <input v-model="settings.update_interval" type="number" class="input" placeholder="30" />
            </div>
          </div>

          <!-- Reading -->
          <div v-if="activeTab === 'reading'" class="space-y-4">
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.articleViewMode') }}</label>
              <select v-model="settings.article_view_mode" class="input">
                <option value="normal">{{ t('filter.all') }}</option>
                <option value="compact">{{ t('settings.reading') }}</option>
                <option value="card">Card</option>
                <option value="gallery">{{ t('filter.imageGallery') }}</option>
              </select>
            </div>
            <div>
              <label class="flex items-center gap-2">
                <input v-model="settings.auto_expand_content" type="checkbox" class="rounded" />
                <span class="text-sm">{{ t('settings.autoExpandContent') }}</span>
              </label>
            </div>
            <div>
              <label class="flex items-center gap-2">
                <input v-model="settings.mark_read_on_scroll" type="checkbox" class="rounded" />
                <span class="text-sm">{{ t('settings.markReadOnScroll') }}</span>
              </label>
            </div>
          </div>

          <!-- Translation -->
          <div v-if="activeTab === 'translation'" class="space-y-4">
            <div>
              <label class="flex items-center gap-2">
                <input v-model="settings.translation_enabled" type="checkbox" class="rounded" />
                <span class="text-sm font-medium">{{ t('settings.translationEnabled') }}</span>
              </label>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.translationProvider') }}</label>
              <select v-model="settings.translation_provider" class="input">
                <option value="google">Google Translate</option>
                <option value="deepl">DeepL</option>
                <option value="ai">AI Translation</option>
              </select>
            </div>
            <div v-if="settings.translation_provider === 'deepl'">
              <label class="block text-sm font-medium mb-1">{{ t('settings.deepLApiKey') }}</label>
              <input v-model="settings.deepl_api_key" type="password" class="input" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.targetLanguage') }}</label>
              <select v-model="settings.translation_target_lang" class="input">
                <option value="en">English</option>
                <option value="zh">中文</option>
                <option value="ja">日本語</option>
                <option value="ko">한국어</option>
              </select>
            </div>
          </div>

          <!-- AI -->
          <div v-if="activeTab === 'ai'" class="space-y-4">
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.aiEndpoint') }}</label>
              <input v-model="settings.ai_endpoint" class="input" placeholder="https://api.openai.com/v1" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.aiApiKey') }}</label>
              <input v-model="settings.ai_api_key" type="password" class="input" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.aiModel') }}</label>
              <input v-model="settings.ai_model" class="input" placeholder="gpt-3.5-turbo" />
            </div>
          </div>

          <!-- Summary -->
          <div v-if="activeTab === 'summary'" class="space-y-4">
            <div>
              <label class="flex items-center gap-2">
                <input v-model="settings.summary_enabled" type="checkbox" class="rounded" />
                <span class="text-sm font-medium">{{ t('settings.summaryEnabled') }}</span>
              </label>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.summaryProvider') }}</label>
              <select v-model="settings.summary_provider" class="input">
                <option value="local">Local (TF-IDF)</option>
                <option value="ai">AI Summary</option>
              </select>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">Max Summary Length</label>
              <input v-model="settings.summary_max_len" type="number" class="input" placeholder="200" />
            </div>
          </div>

          <!-- Network -->
          <div v-if="activeTab === 'network'" class="space-y-4">
            <div>
              <label class="flex items-center gap-2">
                <input v-model="settings.proxy_enabled" type="checkbox" class="rounded" />
                <span class="text-sm font-medium">{{ t('settings.proxyEnabled') }}</span>
              </label>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.proxyUrl') }}</label>
              <input v-model="settings.proxy_url" class="input" placeholder="socks5://localhost:1080" />
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.requestTimeout') }}</label>
              <input v-model="settings.request_timeout" type="number" class="input" placeholder="30" />
            </div>
          </div>

          <!-- Storage -->
          <div v-if="activeTab === 'storage'" class="space-y-4">
            <div>
              <label class="flex items-center gap-2">
                <input v-model="settings.auto_cleanup_enabled" type="checkbox" class="rounded" />
                <span class="text-sm font-medium">{{ t('settings.autoCleanup') }}</span>
              </label>
            </div>
            <div>
              <label class="block text-sm font-medium mb-1">{{ t('settings.keepArticlesDays') }}</label>
              <input v-model="settings.max_article_age_days" type="number" class="input" placeholder="30" />
            </div>
            <div>
              <label class="flex items-center gap-2">
                <input v-model="settings.media_cache_enabled" type="checkbox" class="rounded" />
                <span class="text-sm font-medium">{{ t('settings.mediaCache') }}</span>
              </label>
            </div>
          </div>

          <!-- Integrations -->
          <div v-if="activeTab === 'integrations'" class="space-y-4">
            <div class="p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
              <h3 class="font-medium mb-2">Obsidian</h3>
              <input v-model="settings.obsidian_vault_path" class="input" :placeholder="t('settings.obsidianVaultPath')" />
            </div>
            <div class="p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
              <h3 class="font-medium mb-2">Notion</h3>
              <input v-model="settings.notion_api_key" type="password" class="input mb-2" :placeholder="t('settings.notionApiKey')" />
              <input v-model="settings.notion_page_id" class="input" :placeholder="t('settings.notionPageId')" />
            </div>
            <div class="p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
              <h3 class="font-medium mb-2">FreshRSS</h3>
              <input v-model="settings.freshrss_server" class="input mb-2" :placeholder="t('settings.freshrssServer')" />
              <input v-model="settings.freshrss_username" class="input mb-2" :placeholder="t('settings.freshrssUsername')" />
              <input v-model="settings.freshrss_password" type="password" class="input" :placeholder="t('settings.freshrssPassword')" />
            </div>
            <div class="p-3 bg-gray-50 dark:bg-gray-700 rounded-lg">
              <h3 class="font-medium mb-2">RSSHub</h3>
              <input v-model="settings.rsshub_endpoint" class="input" :placeholder="t('settings.rsshubEndpoint')" />
            </div>
          </div>

          <!-- Rules -->
          <div v-if="activeTab === 'rules'" class="space-y-4">
            <p class="text-sm text-gray-500">{{ t('settings.configureRules') }}</p>
            <button class="btn btn-primary">{{ t('settings.addRule') }}</button>
          </div>

          <!-- Shortcuts -->
          <div v-if="activeTab === 'shortcuts'" class="space-y-2">
            <div v-for="(shortcut, key) in {
              'Ctrl+N': t('sidebar.addFeed'),
              'Ctrl+,': t('common.settings'),
              'Ctrl+R': t('common.refresh'),
              'Ctrl+Shift+A': t('article.markAllAsRead'),
              '↑/↓': t('article.noArticles'),
              'Enter': t('article.noArticles'),
              'Escape': t('common.close'),
              'F': t('article.favorite'),
              'L': t('article.readLater'),
              'H': t('article.hide'),
              'T': t('article.translate'),
              'S': t('article.summarize')
            }" :key="key" class="flex items-center justify-between py-2 border-b border-gray-100 dark:border-gray-700">
              <span class="text-sm">{{ shortcut }}</span>
              <kbd class="px-2 py-1 text-xs bg-gray-100 dark:bg-gray-700 rounded">{{ key }}</kbd>
            </div>
          </div>

          <!-- Tags -->
          <div v-if="activeTab === 'tags'" class="space-y-4">
            <p class="text-sm text-gray-500">{{ t('settings.manageTags') }}</p>
            <button class="btn btn-primary">{{ t('settings.addTag') }}</button>
          </div>

          <!-- Statistics -->
          <div v-if="activeTab === 'statistics'" class="space-y-4">
            <p class="text-sm text-gray-500">{{ t('settings.viewStats') }}</p>
          </div>

          <!-- About -->
          <div v-if="activeTab === 'about'" class="space-y-4 text-center">
            <h3 class="text-xl font-bold">OneRSS</h3>
            <p class="text-sm text-gray-500">{{ t('settings.version') }} 1.0.0</p>
            <p class="text-sm text-gray-500">{{ t('settings.description') }}</p>
          </div>
        </div>
      </div>

      <!-- Footer -->
      <div class="flex justify-end gap-2 p-4 border-t border-gray-200 dark:border-gray-700">
        <button @click="emit('close')" class="btn btn-secondary">
          {{ t('common.cancel') }}
        </button>
        <button 
          @click="saveSettings" 
          :class="['btn', saveSuccess ? 'btn-success' : 'btn-primary']"
          :disabled="saveSuccess"
        >
          {{ saveSuccess ? '✓ ' + t('settings.saved') : t('common.save') }}
        </button>
      </div>
    </div>
  </div>
</template>
