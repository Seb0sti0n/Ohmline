import axios from 'axios'

export const api = axios.create({ baseURL: '/api' })

export interface Health {
  status: string
  db: string
  time: string
}

export const getHealth = () => api.get<Health>('/health').then((r) => r.data)
