import { afterEach, describe, expect, it, vi } from 'vitest'
import { api, errorMessage, getLatestAnalysis, startAnalysis } from '../api'

const httpError = (status?: number, data: unknown = {}) =>
  Object.assign(new Error('x'), {
    isAxiosError: true,
    response: status === undefined ? undefined : { status, data },
  })

afterEach(() => vi.restoreAllMocks())

describe('errorMessage', () => {
  it('says the server is not responding when there was no answer', () => {
    expect(errorMessage(httpError(), 'No se pudo cargar el resumen')).toBe(
      'No se pudo cargar el resumen. El servidor no responde.',
    )
  })

  it('treats any 5xx the same, and never shows the server-side English text', () => {
    for (const status of [500, 502, 503]) {
      const msg = errorMessage(
        httpError(status, { error: 'internal error' }),
        'No se pudo cargar el medidor',
      )
      expect(msg).toBe('No se pudo cargar el medidor. El servidor no responde.')
      expect(msg).not.toContain('internal error')
    }
  })

  it('uses the caller message for client errors and for unknown errors', () => {
    expect(errorMessage(httpError(400, { error: 'invalid id' }), 'No se pudo abrir')).toBe(
      'No se pudo abrir',
    )
    expect(errorMessage(new Error('boom'), 'Algo falló')).toBe('Algo falló')
    expect(errorMessage(undefined)).toBe('No se pudo completar la solicitud')
  })
})

describe('startAnalysis', () => {
  it('returns the id of the new run', async () => {
    vi.spyOn(api, 'post').mockResolvedValue({ data: { id: 12 } })
    expect(await startAnalysis()).toBe(12)
  })

  it('follows the run already in progress when the API answers 409', async () => {
    vi.spyOn(api, 'post').mockRejectedValue(
      httpError(409, { error: 'an analysis is already running', id: 7 }),
    )
    expect(await startAnalysis()).toBe(7)
  })

  it('lets other failures through', async () => {
    vi.spyOn(api, 'post').mockRejectedValue(httpError(500))
    await expect(startAnalysis()).rejects.toBeTruthy()
  })
})

describe('getLatestAnalysis', () => {
  it('returns null (not an error) before the first analysis', async () => {
    vi.spyOn(api, 'get').mockRejectedValue(
      httpError(404, { error: 'no analysis has been run yet' }),
    )
    expect(await getLatestAnalysis()).toBeNull()
  })

  it('returns the run, and lets real failures through', async () => {
    vi.spyOn(api, 'get').mockResolvedValueOnce({ data: { id: 3, status: 'COMPLETED' } })
    expect((await getLatestAnalysis())?.id).toBe(3)
    vi.spyOn(api, 'get').mockRejectedValueOnce(httpError(500))
    await expect(getLatestAnalysis()).rejects.toBeTruthy()
  })
})
