import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'

const login = vi.fn()
vi.mock('@/services/api', async (orig) => ({
  ...(await orig<typeof import('@/services/api')>()),
  login: (e: string, p: string) => login(e, p),
}))

import { useAuthStore } from '../auth'
import { getToken } from '@/services/session'

beforeEach(() => {
  localStorage.clear()
  setActivePinia(createPinia())
  login.mockReset()
})

describe('auth store', () => {
  it('logs in, keeps the session and survives a reload', async () => {
    login.mockResolvedValue({ token: 'jwt-123', user: { email: 'demo@energy.io' } })
    const auth = useAuthStore()
    expect(auth.isAuthenticated).toBe(false)
    await auth.login('demo@energy.io', 'demo123')
    expect(login).toHaveBeenCalledWith('demo@energy.io', 'demo123')
    expect(auth.isAuthenticated).toBe(true)
    expect(getToken()).toBe('jwt-123')

    setActivePinia(createPinia()) // a new page load reads the saved session
    expect(useAuthStore().isAuthenticated).toBe(true)
    expect(useAuthStore().email).toBe('demo@energy.io')
  })

  it('does not authenticate when the login fails', async () => {
    login.mockRejectedValue(new Error('401'))
    const auth = useAuthStore()
    await expect(auth.login('a@b.c', 'x')).rejects.toThrow()
    expect(auth.isAuthenticated).toBe(false)
    expect(getToken()).toBeNull()
  })

  it('logout clears everything', async () => {
    login.mockResolvedValue({ token: 'jwt-123', user: { email: 'demo@energy.io' } })
    const auth = useAuthStore()
    await auth.login('demo@energy.io', 'demo123')
    auth.logout()
    expect(auth.isAuthenticated).toBe(false)
    expect(auth.email).toBe('')
    expect(getToken()).toBeNull()
  })
})
