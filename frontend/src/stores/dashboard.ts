import { ref } from 'vue'
import { defineStore } from 'pinia'
import { errorMessage, getDashboard } from '@/services/api'
import type { DashboardSummary } from '@/types/api'

export const useDashboardStore = defineStore('dashboard', () => {
  const summary = ref<DashboardSummary | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load() {
    loading.value = true
    error.value = null
    try {
      summary.value = await getDashboard()
    } catch (e) {
      error.value = errorMessage(e, 'No se pudo cargar el resumen')
    } finally {
      loading.value = false
    }
  }

  return { summary, loading, error, load }
})
