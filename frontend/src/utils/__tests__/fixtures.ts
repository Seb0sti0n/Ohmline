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
    daily_from: '2026-09-01',
    isolated_spikes: 0,
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

import type { Anomaly, Evidence, ReadingPoint } from '@/types/api'

export function evidence(over: Partial<Evidence> = {}): Evidence {
  return {
    kind: 'consumption_shift',
    window: {
      start: '2026-09-12T14:00:00Z',
      end: '2026-09-14T23:00:00Z',
      hours: 58,
      ongoing: true,
    },
    metrics: {
      baseline_daily_kwh: 1047.7,
      last_day_kwh: 2207.6,
      last_day_variation_pct: 110.7,
      window_variation_pct: 110.5,
      extra_kwh: 2825,
      median_abs_z: 35,
    },
    changed_variables: [
      {
        variable: 'consumption_kwh',
        baseline: 44.066,
        observed: 92.772,
        delta: 48.7,
        delta_pct: 110.5,
        significant: true,
        direction: 'up',
      },
      {
        variable: 'current_a',
        baseline: 202.058,
        observed: 424.697,
        delta: 222.6,
        delta_pct: 110.2,
        significant: true,
        direction: 'up',
      },
      {
        variable: 'power_factor',
        baseline: 0.94,
        observed: 0.74,
        delta: -0.2,
        delta_pct: -21.3,
        significant: true,
        direction: 'down',
      },
      {
        variable: 'voltage_v',
        baseline: 219.832,
        observed: 216.912,
        delta: -2.92,
        delta_pct: -1.3,
        significant: false,
        direction: 'stable',
      },
      {
        variable: 'voltage_std',
        baseline: 1.228,
        observed: 1.34,
        delta: 0.11,
        delta_pct: 9.1,
        significant: false,
        direction: 'stable',
      },
    ],
    events: [],
    rules_fired: [
      'Cambio persistente: 58 h consecutivas con |z| > 4 (z mediano 35,0)',
      'Sin evento compatible que explique el cambio',
    ],
    priority_breakdown: {
      magnitude: 25,
      persistence: 15,
      electrical: 20,
      unexplained: 20,
      extra_energy: 20,
    },
    confidence_breakdown: {
      base: 0.4,
      signal_strength: 0.25,
      independent_signals: 0.2,
      explanation_clarity: 0.09,
    },
    priority_weights: {
      magnitude: 25,
      persistence: 15,
      electrical: 20,
      unexplained: 20,
      extra_energy: 20,
    },
    confidence_weights: {
      base: 0.4,
      signal_strength: 0.25,
      independent_signals: 0.2,
      explanation_clarity: 0.1,
    },
    ...over,
  }
}

export function anomaly(over: Partial<Anomaly> = {}): Anomaly {
  return {
    id: 1,
    anomaly: true,
    meter_id: 'M-109',
    meter_name: 'Tablero principal B',
    analysis_run_id: 1,
    detected_at: '2026-09-25T14:00:00Z',
    type: 'REAL_ANOMALY',
    severity: 'HIGH',
    confidence: 0.94,
    priority_score: 100,
    reason: 'Consumo +110,5% sin evento.',
    explanation: 'Explicación.',
    recommended_action: 'Investigar medidor e instalación.',
    status: 'OPEN',
    window_start: '2026-09-12T14:00:00Z',
    window_end: '2026-09-14T23:00:00Z',
    explanation_source: 'TEMPLATE',
    evidence: evidence(),
    ...over,
  }
}

/** Hourly points: `n` hours from 1 Sep 00:00, values from f(i). */
export function points(
  n: number,
  f: (i: number) => Partial<ReadingPoint> = () => ({}),
): ReadingPoint[] {
  return Array.from({ length: n }, (_, i) => ({
    timestamp: new Date(Date.UTC(2026, 8, 1, i)).toISOString().replace('.000Z', 'Z'),
    consumption_kwh: 40,
    voltage_v: 220,
    current_a: 200,
    power_factor: 0.94,
    baseline_kwh: 40,
    band_low_kwh: 35,
    band_high_kwh: 45,
    baseline_voltage_v: 220,
    baseline_current_a: 200,
    baseline_power_factor: 0.94,
    ...f(i),
  }))
}

import type { MeterPage } from '@/types/api'

/** A page of the meters list. Counts are computed from the items unless given (they cover all meters). */
export function meterPage(items: MeterSummary[], over: Partial<MeterPage> = {}): MeterPage {
  const counts = { ALL: items.length, OK: 0, ALERT: 0, CRITICAL: 0, UNEVALUATED: 0 }
  for (const m of items) counts[m.status]++
  return { items, total: items.length, page: 1, page_size: 10, counts, ...over }
}
