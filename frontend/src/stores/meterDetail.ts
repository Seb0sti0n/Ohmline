import { ref } from 'vue'
import { defineStore } from 'pinia'
import {
  errorMessage,
  getAnomaly,
  getLatestAnalysis,
  getMeter,
  getMeterEvents,
  getReadings,
} from '@/services/api'
import type { AnalysisRun, Anomaly, MeterEvent, MeterSummary, ReadingPoint } from '@/types/api'

export const useMeterDetailStore = defineStore('meterDetail', () => {
  const meter = ref<MeterSummary | null>(null)
  const points = ref<ReadingPoint[]>([])
  const events = ref<MeterEvent[]>([])
  /** The meter's anomaly with its evidence, if the latest analysis found one. */
  const anomaly = ref<Anomaly | null>(null)
  const run = ref<AnalysisRun | null>(null)
  const loading = ref(false)
  const error = ref<string | null>(null)
  const notFound = ref(false)

  let seq = 0

  async function load(meterId: string) {
    const mine = ++seq
    loading.value = true
    error.value = null
    notFound.value = false
    meter.value = null
    try {
      const m = await getMeter(meterId)
      const [pts, evs, latest, detail] = await Promise.all([
        getReadings(meterId),
        getMeterEvents(meterId),
        getLatestAnalysis(),
        m.anomaly ? getAnomaly(m.anomaly.id) : Promise.resolve(null),
      ])
      if (mine !== seq) return
      meter.value = m
      points.value = pts
      events.value = evs
      run.value = latest
      anomaly.value = detail
    } catch (e) {
      if (mine !== seq) return
      if ((e as { response?: { status?: number } }).response?.status === 404) notFound.value = true
      else error.value = errorMessage(e, 'No se pudo cargar el medidor')
    } finally {
      if (mine === seq) loading.value = false
    }
  }

  return { meter, points, events, anomaly, run, loading, error, notFound, load }
})
