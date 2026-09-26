import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest'
import { createPinia, setActivePinia } from 'pinia'
import type { AnalysisRun } from '@/types/api'
import { run, steps } from '@/utils/__tests__/fixtures'

const api = {
  startAnalysis: vi.fn<() => Promise<number>>(),
  getAnalysis: vi.fn<(id: number) => Promise<AnalysisRun>>(),
  getLatestAnalysis: vi.fn<() => Promise<AnalysisRun | null>>(),
  getDashboard: vi.fn(),
  getMeters: vi.fn(),
}
vi.mock('@/services/api', async (orig) => ({
  ...(await orig<typeof import('@/services/api')>()),
  startAnalysis: () => api.startAnalysis(),
  getAnalysis: (id: number) => api.getAnalysis(id),
  getLatestAnalysis: () => api.getLatestAnalysis(),
  getDashboard: () => api.getDashboard(),
  getMeters: () => api.getMeters(),
}))

import { POLL_MS, useAnalysisStore } from '../analysis'
import { useDashboardStore } from '../dashboard'

const running = (n: number) =>
  run({
    status: 'RUNNING',
    summary: null,
    finished_at: null,
    steps: steps(Array(n).fill('DONE').concat('RUNNING')),
  })

beforeEach(() => {
  vi.useFakeTimers()
  setActivePinia(createPinia())
  Object.values(api).forEach((f) => f.mockReset())
  api.getDashboard.mockResolvedValue({ meters: [] })
  api.getMeters.mockResolvedValue([])
})
afterEach(() => vi.useRealTimers())

describe('analysis store', () => {
  it('starts a run, follows it until it completes and refreshes the dashboard', async () => {
    const s = useAnalysisStore()
    api.startAnalysis.mockResolvedValue(7)
    api.getAnalysis
      .mockResolvedValueOnce(running(1))
      .mockResolvedValueOnce(running(4))
      .mockResolvedValueOnce(run({ id: 7 }))

    const p = s.start()
    expect(s.isRunning).toBe(true)
    await vi.advanceTimersByTimeAsync(POLL_MS * 3)
    await p

    expect(api.getAnalysis).toHaveBeenCalledTimes(3)
    expect(api.getAnalysis).toHaveBeenCalledWith(7)
    expect(s.run?.status).toBe('COMPLETED')
    expect(s.isRunning).toBe(false)
    expect(s.error).toBeNull()
    expect(api.getDashboard).toHaveBeenCalledTimes(1) // results changed: dashboard and table reload
    expect(api.getMeters).toHaveBeenCalled()
  })

  it('exposes the progress while running', async () => {
    const s = useAnalysisStore()
    api.startAnalysis.mockResolvedValue(1)
    api.getAnalysis.mockResolvedValueOnce(running(2)).mockResolvedValue(run())
    const p = s.start()
    await vi.advanceTimersByTimeAsync(0)
    expect(s.run?.steps.filter((x) => x.status === 'DONE')).toHaveLength(2)
    expect(s.isRunning).toBe(true)
    await vi.advanceTimersByTimeAsync(POLL_MS)
    await p
  })

  it('does not start a second run while one is in progress', async () => {
    const s = useAnalysisStore()
    api.startAnalysis.mockResolvedValue(1)
    api.getAnalysis.mockResolvedValueOnce(running(0)).mockResolvedValue(run())
    const first = s.start()
    await s.start() // ignored
    expect(api.startAnalysis).toHaveBeenCalledTimes(1)
    await vi.advanceTimersByTimeAsync(POLL_MS)
    await first
  })

  it('reports a failed run', async () => {
    const s = useAnalysisStore()
    api.startAnalysis.mockResolvedValue(3)
    api.getAnalysis.mockResolvedValue(
      run({ status: 'FAILED', summary: { ...run().summary!, error: 'database down' } }),
    )
    await s.start()
    expect(s.error).toBe('El análisis falló: database down')
    expect(s.isRunning).toBe(false)
    expect(api.getDashboard).not.toHaveBeenCalled()
  })

  it('reports an error if starting fails, and allows trying again', async () => {
    const s = useAnalysisStore()
    api.startAnalysis.mockRejectedValueOnce(new Error('network'))
    await s.start()
    expect(s.error).toBeTruthy()
    expect(s.isRunning).toBe(false)
    api.startAnalysis.mockResolvedValue(2)
    api.getAnalysis.mockResolvedValue(run({ id: 2 }))
    await s.start()
    expect(s.error).toBeNull()
    expect(s.run?.id).toBe(2)
  })

  it('reports an error if polling fails', async () => {
    const s = useAnalysisStore()
    api.startAnalysis.mockResolvedValue(1)
    api.getAnalysis.mockRejectedValue(new Error('lost connection'))
    await s.start()
    expect(s.error).toBeTruthy()
    expect(s.isRunning).toBe(false)
  })

  it('stop() cancels the polling loop', async () => {
    const s = useAnalysisStore()
    api.startAnalysis.mockResolvedValue(1)
    api.getAnalysis.mockResolvedValue(running(1))
    const p = s.start()
    await vi.advanceTimersByTimeAsync(POLL_MS * 2)
    const calls = api.getAnalysis.mock.calls.length
    s.stop()
    await vi.advanceTimersByTimeAsync(POLL_MS * 5)
    await p
    expect(api.getAnalysis.mock.calls.length).toBe(calls) // no more requests after stop()
  })

  it('loadLatest shows the last run, or nothing if there is none', async () => {
    const s = useAnalysisStore()
    api.getLatestAnalysis.mockResolvedValueOnce(null)
    await s.loadLatest()
    expect(s.run).toBeNull()
    api.getLatestAnalysis.mockResolvedValueOnce(run({ id: 5 }))
    await s.loadLatest()
    expect(s.run?.id).toBe(5)
    expect(api.getAnalysis).not.toHaveBeenCalled() // finished: nothing to follow
  })

  it('loadLatest resumes following a run that is still in progress (page reloaded mid-run)', async () => {
    const s = useAnalysisStore()
    api.getLatestAnalysis.mockResolvedValue(running(3))
    api.getAnalysis.mockResolvedValueOnce(running(5)).mockResolvedValue(run())
    await s.loadLatest()
    await vi.advanceTimersByTimeAsync(POLL_MS * 2)
    expect(s.run?.status).toBe('COMPLETED')
    expect(useDashboardStore().summary).not.toBeUndefined()
  })
})
