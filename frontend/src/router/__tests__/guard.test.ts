import { beforeEach, describe, expect, it } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import router from '../index'
import { useAuthStore } from '@/stores/auth'
import { saveSession } from '@/services/session'

beforeEach(async () => {
  localStorage.clear()
  setActivePinia(createPinia())
  await router.replace('/login').catch(() => {})
})

describe('route guard', () => {
  it.each(['/', '/meters', '/meters/M-109', '/anomalies', '/anomalies/1'])(
    'sends %s to login without a session',
    async (path) => {
      await router.push(path)
      expect(router.currentRoute.value.name).toBe('login')
    },
  )

  it('lets a logged-in user in and keeps them off the login page', async () => {
    saveSession('jwt', 'demo@energy.io')
    setActivePinia(createPinia()) // the store reads the saved session
    expect(useAuthStore().isAuthenticated).toBe(true)
    await router.push('/meters')
    expect(router.currentRoute.value.name).toBe('meters')
    await router.push('/login')
    expect(router.currentRoute.value.name).toBe('dashboard')
  })

  it('redirects unknown paths to the dashboard (then to login if needed)', async () => {
    await router.push('/does-not-exist')
    expect(router.currentRoute.value.name).toBe('login')
    saveSession('jwt', 'demo@energy.io')
    setActivePinia(createPinia())
    await router.push('/does-not-exist')
    expect(router.currentRoute.value.name).toBe('dashboard')
  })
})
