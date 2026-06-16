<script setup lang="ts">
import { ref, computed } from 'vue'

const props = defineProps<{
  src: string
  type?: 'video' | 'youtube' | 'vimeo'
}>()

const isPlaying = ref(false)

const embedUrl = computed(() => {
  if (props.type === 'youtube') {
    // Extract YouTube video ID
    const match = props.src.match(/(?:youtube\.com\/watch\?v=|youtu\.be\/)([^&\s]+)/)
    if (match) {
      return `https://www.youtube.com/embed/${match[1]}`
    }
  } else if (props.type === 'vimeo') {
    // Extract Vimeo video ID
    const match = props.src.match(/vimeo\.com\/(\d+)/)
    if (match) {
      return `https://player.vimeo.com/video/${match[1]}`
    }
  }
  return props.src
})
</script>

<template>
  <div class="relative aspect-video rounded-lg overflow-hidden bg-black">
    <!-- YouTube / Vimeo Embed -->
    <iframe 
      v-if="type === 'youtube' || type === 'vimeo'"
      :src="embedUrl"
      class="absolute inset-0 w-full h-full"
      frameborder="0"
      allowfullscreen
      allow="accelerometer; autoplay; clipboard-write; encrypted-media; gyroscope; picture-in-picture"
    />

    <!-- Direct Video -->
    <video 
      v-else
      :src="src"
      controls
      class="absolute inset-0 w-full h-full object-contain"
    />
  </div>
</template>
