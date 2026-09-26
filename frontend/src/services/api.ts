import axios, { AxiosError } from 'axios'
import { clearSession, getToken } from './session'
import type { AnalysisRun, DashboardSummary, MeterQuery, MeterSummary } from '@/types/api'

export const api = axios.create({ baseURL: '/api' })

api.interceptors.request.use((config) => {
  const token = getToken()
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

// An expired or invalid token sends the user back to the login screen (set by main.ts).
let onUnauthorized: () => void = () => {}
export const setUnauthorizedHandler = (fn: () => void) => (onUnauthorized = fn)

api.interceptors.response.use(
  (r) => r,
  (error: AxiosError) => {
    const isLogin = error.config?.url?.includes('/auth/login')
    if (error.response?.status === 401 && !isLogin) {
      clearSession()
      onUnauthorized()
    }
    return Promise.reject(error)
  },
)

/** A readable message for any failed request. */
export function errorMessage(
  error: unknown,
  fallback = 'No se pudo completar la solicitud',
): string {
  if (axios.isAxiosError(error)) {
    if (!error.response) return 'No se pudo conectar con el servidor'
    const msg = (error.response.data as { error?: string } | undefined)?.error
    if (msg) return msg
  }
  return fallback
}

export const login = (email: string, password: string) =>
  api
    .post<{ token: string; user: { email: string } }>('/auth/login', { email, password })
    .then((r) => r.data)

export const getDashboard = () =>
  api.get<DashboardSummary>('/dashboard/summary').then((r) => r.data)

export const getMeters = (params: MeterQuery = {}) =>
  api.get<MeterSummary[]>('/meters', { params }).then((r) => r.data)

/** Starts an analysis. If one is already running the API answers 409 with its id, which we follow. */
export async function startAnalysis(): Promise<number> {
  try {
    return (await api.post<{ id: number }>('/ai/analyze')).data.id
  } catch (e) {
    if (axios.isAxiosError(e) && e.response?.status === 409)
      return (e.response.data as { id: number }).id
    throw e
  }
}

export const getAnalysis = (id: number) =>
  api.get<AnalysisRun>(`/ai/analysis/${id}`).then((r) => r.data)

/** The latest run, or null if none has been executed yet. */
export async function getLatestAnalysis(): Promise<AnalysisRun | null> {
  try {
    return (await api.get<AnalysisRun>('/ai/analysis/latest')).data
  } catch (e) {
    if (axios.isAxiosError(e) && e.response?.status === 404) return null
    throw e
  }
}
