import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import type { MeterPage, MeterQuery } from '@/types/api'
import { meter, meterPage } from '@/utils/__tests__/fixtures'

const getMeters = vi.fn<(q?: MeterQuery) => Promise<MeterPage>>()
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
const lastQuery = () => getMeters.mock.lastCall![0]

beforeEach(() => {
  setActivePinia(createPinia())
  getMeters.mockReset()
  getMeters.mockResolvedValue(meterPage(everyone))
})

describe('meters store', () => {
  it('loads the first page with the counters', async () => {
    const s = useMetersStore()
    await s.load()
    expect(s.rows).toHaveLength(5)
    expect(s.total).toBe(5)
    expect(s.counts).toEqual({ ALL: 5, OK: 2, ALERT: 2, CRITICAL: 1, UNEVALUATED: 0 })
    expect(s.unevaluated).toBe(false)
    // default query: severity, descending, first page of 10
    expect(getMeters).toHaveBeenCalledWith({
      status: undefined,
      search: undefined,
      sort: 'severity',
      order: 'desc',
      page: 1,
      page_size: 10,
    })
  })

  it('sends the filter and search to the API, trimming the search', async () => {
    const s = useMetersStore()
    await s.load()
    await s.setFilter('ALERT')
    expect(lastQuery()).toMatchObject({ status: 'ALERT' })
    await s.setSearch('  11 ')
    expect(lastQuery()).toMatchObject({ status: 'ALERT', search: '11' })
    await s.setFilter('ALL')
    expect(lastQuery()).toMatchObject({ status: undefined, search: '11' })
  })

  it('keeps the counters of all meters when the table is filtered', async () => {
    const s = useMetersStore()
    await s.load()
    getMeters.mockResolvedValue(
      meterPage([meter('M-109', { status: 'CRITICAL' })], {
        counts: { ALL: 5, OK: 2, ALERT: 2, CRITICAL: 1, UNEVALUATED: 0 },
      }),
    )
    await s.setFilter('CRITICAL')
    expect(s.rows).toHaveLength(1)
    expect(s.total).toBe(1)
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
    expect(lastQuery()).toMatchObject({ sort: 'variation', order: 'asc' })
  })

  it('ignores a slow response that arrives after a newer one', async () => {
    const s = useMetersStore()
    await s.load()
    let releaseSlow!: (v: MeterPage) => void
    getMeters.mockReturnValueOnce(new Promise((r) => (releaseSlow = r)))
    const slow = s.setSearch('1') // sent first, answered last
    getMeters.mockResolvedValueOnce(meterPage([meter('M-112')]))
    await s.setSearch('112') // sent second, answered first
    releaseSlow(meterPage([meter('M-101'), meter('M-102')]))
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

  describe('pagination', () => {
    // 25 meters: 3 pages of 10
    const many = Array.from({ length: 25 }, (_, i) => meter(`M-${100 + i}`))
    const respond = (q?: MeterQuery) => {
      const size = q?.page_size ?? 10
      const start = ((q?.page ?? 1) - 1) * size
      return meterPage(many.slice(start, start + size), {
        total: 25,
        page: q?.page ?? 1,
        page_size: size,
        counts: { ALL: 25, OK: 25, ALERT: 0, CRITICAL: 0, UNEVALUATED: 0 },
      })
    }
    beforeEach(() => getMeters.mockImplementation(async (q) => respond(q)))

    it('knows how many pages there are', async () => {
      const s = useMetersStore()
      await s.load()
      expect(s.total).toBe(25)
      expect(s.pageCount).toBe(3)
      expect(s.rows).toHaveLength(10)
    })

    it('goes to another page, and never outside 1..pageCount', async () => {
      const s = useMetersStore()
      await s.load()
      await s.setPage(3)
      expect(s.page).toBe(3)
      expect(s.rows).toHaveLength(5)
      expect(lastQuery()).toMatchObject({ page: 3 })
      await s.setPage(99)
      expect(s.page).toBe(3)
      await s.setPage(0)
      expect(s.page).toBe(1)
    })

    it('does not ask the API again for the page that is already shown', async () => {
      const s = useMetersStore()
      await s.load()
      getMeters.mockClear()
      await s.setPage(1)
      expect(getMeters).not.toHaveBeenCalled()
    })

    it('starts again from page 1 when the filter, search, sort or page size change', async () => {
      const s = useMetersStore()
      await s.load()
      const actions: [string, () => Promise<unknown>][] = [
        ['filter', () => s.setFilter('OK')],
        ['search', () => s.setSearch('10')],
        ['sort', () => s.toggleSort('consumption')],
        ['page size', () => s.setPageSize(25)],
      ]
      for (const [name, act] of actions) {
        await s.setPage(2)
        if (s.page !== 2) await s.setPage(1) // (with 25 per page there is only one page)
        await act()
        expect(s.page, `after changing the ${name}`).toBe(1)
      }
    })

    it('changes the page size and recomputes the pages', async () => {
      const s = useMetersStore()
      await s.load()
      await s.setPage(3)
      await s.setPageSize(25)
      expect(s.pageSize).toBe(25)
      expect(s.page).toBe(1)
      expect(s.pageCount).toBe(1)
      expect(s.rows).toHaveLength(25)
      expect(lastQuery()).toMatchObject({ page_size: 25, page: 1 })
    })

    it('recovers when the page it asked for no longer exists', async () => {
      const s = useMetersStore()
      await s.load() // 25 meters: 3 pages
      // the data shrinks to 12 meters (2 pages) while page 3 is requested
      const twelve = many.slice(0, 12)
      getMeters.mockImplementation(async (q) => {
        const start = ((q?.page ?? 1) - 1) * 10
        return meterPage(twelve.slice(start, start + 10), { total: 12, page: q?.page ?? 1 })
      })
      await s.setPage(3)
      expect(s.page).toBe(2) // it fell back to the last page that exists...
      expect(s.rows).toHaveLength(2) // ...and shows it instead of an empty table
      expect(s.total).toBe(12)
      expect(s.pageCount).toBe(2)
    })
  })

  describe('before the first analysis', () => {
    const fresh = [
      meter('M-101', { status: 'UNEVALUATED' }),
      meter('M-102', { status: 'UNEVALUATED' }),
    ]

    it('knows that no meter has a status yet', async () => {
      getMeters.mockResolvedValue(meterPage(fresh))
      const s = useMetersStore()
      await s.load()
      expect(s.unevaluated).toBe(true)
      expect(s.counts).toMatchObject({ ALL: 2, UNEVALUATED: 2, OK: 0, ALERT: 0, CRITICAL: 0 })
    })

    it('is not "unevaluated" when there are no meters or only some lack a status', async () => {
      const s = useMetersStore()
      getMeters.mockResolvedValue(meterPage([]))
      await s.load()
      expect(s.unevaluated).toBe(false)
      getMeters.mockResolvedValue(
        meterPage([meter('M-101', { status: 'UNEVALUATED' }), meter('M-102')]),
      )
      await s.load()
      expect(s.unevaluated).toBe(false)
    })

    it('drops a leftover status filter, which would show an empty table', async () => {
      const s = useMetersStore()
      await s.load()
      await s.setFilter('CRITICAL') // chosen while an analysis existed
      getMeters.mockResolvedValue(meterPage(fresh)) // ...then the data was reset
      await s.load()
      expect(s.filter).toBe('ALL')
      expect(lastQuery()).toMatchObject({ status: undefined, page: 1 })
      expect(s.rows).toHaveLength(2)
    })
  })
})
