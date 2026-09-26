import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { login as apiLogin } from '@/services/api'
import { clearSession, getEmail, getToken, saveSession } from '@/services/session'

export const useAuthStore = defineStore('auth', () => {
  const token = ref<string | null>(getToken())
  const email = ref(getEmail())
  const isAuthenticated = computed(() => !!token.value)

  async function login(userEmail: string, password: string) {
    const res = await apiLogin(userEmail, password)
    saveSession(res.token, res.user.email)
    token.value = res.token
    email.value = res.user.email
  }

  function logout() {
    clearSession()
    token.value = null
    email.value = ''
  }

  return { token, email, isAuthenticated, login, logout }
})
