// Shapes returned by the backend (see backend/openapi.yaml).

export type MeterStatus = 'OK' | 'ALERT' | 'CRITICAL'
export type Severity = 'HIGH' | 'MEDIUM' | 'LOW'
export type AnomalyType = 'REAL_ANOMALY' | 'EXPLAINABLE_ANOMALY' | 'FALSE_POSITIVE' | 'DATA_QUALITY'
export type AnomalyStatus = 'OPEN' | 'ACKNOWLEDGED' | 'RESOLVED'

export interface Anomaly {
  id: number
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
  evidence?: Record<string, unknown>
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
}
