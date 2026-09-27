// Shapes returned by the backend (see backend/openapi.yaml).

/** UNEVALUATED until the first analysis completes: the API does not claim a meter is fine before that. */
export type MeterStatus = 'OK' | 'ALERT' | 'CRITICAL' | 'UNEVALUATED'
export type Severity = 'HIGH' | 'MEDIUM' | 'LOW'
export type AnomalyType = 'REAL_ANOMALY' | 'EXPLAINABLE_ANOMALY' | 'FALSE_POSITIVE' | 'DATA_QUALITY'
export type AnomalyStatus = 'OPEN' | 'ACKNOWLEDGED' | 'RESOLVED'

export interface Anomaly {
  id: number
  /** Always true: every row is a finding (the output format of the brief). */
  anomaly: boolean
  meter_id: string
  meter_name: string
  analysis_run_id: number
  detected_at: string
  type: AnomalyType
  severity: Severity
  confidence: number
  priority_score: number
  reason: string
  explanation: string
  recommended_action: string
  status: AnomalyStatus
  window_start: string | null
  window_end: string | null
  explanation_source: 'LLM' | 'TEMPLATE'
  evidence?: Evidence
}

export interface Change {
  variable: 'consumption_kwh' | 'current_a' | 'power_factor' | 'voltage_v' | 'voltage_std'
  baseline: number
  observed: number
  delta: number
  delta_pct: number
  significant: boolean
  direction: 'up' | 'down' | 'stable'
}

export interface EventMatch {
  type: EventType
  timestamp: string
  description: string
  compatible: boolean
  quality: number
  reason: string
}

export type EventType = 'SCHEDULED_OUTAGE' | 'OPERATIONAL_CHANGE' | 'DATA_QUALITY' | 'UNKNOWN'

export interface QualityInfo {
  total_flags: number
  by_kind: Record<string, number>
  recurrent: boolean
  first_flag: string
  last_flag: string
  max_flags_in_24h: number
  flagged_hours: number
  window_hours: number
}

/** What backs an anomaly's conclusion (see engine.Evidence). */
export interface Evidence {
  kind: 'consumption_shift' | 'data_quality'
  window: { start: string; end: string; hours: number; ongoing: boolean }
  metrics: {
    baseline_daily_kwh: number
    last_day_kwh: number
    last_day_variation_pct: number
    window_variation_pct: number
    extra_kwh: number
    median_abs_z: number
  }
  changed_variables: Change[]
  events: EventMatch[]
  rules_fired: string[]
  quality?: QualityInfo
  priority_breakdown: Record<string, number>
  confidence_breakdown: Record<string, number>
  priority_weights: Record<string, number>
  confidence_weights: Record<string, number>
}

export interface MeterSummary {
  meter_id: string
  name: string
  location: string
  status: MeterStatus
  consumption_kwh: number
  baseline_kwh: number
  variation_pct: number
  daily_kwh: number[]
  daily_from: string
  /** Short runs outside the expected band: counted from the data, normal noise, never an anomaly. */
  isolated_spikes: number
  anomaly: Anomaly | null
}

export type StepStatus = 'PENDING' | 'RUNNING' | 'DONE'

export interface RunStep {
  name: string
  status: StepStatus
  duration_ms: number
}

export interface RunSummary {
  meters_analyzed: number
  readings_count: number
  events_count: number
  anomalies: number
  high_priority: number
  avg_confidence: number
  duration_ms: number
  llm_enabled: boolean
  llm_explanations: number
  error?: string
}

export interface AnalysisRun {
  id: number
  status: 'PENDING' | 'RUNNING' | 'COMPLETED' | 'FAILED'
  current_step: string | null
  steps: RunStep[]
  summary: RunSummary | null
  started_at: string
  finished_at: string | null
}

export interface DailyPoint {
  date: string
  consumption_kwh: number
  baseline_kwh: number
}

export interface DashboardSummary {
  meters_count: number
  total_consumption_kwh: number
  has_results: boolean
  anomalies_count: number
  high_priority_count: number
  high_priority_meters: string[]
  anomalies_by_type: Partial<Record<AnomalyType, number>>
  avg_confidence: number | null
  last_analysis: AnalysisRun | null
  daily_consumption: DailyPoint[]
  top_priorities: Anomaly[]
  meters: MeterSummary[]
  generated_at: string
}

export type MeterFilter = 'ALL' | MeterStatus
export type MeterSort = 'severity' | 'consumption' | 'variation'
export type SortOrder = 'asc' | 'desc'

export interface MeterQuery {
  status?: MeterStatus
  search?: string
  sort?: MeterSort
  order?: SortOrder
  page?: number
  page_size?: number
}

/** Number of meters per status over all meters (ignoring filters), for the filter chips. */
export type MeterCounts = Record<'ALL' | MeterStatus, number>

/** One page of the meters table. `total` is what matches the filters, `counts` covers every meter. */
export interface MeterPage {
  items: MeterSummary[]
  total: number
  page: number
  page_size: number
  counts: MeterCounts
}

/** One point of a meter series with the expected values for the same hour (or day). */
export interface ReadingPoint {
  timestamp: string
  consumption_kwh: number
  voltage_v: number
  current_a: number
  power_factor: number
  baseline_kwh: number
  band_low_kwh: number
  band_high_kwh: number
  baseline_voltage_v: number
  baseline_current_a: number
  baseline_power_factor: number
}

export interface MeterEvent {
  timestamp: string
  type: EventType
  description: string
}

export interface AnomalyQuery {
  type?: AnomalyType
  severity?: Severity
  status?: AnomalyStatus
}
