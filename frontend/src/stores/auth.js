import { computed, ref } from 'vue'
import { defineStore } from 'pinia'

const SESSION_KEY = 'cinema-booking-session'

function loadSession() {
  try {
    return JSON.parse(localStorage.getItem(SESSION_KEY) || '{}')
  } catch {
    return {}
  }
}

export const useAuthStore = defineStore('auth', () => {
  const saved = loadSession()
  const token = ref(saved.token || '')
  const user = ref(saved.user || null)

  const isAuthenticated = computed(() => Boolean(token.value))
  const isAdmin = computed(() => user.value?.role === 'ADMIN')

  function setSession(nextToken, nextUser) {
    token.value = nextToken
    user.value = nextUser
    localStorage.setItem(SESSION_KEY, JSON.stringify({ token: nextToken, user: nextUser }))
  }

  function logout() {
    token.value = ''
    user.value = null
    localStorage.removeItem(SESSION_KEY)
  }

  return {
    token,
    user,
    isAuthenticated,
    isAdmin,
    setSession,
    logout,
  }
})
