import type { AnomalyType, MeterStatus, Severity } from '@/types/api'

// Full class names (not built from pieces) so Tailwind can see them.

export const STATUS_LABEL: Record<MeterStatus, string> = {
  OK: 'Normal',
  ALERT: 'Alerta',
  CRITICAL: 'Crítico',
}

export const STATUS_TEXT: Record<MeterStatus, string> = {
  OK: 'text-status-ok',
  ALERT: 'text-status-alert',
  CRITICAL: 'text-status-critical',
}

export const STATUS_BADGE: Record<MeterStatus, string> = {
  OK: 'bg-status-ok-bg text-status-ok',
  ALERT: 'bg-status-alert-bg text-status-alert',
  CRITICAL: 'bg-status-critical-bg text-status-critical',
}

export const TILE_STYLE: Record<MeterStatus, string> = {
  OK: 'bg-status-ok-bg text-status-ok',
  ALERT: 'bg-status-alert-bg text-status-alert',
  CRITICAL: 'bg-status-critical text-white',
}
export const TILE_UNEVALUATED = 'bg-type-false-positive-bg text-type-false-positive'

export const TYPE_LABEL: Record<AnomalyType, string> = {
  REAL_ANOMALY: 'Anomalía real',
  DATA_QUALITY: 'Calidad de datos',
  EXPLAINABLE_ANOMALY: 'Explicable',
  FALSE_POSITIVE: 'Falso positivo',
}

/** Short labels for tight spaces such as the priority list. */
export const TYPE_LABEL_SHORT: Record<AnomalyType, string> = {
  REAL_ANOMALY: 'Real',
  DATA_QUALITY: 'Datos',
  EXPLAINABLE_ANOMALY: 'Explicable',
  FALSE_POSITIVE: 'Falso positivo',
}

export const TYPE_BADGE: Record<AnomalyType, string> = {
  REAL_ANOMALY: 'bg-status-critical-bg text-status-critical',
  DATA_QUALITY: 'bg-type-data-quality-bg text-type-data-quality',
  EXPLAINABLE_ANOMALY: 'bg-type-explainable-bg text-type-explainable',
  FALSE_POSITIVE: 'bg-type-false-positive-bg text-type-false-positive',
}

export const SEVERITY_LABEL: Record<Severity, string> = {
  HIGH: 'Alta',
  MEDIUM: 'Media',
  LOW: 'Baja',
}
