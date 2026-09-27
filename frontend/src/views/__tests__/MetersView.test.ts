import { beforeEach, describe, expect, it, vi } from 'vitest'
import { flushPromises, mount } from '@vue/test-utils'
import { createPinia, setActivePinia } from 'pinia'
import type { MeterSummary } from '@/types/api'
import { meter } from '@/utils/__tests__/fixtures'

const getMeters = vi.fn<() => Promise<MeterSummary[]>>()
vi.mock('@/services/api', async (orig) => ({
  ...(await orig<typeof import('@/services/api')>()),
  getMeters: () => getMeters(),
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
    getMeters.mockResolvedValue([
      meter('M-101', { status: 'UNEVALUATED' }),
      meter('M-102', { status: 'UNEVALUATED' }),
    ])
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
    getMeters.mockResolvedValue([
      meter('M-109', { status: 'CRITICAL' }),
      meter('M-112', { status: 'ALERT' }),
      meter('M-101'),
    ])
    const w = mount(MetersView, { global: { stubs } })
    await flushPromises()

    expect(chips(w).every((c) => c.attributes('disabled') === undefined)).toBe(true)
    expect(chips(w).map((c) => c.text())).toEqual(['Todos3', 'Normales1', 'Alertas1', 'Críticos1'])
    expect(w.text()).not.toContain('aún no se han evaluado')
  })
})
