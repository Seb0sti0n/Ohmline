import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { errorMessage, getAnalysis, getLatestAnalysis, startAnalysis } from '@/services/api'
import { useDashboardStore } from './dashboard'
import { useMetersStore } from './meters'
import type { AnalysisRun } from '@/types/api'

export const POLL_MS = 500

const sleep = (ms: number) => new Promise((resolve) => setTimeout(resolve, ms))

export const useAnalysisStore = defineStore('analysis', () => {
  /** The latest run (running or finished), or null if none has been executed. */
  const run = ref<AnalysisRun | null>(null)
  const error = ref<string | null>(null)
  const starting = ref(false)

  const inProgress = (r: AnalysisRun | null) => r?.status === 'RUNNING' || r?.status === 'PENDING'
  const isRunning = computed(() => starting.value || inProgress(run.value))

  // Incremented to cancel the polling loop of a previous run (logout, new run, unmount).
  let generation = 0

  function stop() {
    generation++
    starting.value = false
  }

  /** Polls a run until it finishes, updating `run` so the stepper follows it. */
  async function follow(id: number) {
    const mine = ++generation
    while (mine === generation) {
      let latest: AnalysisRun
      try {
        latest = await getAnalysis(id)
      } catch (e) {
        if (mine === generation) {
          error.value = errorMessage(e, 'No se pudo consultar el análisis')
          starting.value = false
        }
        return
      }
      if (mine !== generation) return
      run.value = latest
      starting.value = false
      if (!inProgress(latest)) {
        if (latest.status === 'COMPLETED') {
          // New results: refresh everything that depends on them.
          await Promise.all([useDashboardStore().load(), useMetersStore().load()])
        } else {
          error.value = latest.summary?.error
            ? `El análisis falló: ${latest.summary.error}`
            : 'El análisis falló'
        }
        return
      }
      await sleep(POLL_MS)
    }
  }

  /** Reads the latest run; if one is still running (e.g. after a page reload) it keeps following it. */
  async function loadLatest() {
    try {
      run.value = await getLatestAnalysis()
    } catch (e) {
      error.value = errorMessage(e, 'No se pudo consultar el último análisis')
      return
    }
    if (run.value && inProgress(run.value)) void follow(run.value.id)
  }

  async function start() {
    if (isRunning.value) return
    error.value = null
    starting.value = true
    try {
      const id = await startAnalysis()
      await follow(id)
    } catch (e) {
      error.value = errorMessage(e, 'No se pudo iniciar el análisis')
      starting.value = false
    }
  }

  return { run, error, starting, isRunning, loadLatest, start, stop }
})
