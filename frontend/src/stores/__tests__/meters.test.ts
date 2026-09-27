import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import type { MeterQuery, MeterSummary } from '@/types/api'
import { meter } from '@/utils/__tests__/fixtures'

const getMeters = vi.fn<(q?: MeterQuery) => Promise<MeterSummary[]>>()
vi.mock('@/services/api', async (orig) => ({
  ...(await orig<typeof import('@/services/api')>()),
  getMeters: (q?: MeterQuery) => getMeters(q),
}))

import { useMetersStore } from '../meters'

const everyone = [
  meter('M-109', { status: 'CRITICAL' }),
  meter('M-112', { status: 'ALERT' }),
  meter('M-104', { status: 'ALERT' }),
  meter('M-101'),
  meter('M-102'),
]

beforeEach(() => {
  setActivePinia(createPinia())
  getMeters.mockReset()
  getMeters.mockResolvedValue(everyone)
})

describe('meters store', () => {
  it('loads the table and the counters', async () => {
    const s = useMetersStore()
    await s.load()
    expect(s.rows).toHaveLength(5)
    expect(s.counts).toEqual({ ALL: 5, OK: 2, ALERT: 2, CRITICAL: 1, UNEVALUATED: 0 })
    expect(s.unevaluated).toBe(false)
    // default query: severity, descending, nothing else
    expect(getMeters).toHaveBeenCalledWith({
      status: undefined,
      search: undefined,
      sort: 'severity',
      order: 'desc',
    })
  })

  it('sends the filter and search to the API, trimming the search', async () => {
    const s = useMetersStore()
    await s.load()
    await s.setFilter('ALERT')
    expect(getMeters).toHaveBeenLastCalledWith(expect.objectContaining({ status: 'ALERT' }))
    await s.setSearch('  11 ')
    expect(getMeters).toHaveBeenLastCalledWith(
      expect.objectContaining({ status: 'ALERT', search: '11' }),
    )
    await s.setFilter('ALL')
    expect(getMeters).toHaveBeenLastCalledWith(
      expect.objectContaining({ status: undefined, search: '11' }),
    )
  })

  it('keeps the counters when the table is filtered', async () => {
    const s = useMetersStore()
    await s.load()
    getMeters.mockResolvedValue([meter('M-109', { status: 'CRITICAL' })])
    await s.setFilter('CRITICAL')
    expect(s.rows).toHaveLength(1)
    expect(s.counts.ALL).toBe(5)
  })

  it('toggles the direction on the active column and starts descending on a new one', async () => {
    const s = useMetersStore()
    await s.load()
    await s.toggleSort('severity')
    expect([s.sort, s.order]).toEqual(['severity', 'asc'])
    await s.toggleSort('severity')
    expect([s.sort, s.order]).toEqual(['severity', 'desc'])
    await s.toggleSort('variation')
    expect([s.sort, s.order]).toEqual(['variation', 'desc'])
    await s.toggleSort('variation')
    expect(getMeters).toHaveBeenLastCalledWith(
      expect.objectContaining({ sort: 'variation', order: 'asc' }),
    )
  })

  it('ignores a slow response that arrives after a newer one', async () => {
    const s = useMetersStore()
    await s.load()
    let releaseSlow!: (v: MeterSummary[]) => void
    getMeters.mockReturnValueOnce(new Promise((r) => (releaseSlow = r)))
    const slow = s.setSearch('1') // sent first, answered last
    getMeters.mockResolvedValueOnce([meter('M-112')])
    await s.setSearch('112') // sent second, answered first
    releaseSlow([meter('M-101'), meter('M-102')])
    await slow
    expect(s.rows.map((m) => m.meter_id)).toEqual(['M-112'])
  })

  it('reports an error without wiping the rows already shown', async () => {
    const s = useMetersStore()
    await s.load()
    getMeters.mockRejectedValueOnce(new Error('boom'))
    await s.setFilter('OK')
    expect(s.error).toBeTruthy()
    expect(s.rows).toHaveLength(5)
    await s.setFilter('ALL')
    expect(s.error).toBeNull()
  })

  describe('before the first analysis', () => {
    const fresh = [
      meter('M-101', { status: 'UNEVALUATED' }),
      meter('M-102', { status: 'UNEVALUATED' }),
    ]

    it('knows that no meter has a status yet', async () => {
      getMeters.mockResolvedValue(fresh)
      const s = useMetersStore()
      await s.load()
      expect(s.unevaluated).toBe(true)
      expect(s.counts).toMatchObject({ ALL: 2, UNEVALUATED: 2, OK: 0, ALERT: 0, CRITICAL: 0 })
    })

    it('is not "unevaluated" when the list is empty or only some meters lack a status', async () => {
      const s = useMetersStore()
      getMeters.mockResolvedValue([])
      await s.load()
      expect(s.unevaluated).toBe(false)
      getMeters.mockResolvedValue([meter('M-101', { status: 'UNEVALUATED' }), meter('M-102')])
      await s.load()
      expect(s.unevaluated).toBe(false)
    })

    it('drops a leftover status filter, which would show an empty table', async () => {
      const s = useMetersStore()
      await s.load()
      await s.setFilter('CRITICAL') // chosen while an analysis existed
      getMeters.mockResolvedValue(fresh) // ...then the data was reset
      await s.load()
      expect(s.filter).toBe('ALL')
      expect(getMeters).toHaveBeenLastCalledWith(expect.objectContaining({ status: undefined }))
      expect(s.rows).toHaveLength(2)
    })
  })
})
