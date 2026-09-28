import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'

const login = vi.fn()
vi.mock('@/services/api', async (orig) => ({
  ...(await orig<typeof import('@/services/api')>()),
  login: (e: string, p: string) => login(e, p),
}))

const push = vi.fn()
vi.mock('vue-router', () => ({ useRouter: () => ({ push }) }))

import LoginForm from '../LoginForm.vue'

const httpError = (status: number) =>
  Object.assign(new Error('x'), { isAxiosError: true, response: { status, data: {} } })

beforeEach(() => {
  localStorage.clear()
  setActivePinia(createPinia())
  login.mockReset()
  push.mockReset()
})

describe('LoginForm', () => {
  it('comes prefilled with the demo credentials', () => {
    const w = mount(LoginForm)
    expect((w.find('input[type=email]').element as HTMLInputElement).value).toBe('demo@energy.io')
    expect((w.find('input[type=password]').element as HTMLInputElement).value).toBe('Ohmline#2026')
  })

  it('logs in and goes to the dashboard', async () => {
    login.mockResolvedValue({ token: 't', user: { email: 'demo@energy.io' } })
    const w = mount(LoginForm)
    await w.find('form').trigger('submit')
    await flushPromises()
    expect(login).toHaveBeenCalledWith('demo@energy.io', 'Ohmline#2026')
    expect(push).toHaveBeenCalledWith({ name: 'dashboard' })
    expect(w.find('[role=alert]').exists()).toBe(false)
  })

  it('shows a clear message for wrong credentials and stays on the page', async () => {
    login.mockRejectedValue(httpError(401))
    const w = mount(LoginForm)
    await w.find('input[type=password]').setValue('wrong')
    await w.find('form').trigger('submit')
    await flushPromises()
    expect(w.find('[role=alert]').text()).toBe('Correo o contraseña incorrectos')
    expect(push).not.toHaveBeenCalled()
    expect(w.find('button[type=submit]').attributes('disabled')).toBeUndefined() // can try again
  })

  it('shows a different message when the server cannot be reached', async () => {
    login.mockRejectedValue(Object.assign(new Error('Network Error'), { isAxiosError: true }))
    const w = mount(LoginForm)
    await w.find('form').trigger('submit')
    await flushPromises()
    expect(w.find('[role=alert]').text()).toBe(
      'No se pudo iniciar sesión. El servidor no responde.',
    )
  })

  it('disables the button while signing in, so it cannot be submitted twice', async () => {
    let finish!: (v: unknown) => void
    login.mockReturnValue(new Promise((r) => (finish = r)))
    const w = mount(LoginForm)
    await w.find('form').trigger('submit')
    expect(w.find('button[type=submit]').attributes('disabled')).toBeDefined()
    await w.find('form').trigger('submit') // ignored
    expect(login).toHaveBeenCalledTimes(1)
    finish({ token: 't', user: { email: 'demo@energy.io' } })
    await flushPromises()
  })
})
