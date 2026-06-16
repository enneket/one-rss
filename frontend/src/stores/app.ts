import { defineStore } from 'pinia'
import { ref, computed } from 'vue'
import axios from 'axios'

export interface Feed {
  id: number
  title: string
  url: string
  link: string
  description: string
  category: string
  image_url: string
  position: number
  last_updated: string
  last_error: string
  hide_from_timeline: boolean
  proxy_url: string
  proxy_enabled: boolean
  refresh_interval: number
  is_image_mode: boolean
  type: string
  article_view_mode: string
  auto_expand_content: string
  tags?: Tag[]
}

export interface Article {
  id: number
  feed_id: number
  title: string
  url: string
  image_url: string
  audio_url: string
  video_url: string
  published_at: string
  is_read: boolean
  is_favorite: boolean
  is_hidden: boolean
  is_read_later: boolean
  feed_title: string
  author: string
  translated_title: string
  summary: string
  unique_id: string
}

export interface Tag {
  id: number
  name: string
  color: string
  position: number
}

export interface UnreadCounts {
  total: number
  feedCounts: Record<number, number>
}

export type Filter = 'all' | 'unread' | 'favorites' | 'readLater' | 'imageGallery' | ''

export const useAppStore = defineStore('app', () => {
  const articles = ref<Article[]>([])
  const feeds = ref<Feed[]>([])
  const tags = ref<Tag[]>([])
  const unreadCounts = ref<UnreadCounts>({ total: 0, feedCounts: {} })
  const currentFilter = ref<Filter>('all')
  const currentFeedId = ref<number | null>(null)
  const currentCategory = ref<string | null>(null)
  const currentArticleId = ref<number | null>(null)
  const isLoading = ref(false)
  const page = ref(1)
  const hasMore = ref(true)
  const searchQuery = ref('')

  const feedMap = computed(() => {
    const map: Record<number, Feed> = {}
    feeds.value.forEach(feed => {
      map[feed.id] = feed
    })
    return map
  })

  const tagMap = computed(() => {
    const map: Record<number, Tag> = {}
    tags.value.forEach(tag => {
      map[tag.id] = tag
    })
    return map
  })

  const filteredArticles = computed(() => {
    let filtered = articles.value

    if (currentFeedId.value) {
      filtered = filtered.filter(a => a.feed_id === currentFeedId.value)
    }

    if (searchQuery.value) {
      const query = searchQuery.value.toLowerCase()
      filtered = filtered.filter(a =>
        a.title.toLowerCase().includes(query) ||
        a.author?.toLowerCase().includes(query) ||
        a.feed_title?.toLowerCase().includes(query)
      )
    }

    return filtered
  })

  async function fetchFeeds() {
    try {
      const response = await axios.get('/api/feeds')
      feeds.value = response.data
    } catch (error) {
      console.error('Failed to fetch feeds:', error)
    }
  }

  async function fetchArticles(append = false) {
    try {
      isLoading.value = true
      const params: Record<string, any> = {
        page: page.value,
        limit: 50
      }

      if (currentFeedId.value) {
        params.feed_id = currentFeedId.value
      }

      if (currentFilter.value) {
        params.filter = currentFilter.value
      }

      const response = await axios.get('/api/articles', { params })
      
      if (append) {
        articles.value = [...articles.value, ...response.data]
      } else {
        articles.value = response.data
      }

      hasMore.value = response.data.length === 50
    } catch (error) {
      console.error('Failed to fetch articles:', error)
    } finally {
      isLoading.value = false
    }
  }

  async function fetchUnreadCounts() {
    try {
      const response = await axios.get('/api/articles/unread-counts')
      unreadCounts.value = response.data
    } catch (error) {
      console.error('Failed to fetch unread counts:', error)
    }
  }

  async function fetchTags() {
    try {
      const response = await axios.get('/api/tags')
      tags.value = response.data
    } catch (error) {
      console.error('Failed to fetch tags:', error)
    }
  }

  function setFilter(filter: Filter) {
    currentFilter.value = filter
    currentFeedId.value = null
    currentCategory.value = null
    page.value = 1
    fetchArticles()
  }

  function setFeed(feedId: number | null) {
    currentFeedId.value = feedId
    currentFilter.value = ''
    page.value = 1
    fetchArticles()
  }

  function setCategory(category: string | null) {
    currentCategory.value = category
    currentFeedId.value = null
    currentFilter.value = ''
    page.value = 1
    fetchArticles()
  }

  async function loadMore() {
    if (!hasMore.value || isLoading.value) return
    page.value++
    await fetchArticles(true)
  }

  async function markAsRead(articleId: number) {
    try {
      await axios.post('/api/articles/read', { id: articleId, is_read: true })
      const article = articles.value.find(a => a.id === articleId)
      if (article) {
        article.is_read = true
      }
      await fetchUnreadCounts()
    } catch (error) {
      console.error('Failed to mark as read:', error)
    }
  }

  async function toggleFavorite(articleId: number) {
    try {
      await axios.post('/api/articles/favorite', { id: articleId })
      const article = articles.value.find(a => a.id === articleId)
      if (article) {
        article.is_favorite = !article.is_favorite
      }
    } catch (error) {
      console.error('Failed to toggle favorite:', error)
    }
  }

  async function toggleReadLater(articleId: number) {
    try {
      await axios.post('/api/articles/toggle-read-later', { id: articleId })
      const article = articles.value.find(a => a.id === articleId)
      if (article) {
        article.is_read_later = !article.is_read_later
      }
    } catch (error) {
      console.error('Failed to toggle read later:', error)
    }
  }

  async function toggleHide(articleId: number) {
    try {
      await axios.post('/api/articles/toggle-hide', { id: articleId })
      const article = articles.value.find(a => a.id === articleId)
      if (article) {
        article.is_hidden = !article.is_hidden
      }
    } catch (error) {
      console.error('Failed to toggle hide:', error)
    }
  }

  async function markAllAsRead(feedId?: number) {
    try {
      await axios.post('/api/articles/mark-all-read', { feed_id: feedId })
      articles.value.forEach(article => {
        if (!feedId || article.feed_id === feedId) {
          article.is_read = true
        }
      })
      await fetchUnreadCounts()
    } catch (error) {
      console.error('Failed to mark all as read:', error)
    }
  }

  async function addFeed(url: string, category?: string) {
    try {
      const response = await axios.post('/api/feeds/add', { url, category })
      await fetchFeeds()
      await fetchUnreadCounts()
      return response.data
    } catch (error) {
      console.error('Failed to add feed:', error)
      throw error
    }
  }

  async function deleteFeed(feedId: number) {
    try {
      await axios.post('/api/feeds/delete', { id: feedId })
      feeds.value = feeds.value.filter(f => f.id !== feedId)
      if (currentFeedId.value === feedId) {
        currentFeedId.value = null
      }
      await fetchArticles()
      await fetchUnreadCounts()
    } catch (error) {
      console.error('Failed to delete feed:', error)
      throw error
    }
  }

  function selectArticle(articleId: number) {
    currentArticleId.value = articleId
    markAsRead(articleId)
  }

  async function refreshAll() {
    try {
      await axios.post('/api/refresh')
      await fetchArticles()
      await fetchUnreadCounts()
    } catch (error) {
      console.error('Failed to refresh:', error)
    }
  }

  return {
    articles,
    feeds,
    tags,
    unreadCounts,
    currentFilter,
    currentFeedId,
    currentCategory,
    currentArticleId,
    isLoading,
    page,
    hasMore,
    searchQuery,
    feedMap,
    tagMap,
    filteredArticles,
    fetchFeeds,
    fetchArticles,
    fetchUnreadCounts,
    fetchTags,
    setFilter,
    setFeed,
    setCategory,
    loadMore,
    markAsRead,
    toggleFavorite,
    toggleReadLater,
    toggleHide,
    markAllAsRead,
    addFeed,
    deleteFeed,
    selectArticle,
    refreshAll
  }
})
