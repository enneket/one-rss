<script setup lang="ts">
import { ref, onMounted, computed } from 'vue'
import { PhList } from '@phosphor-icons/vue'

interface TocItem {
  id: string
  text: string
  level: number
}

const props = defineProps<{
  content: string
}>()

const isOpen = ref(false)
const activeId = ref('')

const tocItems = computed(() => {
  const items: TocItem[] = []
  const regex = /<h([1-6])[^>]*id="([^"]*)"[^>]*>(.*?)<\/h[1-6]>/gi
  let match

  while ((match = regex.exec(props.content)) !== null) {
    items.push({
      level: parseInt(match[1]),
      id: match[2],
      text: match[3].replace(/<[^>]+>/g, '')
    })
  }

  return items
})

function scrollTo(id: string) {
  const element = document.getElementById(id)
  if (element) {
    element.scrollIntoView({ behavior: 'smooth' })
    activeId.value = id
  }
}
</script>

<template>
  <div class="relative">
    <!-- Toggle Button -->
    <button 
      @click="isOpen = !isOpen"
      class="p-2 rounded-lg bg-white dark:bg-gray-800 shadow-lg border border-gray-200 dark:border-gray-700 hover:bg-gray-50 dark:hover:bg-gray-700 transition-colors"
    >
      <PhList :size="20" />
    </button>

    <!-- TOC Panel -->
    <div 
      v-if="isOpen && tocItems.length > 0"
      class="absolute right-0 top-12 w-64 max-h-[60vh] overflow-y-auto bg-white dark:bg-gray-800 rounded-lg shadow-xl border border-gray-200 dark:border-gray-700"
    >
      <div class="p-3">
        <h3 class="text-sm font-semibold text-gray-700 dark:text-gray-300 mb-2">
          Table of Contents
        </h3>
        <nav class="space-y-1">
          <button
            v-for="item in tocItems"
            :key="item.id"
            @click="scrollTo(item.id)"
            :class="[
              'block w-full text-left text-sm py-1 px-2 rounded transition-colors',
              activeId === item.id
                ? 'bg-primary-50 text-primary-600 dark:bg-primary-900/20 dark:text-primary-400'
                : 'text-gray-600 dark:text-gray-400 hover:bg-gray-100 dark:hover:bg-gray-700'
            ]"
            :style="{ paddingLeft: (item.level - 1) * 12 + 8 + 'px' }"
          >
            {{ item.text }}
          </button>
        </nav>
      </div>
    </div>
  </div>
</template>
