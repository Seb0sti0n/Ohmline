import type { AnalysisRun, AnomalyType, DashboardSummary } from '@/types/api'
import { formatInt, formatLocalTime, formatSeconds, plural } from './format'

export const STEP_NAMES = [
  'Lecturas',
  'Baseline',
  'Detección',
  'Correlación',
  'Eventos',
  'Explicación',
  'Recomendación',
]

export interface StepView {
  label: string
  number: number
  state: 'done' | 'active' | 'pending'
  detail: string
}

type ByType = Partial<Record<AnomalyType, number>>

/** What each stage found, once the run has finished and its summary is known. */
function finishedDetail(index: number, run: AnalysisRun, byType: ByType): string {
  const s = run.summary
  if (!s) return ''
  const irregular = byType.DATA_QUALITY ?? 0
  const changes =
    (byType.REAL_ANOMALY ?? 0) + (byType.EXPLAINABLE_ANOMALY ?? 0) + (byType.FALSE_POSITIVE ?? 0)
  switch (index) {
    case 0:
      return `${formatInt(s.readings_count)} lecturas`
    case 1:
      return `${formatInt(s.meters_analyzed * 24)} perfiles horarios`
    case 2: {
      const parts = []
      if (changes > 0) parts.push(plural(changes, 'cambio', 'cambios'))
      if (irregular > 0) parts.push(plural(irregular, 'serie irregular', 'series irregulares'))
      return parts.join(', ') || 'Sin hallazgos'
    }
    case 3:
      return 'Corriente, FP, voltaje'
    case 4:
      return `${plural(s.events_count, 'evento cruzado', 'eventos cruzados')}`
    case 5:
      return plural(s.anomalies, 'explicación', 'explicaciones')
    default:
      return plural(s.anomalies, 'acción', 'acciones')
  }
}

/** The seven stepper steps for the current run (or all pending if there is none). */
export function buildSteps(run: AnalysisRun | null, byType: ByType = {}): StepView[] {
  return STEP_NAMES.map((label, i) => {
    const step = run?.steps?.[i]
    let state: StepView['state'] = 'pending'
    if (run?.status === 'COMPLETED') state = 'done'
    else if (step?.status === 'DONE') state = 'done'
    else if (step?.status === 'RUNNING') state = 'active'

    let detail = ''
    if (state === 'active') detail = 'En curso'
    else if (state === 'done' && run) {
      detail =
        run.status === 'COMPLETED'
          ? finishedDetail(i, run, byType)
          : formatSeconds(step?.duration_ms ?? 0)
    }
    return { label, number: i + 1, state, detail }
  })
}

/** "Paso 3 de 7", "Completado hoy a las 09:42 en 4,6 s", "Sin ejecutar"... */
export function statusText(run: AnalysisRun | null, now: Date = new Date()): string {
  if (!run) return 'Sin ejecutar'
  if (run.status === 'FAILED') return 'El análisis falló'
  if (run.status !== 'COMPLETED') {
    const active = run.steps.findIndex((s) => s.status === 'RUNNING')
    const done = run.steps.filter((s) => s.status === 'DONE').length
    return `Paso ${(active >= 0 ? active : done) + 1} de ${STEP_NAMES.length}`
  }
  const when = run.finished_at ? formatLocalTime(run.finished_at, now) : null
  const duration = run.summary ? ` en ${formatSeconds(run.summary.duration_ms)}` : ''
  if (!when) return `Completado${duration}`
  return `Completado ${when.day === 'hoy' ? 'hoy' : `el ${when.day}`} a las ${when.time}${duration}`
}

/** The banner shown when a run has finished: "4 anomalías detectadas, 2 requieren atención prioritaria". */
export function resultText(run: AnalysisRun | null): string | null {
  const s = run?.status === 'COMPLETED' ? run.summary : null
  if (!s) return null
  if (s.anomalies === 0) return 'No se detectaron anomalías'
  const found = plural(s.anomalies, 'anomalía detectada', 'anomalías detectadas')
  if (s.high_priority === 0) return `${found}, ninguna requiere atención prioritaria`
  return `${found}, ${s.high_priority} ${s.high_priority === 1 ? 'requiere' : 'requieren'} atención prioritaria`
}

export type { DashboardSummary }
