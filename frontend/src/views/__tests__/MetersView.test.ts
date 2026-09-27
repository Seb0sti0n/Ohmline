import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import type { MeterPage } from '@/types/api'
import { meter, meterPage } from '@/utils/__tests__/fixtures'

const getMeters = vi.fn<(q?: unknown) => Promise<MeterPage>>()
vi.mock('@/services/api', async (orig) => ({
  ...(await orig<typeof import('@/services/api')>()),
  getMeters: (q?: unknown) => getMeters(q),
}))

import MetersView from '../MetersView.vue'

const stubs = {
  RouterLink: { props: ['to'], template: '<a :href="String(to)"><slot /></a>' },
  MetersTable: true,
}

const chips = (w: ReturnType<typeof mount>) => w.findAll('[role=group] button')

beforeEach(() => {
  setActivePinia(createPinia())
  getMeters.mockReset()
})

describe('MetersView status filters', () => {
  it('locks the status filters and explains why until the first analysis', async () => {
    getMeters.mockResolvedValue(
      meterPage([
        meter('M-101', { status: 'UNEVALUATED' }),
        meter('M-102', { status: 'UNEVALUATED' }),
      ]),
    )
    const w = mount(MetersView, { global: { stubs } })
    await flushPromises()

    const [all, ...statuses] = chips(w)
    expect(all.attributes('disabled')).toBeUndefined()
    expect(all.text()).toContain('2') // the total is real
    for (const chip of statuses) {
      expect(chip.attributes('disabled')).toBeDefined()
      expect(chip.attributes('title')).toContain('análisis')
      expect(chip.text()).not.toMatch(/\d/) // no misleading "0" counters
    }
    expect(w.text()).toContain('Los medidores aún no se han evaluado')
    expect(w.find('a[href="/"]').text()).toBe('Ejecuta el análisis IA')
  })

  it('unlocks them, with real counters, once there is an analysis', async () => {
    getMeters.mockResolvedValue(
      meterPage([
        meter('M-109', { status: 'CRITICAL' }),
        meter('M-112', { status: 'ALERT' }),
        meter('M-101'),
      ]),
    )
    const w = mount(MetersView, { global: { stubs } })
    await flushPromises()

    expect(chips(w).every((c) => c.attributes('disabled') === undefined)).toBe(true)
    expect(chips(w).map((c) => c.text())).toEqual(['Todos3', 'Normales1', 'Alertas1', 'Críticos1'])
    expect(w.text()).not.toContain('aún no se han evaluado')
  })

  it('shows the pager under the table, and hides it when nothing matches', async () => {
    const many = Array.from({ length: 10 }, (_, i) => meter(`M-${100 + i}`))
    getMeters.mockResolvedValue(
      meterPage(many, {
        total: 23,
        counts: { ALL: 23, OK: 23, ALERT: 0, CRITICAL: 0, UNEVALUATED: 0 },
      }),
    )
    const w = mount(MetersView, { global: { stubs } })
    await flushPromises()
    expect(w.text()).toContain('Mostrando 1–10 de 23 medidores')
    expect(
      w
        .findAll('nav button[aria-label^="Página "]')
        .filter((b) => /^\d+$/.test(b.text()))
        .map((b) => b.text()),
    ).toEqual(['1', '2', '3'])

    getMeters.mockResolvedValue(
      meterPage([], {
        total: 0,
        counts: { ALL: 23, OK: 23, ALERT: 0, CRITICAL: 0, UNEVALUATED: 0 },
      }),
    )
    await w.find('input[type=search]').setValue('zzz')
    await new Promise((r) => setTimeout(r, 300)) // the search waits for a pause in typing
    await flushPromises()
    expect(w.text()).toContain('Ningún medidor coincide')
    expect(w.text()).not.toContain('Mostrando')
  })
})
