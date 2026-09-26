import type { AnalysisRun, DashboardSummary, MeterSummary, RunStep } from '@/types/api'

export const NAMES = [
  'Lecturas',
  'Baseline',
  'Detección',
  'Correlación',
  'Eventos',
  'Explicación',
  'Recomendación',
]

export const steps = (statuses: RunStep['status'][], ms = 100): RunStep[] =>
  NAMES.map((name, i) => ({
    name,
    status: statuses[i] ?? 'PENDING',
    duration_ms: statuses[i] === 'DONE' ? ms : 0,
  }))

export const done = Array(7).fill('DONE') as RunStep['status'][]

export function run(over: Partial<AnalysisRun> = {}): AnalysisRun {
  return {
    id: 1,
    status: 'COMPLETED',
    current_step: null,
    steps: steps(done),
    summary: {
      meters_analyzed: 12,
      readings_count: 4032,
      events_count: 4,
      anomalies: 4,
      high_priority: 2,
      avg_confidence: 0.9,
      duration_ms: 4600,
      llm_enabled: false,
      llm_explanations: 0,
    },
    started_at: '2026-09-25T14:00:00Z',
    finished_at: '2026-09-25T14:00:05Z',
    ...over,
  }
}

export function meter(id: string, over: Partial<MeterSummary> = {}): MeterSummary {
  return {
    meter_id: id,
    name: `Medidor ${id}`,
    location: 'Planta Norte',
    status: 'OK',
    consumption_kwh: 700,
    baseline_kwh: 700,
    variation_pct: 0,
    daily_kwh: Array(14).fill(700),
    anomaly: null,
    ...over,
  }
}

export function summary(over: Partial<DashboardSummary> = {}): DashboardSummary {
  return {
    meters_count: 12,
    total_consumption_kwh: 155251,
    has_results: true,
    anomalies_count: 4,
    high_priority_count: 2,
    high_priority_meters: ['M-109', 'M-112'],
    anomalies_by_type: {
      REAL_ANOMALY: 1,
      DATA_QUALITY: 1,
      EXPLAINABLE_ANOMALY: 1,
      FALSE_POSITIVE: 1,
    },
    avg_confidence: 0.9,
    last_analysis: run(),
    daily_consumption: Array.from({ length: 14 }, (_, i) => ({
      date: `2026-09-${String(i + 1).padStart(2, '0')}`,
      consumption_kwh: 10800,
      baseline_kwh: 10758,
    })),
    top_priorities: [],
    meters: [meter('M-101'), meter('M-102', { location: 'Planta Sur' })],
    generated_at: '2026-09-25T14:00:00Z',
    ...over,
  }
}
