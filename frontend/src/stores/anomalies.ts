import { ref } from 'vue'
import { defineStore } from 'pinia'
import { errorMessage, getAnomalies } from '@/services/api'
import type { Anomaly } from '@/types/api'

export const useAnomaliesStore = defineStore('anomalies', () => {
  /** Anomalies of the latest completed analysis, highest priority first. */
  const list = ref<Anomaly[]>([])
  const loading = ref(false)
  const error = ref<string | null>(null)

  async function load() {
    loading.value = true
    error.value = null
    try {
      list.value = await getAnomalies()
    } catch (e) {
      error.value = errorMessage(e, 'No se pudieron cargar las anomalías')
    } finally {
      loading.value = false
    }
  }

  return { list, loading, error, load }
})
