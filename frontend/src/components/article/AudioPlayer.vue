<script setup lang="ts">
import { ref, onMounted, onUnmounted, watch } from 'vue'
import { PhPlay, PhPause, PhSpeakerHigh, PhSpeakerX } from '@phosphor-icons/vue'

const props = defineProps<{
  src: string
}>()

const audio = ref<HTMLAudioElement | null>(null)
const isPlaying = ref(false)
const currentTime = ref(0)
const duration = ref(0)
const volume = ref(1)
const isMuted = ref(false)
const playbackRate = ref(1)

const rates = [0.5, 0.75, 1, 1.25, 1.5, 2]

onMounted(() => {
  audio.value = new Audio(props.src)
  
  audio.value.addEventListener('timeupdate', () => {
    currentTime.value = audio.value?.currentTime || 0
  })
  
  audio.value.addEventListener('loadedmetadata', () => {
    duration.value = audio.value?.duration || 0
  })
  
  audio.value.addEventListener('ended', () => {
    isPlaying.value = false
  })
})

onUnmounted(() => {
  audio.value?.pause()
  audio.value = null
})

function togglePlay() {
  if (!audio.value) return
  
  if (isPlaying.value) {
    audio.value.pause()
  } else {
    audio.value.play()
  }
  isPlaying.value = !isPlaying.value
}

function seek(e: MouseEvent) {
  if (!audio.value || !duration.value) return
  
  const rect = (e.target as HTMLElement).getBoundingClientRect()
  const x = e.clientX - rect.left
  const percent = x / rect.width
  audio.value.currentTime = percent * duration.value
}

function toggleMute() {
  if (!audio.value) return
  
  isMuted.value = !isMuted.value
  audio.value.muted = isMuted.value
}

function setVolume(e: Event) {
  if (!audio.value) return
  
  const value = parseFloat((e.target as HTMLInputElement).value)
  volume.value = value
  audio.value.volume = value
}

function setPlaybackRate(rate: number) {
  if (!audio.value) return
  
  playbackRate.value = rate
  audio.value.playbackRate = rate
}

function formatTime(seconds: number): string {
  const mins = Math.floor(seconds / 60)
  const secs = Math.floor(seconds % 60)
  return `${mins}:${secs.toString().padStart(2, '0')}`
}
</script>

<template>
  <div class="bg-gray-50 dark:bg-gray-800 rounded-lg p-4">
    <!-- Progress Bar -->
    <div 
      class="h-1.5 bg-gray-200 dark:bg-gray-700 rounded-full cursor-pointer mb-3"
      @click="seek"
    >
      <div 
        class="h-full bg-primary-500 rounded-full transition-all"
        :style="{ width: duration ? (currentTime / duration * 100) + '%' : '0%' }"
      />
    </div>

    <!-- Controls -->
    <div class="flex items-center gap-4">
      <!-- Play/Pause -->
      <button 
        @click="togglePlay"
        class="p-2 rounded-full bg-primary-500 text-white hover:bg-primary-600 transition-colors"
      >
        <PhPause v-if="isPlaying" :size="20" />
        <PhPlay v-else :size="20" />
      </button>

      <!-- Time -->
      <div class="text-sm text-gray-600 dark:text-gray-400">
        {{ formatTime(currentTime) }} / {{ formatTime(duration) }}
      </div>

      <!-- Playback Rate -->
      <div class="flex items-center gap-1">
        <button
          v-for="rate in rates"
          :key="rate"
          @click="setPlaybackRate(rate)"
          :class="[
            'px-1.5 py-0.5 text-xs rounded transition-colors',
            playbackRate === rate
              ? 'bg-primary-500 text-white'
              : 'bg-gray-200 dark:bg-gray-700 hover:bg-gray-300 dark:hover:bg-gray-600'
          ]"
        >
          {{ rate }}x
        </button>
      </div>

      <!-- Volume -->
      <div class="flex items-center gap-2 ml-auto">
        <button 
          @click="toggleMute"
          class="p-1 text-gray-500 hover:text-gray-700 dark:hover:text-gray-300"
        >
          <PhSpeakerX v-if="isMuted" :size="18" />
          <PhSpeakerHigh v-else :size="18" />
        </button>
        <input 
          type="range" 
          min="0" 
          max="1" 
          step="0.1" 
          :value="isMuted ? 0 : volume"
          @input="setVolume"
          class="w-20"
        />
      </div>
    </div>
  </div>
</template>
