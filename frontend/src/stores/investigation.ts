import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import {
  errorMessage,
  getAnomalies,
  getAnomaly,
  getMeter,
  patchAnomalyStatus,
} from '@/services/api'
import type { Anomaly, AnomalyStatus, MeterSummary } from '@/types/api'

/** Everything the investigation screen shows: the anomaly with its evidence, its meter and its rank. */
export const useInvestigationStore = defineStore('investigation', () => {
  const anomaly = ref<Anomaly | null>(null)
  const meter = ref<MeterSummary | null>(null)
  const rank = ref<{ rank: number; total: number } | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const notFound = ref(false)
  const updating = ref(false)
  const updateError = ref<string | null>(null)

  const status = computed(() => anomaly.value?.status ?? null)

  let seq = 0 // only the latest load may fill the screen

  async function load(id: number) {
    const mine = ++seq
    loading.value = true
    error.value = null
    notFound.value = false
    anomaly.value = null
    try {
      const a = await getAnomaly(id)
      const [m, all] = await Promise.all([getMeter(a.meter_id), getAnomalies()])
      if (mine !== seq) return
      anomaly.value = a
      meter.value = m
      const i = all.findIndex((x) => x.id === id)
      rank.value = i < 0 ? null : { rank: i + 1, total: all.length }
    } catch (e) {
      if (mine !== seq) return
      if ((e as { response?: { status?: number } }).response?.status === 404) notFound.value = true
      else error.value = errorMessage(e, 'No se pudo cargar la investigación')
    } finally {
      if (mine === seq) loading.value = false
    }
  }

  /** Acknowledge, resolve or reopen. The screen keeps showing the previous state if it fails. */
  async function setStatus(next: AnomalyStatus) {
    if (!anomaly.value || updating.value) return
    updating.value = true
    updateError.value = null
    try {
      const updated = await patchAnomalyStatus(anomaly.value.id, next)
      anomaly.value = { ...anomaly.value, status: updated.status }
    } catch (e) {
      updateError.value = errorMessage(e, 'No se pudo actualizar el estado')
    } finally {
      updating.value = false
    }
  }

  return {
    anomaly,
    meter,
    rank,
    loading,
    error,
    notFound,
    updating,
    updateError,
    status,
    load,
    setStatus,
  }
})
