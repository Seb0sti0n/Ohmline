import { beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import type { Anomaly, AnomalyStatus, MeterSummary } from '@/types/api'
import { anomaly, meter } from '@/utils/__tests__/fixtures'

const api = {
  getAnomaly: vi.fn<(id: number) => Promise<Anomaly>>(),
  getAnomalies: vi.fn<() => Promise<Anomaly[]>>(),
  getMeter: vi.fn<(id: string) => Promise<MeterSummary>>(),
  patchAnomalyStatus: vi.fn<(id: number, s: AnomalyStatus) => Promise<Anomaly>>(),
}
vi.mock('@/services/api', async (orig) => ({
  ...(await orig<typeof import('@/services/api')>()),
  getAnomaly: (id: number) => api.getAnomaly(id),
  getAnomalies: () => api.getAnomalies(),
  getMeter: (id: string) => api.getMeter(id),
  patchAnomalyStatus: (id: number, s: AnomalyStatus) => api.patchAnomalyStatus(id, s),
}))

import { useInvestigationStore } from '../investigation'

const list = [
  anomaly({ id: 5, meter_id: 'M-109' }),
  anomaly({ id: 6, meter_id: 'M-112', type: 'DATA_QUALITY' }),
  anomaly({ id: 7, meter_id: 'M-104' }),
]
const httpError = (status: number) =>
  Object.assign(new Error('x'), { isAxiosError: true, response: { status, data: {} } })

beforeEach(() => {
  setActivePinia(createPinia())
  Object.values(api).forEach((f) => f.mockReset())
  api.getAnomalies.mockResolvedValue(list)
  api.getAnomaly.mockImplementation(async (id) => list.find((a) => a.id === id)!)
  api.getMeter.mockImplementation(async (id) => meter(id))
})

describe('investigation store', () => {
  it('loads the anomaly, its meter and its rank by priority', async () => {
    const s = useInvestigationStore()
    await s.load(6)
    expect(s.anomaly?.id).toBe(6)
    expect(s.meter?.meter_id).toBe('M-112')
    expect(s.rank).toEqual({ rank: 2, total: 3 })
    expect(s.loading).toBe(false)
    expect(api.getMeter).toHaveBeenCalledWith('M-112')
  })

  it('reports a missing anomaly as not found (not as an error)', async () => {
    api.getAnomaly.mockRejectedValue(httpError(404))
    const s = useInvestigationStore()
    await s.load(99)
    expect(s.notFound).toBe(true)
    expect(s.error).toBeNull()
    expect(s.anomaly).toBeNull()
  })

  it('reports other failures as errors, and recovers on retry', async () => {
    api.getAnomaly.mockRejectedValueOnce(httpError(500))
    const s = useInvestigationStore()
    await s.load(5)
    expect(s.error).toBeTruthy()
    expect(s.notFound).toBe(false)
    await s.load(5)
    expect(s.error).toBeNull()
    expect(s.anomaly?.id).toBe(5)
  })

  it('shows only the latest load when navigating quickly between anomalies', async () => {
    const s = useInvestigationStore()
    let releaseFirst!: (a: Anomaly) => void
    api.getAnomaly.mockReturnValueOnce(new Promise((r) => (releaseFirst = r)))
    const first = s.load(5) // slow
    const second = s.load(7) // fast
    await second
    releaseFirst(list[0])
    await first
    expect(s.anomaly?.id).toBe(7)
    expect(s.loading).toBe(false)
  })

  describe('status changes', () => {
    beforeEach(async () => {
      await useInvestigationStore().load(5)
    })

    it('acknowledges, resolves and reopens through the API', async () => {
      const s = useInvestigationStore()
      for (const next of ['ACKNOWLEDGED', 'RESOLVED', 'OPEN'] as const) {
        api.patchAnomalyStatus.mockResolvedValueOnce({ ...list[0], status: next })
        await s.setStatus(next)
        expect(api.patchAnomalyStatus).toHaveBeenLastCalledWith(5, next)
        expect(s.status).toBe(next)
      }
      expect(s.updateError).toBeNull()
    })

    it('keeps the previous status and shows an error if the update fails', async () => {
      const s = useInvestigationStore()
      api.patchAnomalyStatus.mockRejectedValueOnce(httpError(500))
      await s.setStatus('RESOLVED')
      expect(s.status).toBe('OPEN')
      expect(s.updateError).toBeTruthy()
      api.patchAnomalyStatus.mockResolvedValueOnce({ ...list[0], status: 'RESOLVED' })
      await s.setStatus('RESOLVED')
      expect(s.status).toBe('RESOLVED')
      expect(s.updateError).toBeNull()
    })

    it('ignores a second click while an update is in flight', async () => {
      const s = useInvestigationStore()
      let finish!: (a: Anomaly) => void
      api.patchAnomalyStatus.mockReturnValueOnce(new Promise((r) => (finish = r)))
      const first = s.setStatus('ACKNOWLEDGED')
      await s.setStatus('RESOLVED') // ignored
      expect(api.patchAnomalyStatus).toHaveBeenCalledTimes(1)
      expect(s.updating).toBe(true)
      finish({ ...list[0], status: 'ACKNOWLEDGED' })
      await first
      expect(s.updating).toBe(false)
    })

    it('keeps the evidence when the status changes', async () => {
      const s = useInvestigationStore()
      api.patchAnomalyStatus.mockResolvedValueOnce({
        ...list[0],
        status: 'RESOLVED',
        evidence: undefined,
      })
      await s.setStatus('RESOLVED')
      expect(s.anomaly?.evidence).toBeDefined()
    })
  })
})
