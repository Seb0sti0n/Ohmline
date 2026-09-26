import type { AnalysisRun, AnomalyType, DashboardSummary } from '@/types/api'
import { formatDecimal, formatInt, formatLocalTime, joinList, plural } from './format'

export interface Kpi {
  label: string
  value: string
  unit?: string
  sub: string
  tone?: 'critical'
  /** Smaller value text, for kpis whose value is a phrase rather than a number. */
  compact?: boolean
}

const PENDING = 'Pendiente de análisis'

const TYPE_PARTS: [AnomalyType, string, string][] = [
  ['REAL_ANOMALY', 'real', 'reales'],
  ['DATA_QUALITY', 'de datos', 'de datos'],
  ['EXPLAINABLE_ANOMALY', 'explicable', 'explicables'],
  ['FALSE_POSITIVE', 'descartada', 'descartadas'],
]

/** "1 real, 1 de datos, 1 explicable, 1 descartada" */
export function typeBreakdown(byType: DashboardSummary['anomalies_by_type']): string {
  const parts = TYPE_PARTS.filter(([t]) => (byType[t] ?? 0) > 0).map(([t, one, many]) => {
    const n = byType[t] ?? 0
    return `${formatInt(n)} ${n === 1 ? one : many}`
  })
  return parts.join(', ') || 'Sin hallazgos'
}

function lastAnalysis(run: AnalysisRun | null, now: Date): Pick<Kpi, 'value' | 'sub'> {
  if (!run) return { value: 'Sin ejecutar', sub: 'Pulsa Run AI Analysis' }
  if (run.status === 'FAILED')
    return { value: 'Falló', sub: 'Revisa el error e inténtalo de nuevo' }
  if (run.status !== 'COMPLETED') return { value: 'En curso', sub: 'Analizando lecturas' }
  const when = run.finished_at ? formatLocalTime(run.finished_at, now) : null
  const value = when ? `${when.day === 'hoy' ? 'Hoy' : when.day}, ${when.time}` : 'Completado'
  return { value, sub: 'Completado' }
}

/**
 * The six KPIs of the dashboard. Analysis-derived ones show "—" until a completed analysis exists.
 * `run` is the live run (it may be newer than what the summary knows).
 */
export function buildKpis(
  summary: DashboardSummary,
  run: AnalysisRun | null,
  now: Date = new Date(),
): Kpi[] {
  const has = summary.has_results
  const locations = new Set(summary.meters.map((m) => m.location)).size
  const days = summary.daily_consumption.length
  const high = summary.high_priority_count

  return [
    {
      label: 'Medidores',
      value: formatInt(summary.meters_count),
      sub: plural(locations, 'ubicación', 'ubicaciones'),
    },
    {
      label: 'Consumo del periodo',
      value: formatInt(summary.total_consumption_kwh),
      unit: 'kWh',
      sub: plural(days, 'día', 'días'),
    },
    {
      label: 'Anomalías IA',
      value: has ? formatInt(summary.anomalies_count) : '—',
      sub: has ? typeBreakdown(summary.anomalies_by_type) : PENDING,
    },
    {
      label: 'Alta prioridad',
      value: has ? formatInt(high) : '—',
      sub: has ? (high > 0 ? joinList(summary.high_priority_meters) : 'Ninguna') : PENDING,
      tone: has && high > 0 ? 'critical' : undefined,
    },
    {
      label: 'Confianza IA media',
      value:
        has && summary.avg_confidence !== null ? formatDecimal(summary.avg_confidence, 2) : '—',
      sub: has
        ? summary.anomalies_count > 0
          ? `Media de ${plural(summary.anomalies_count, 'hallazgo', 'hallazgos')}`
          : 'Sin hallazgos'
        : PENDING,
    },
    { label: 'Último análisis', compact: true, ...lastAnalysis(run, now) },
  ]
}
