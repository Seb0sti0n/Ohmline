import type { AnomalyStatus, AnomalyType, EventType, MeterStatus, Severity } from '@/types/api'

// Full class names (not built from pieces) so Tailwind can see them.

export const STATUS_LABEL: Record<MeterStatus, string> = {
  OK: 'Normal',
  ALERT: 'Alerta',
  CRITICAL: 'Crítico',
  UNEVALUATED: 'Sin evaluar',
}

export const STATUS_TEXT: Record<MeterStatus, string> = {
  OK: 'text-status-ok',
  ALERT: 'text-status-alert',
  CRITICAL: 'text-status-critical',
  UNEVALUATED: 'text-ink-muted',
}

export const STATUS_BADGE: Record<MeterStatus, string> = {
  OK: 'bg-status-ok-bg text-status-ok',
  ALERT: 'bg-status-alert-bg text-status-alert',
  CRITICAL: 'bg-status-critical-bg text-status-critical',
  UNEVALUATED: 'bg-type-false-positive-bg text-type-false-positive',
}

export const TILE_STYLE: Record<MeterStatus, string> = {
  OK: 'bg-status-ok-bg text-status-ok',
  ALERT: 'bg-status-alert-bg text-status-alert',
  CRITICAL: 'bg-status-critical text-white',
  UNEVALUATED: 'bg-type-false-positive-bg text-type-false-positive',
}
export const TILE_UNEVALUATED = TILE_STYLE.UNEVALUATED

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

export const SEVERITY_BADGE: Record<Severity, string> = {
  HIGH: 'bg-status-critical-bg text-status-critical',
  MEDIUM: 'bg-status-alert-bg text-status-alert',
  LOW: 'bg-type-false-positive-bg text-type-false-positive',
}

/** The text of the action button for each anomaly type (Anomalías screen). */
export const ACTION_LABEL: Record<AnomalyType, string> = {
  REAL_ANOMALY: 'Investigar',
  DATA_QUALITY: 'Validar medidor',
  EXPLAINABLE_ANOMALY: 'Validar operación',
  FALSE_POSITIVE: 'No escalar',
}

export const TYPE_HELP: Record<AnomalyType, string> = {
  REAL_ANOMALY: 'Cambio persistente sin evento que lo explique.',
  DATA_QUALITY: 'Lecturas eléctricas incoherentes; no confiar hasta validar.',
  EXPLAINABLE_ANOMALY: 'Cambio real que coincide con un evento operativo.',
  FALSE_POSITIVE: 'Desviación temporal explicada por un evento programado.',
}

export const ANOMALY_STATUS_LABEL: Record<AnomalyStatus, string> = {
  OPEN: 'Abierta',
  ACKNOWLEDGED: 'En investigación',
  RESOLVED: 'Resuelta',
}

export const ANOMALY_STATUS_BADGE: Record<AnomalyStatus, string> = {
  OPEN: 'bg-status-critical-bg text-status-critical',
  ACKNOWLEDGED: 'bg-status-alert-bg text-status-alert',
  RESOLVED: 'bg-status-ok-bg text-status-ok',
}

export const EVENT_LABEL: Record<EventType, string> = {
  SCHEDULED_OUTAGE: 'Parada programada',
  OPERATIONAL_CHANGE: 'Cambio operativo',
  DATA_QUALITY: 'Calidad de datos',
  UNKNOWN: 'Desconocido',
}

export const EVENT_BADGE: Record<EventType, string> = {
  SCHEDULED_OUTAGE: 'bg-type-explainable-bg text-type-explainable',
  OPERATIONAL_CHANGE: 'bg-type-explainable-bg text-type-explainable',
  DATA_QUALITY: 'bg-type-data-quality-bg text-type-data-quality',
  UNKNOWN: 'bg-type-false-positive-bg text-type-false-positive',
}

const EVENT_TEXT: Record<string, string> = {
  'New production line activated': 'Nueva línea de producción activada.',
  'Scheduled maintenance outage for 12 hours': 'Parada de mantenimiento programada de 12 horas.',
  'No operational event reported': 'No se reportó ningún evento operativo.',
  'Intermittent readings and abnormal electrical jumps':
    'Lecturas intermitentes y saltos eléctricos anormales.',
}

/** Event descriptions come from the data in English; the known ones are shown in Spanish. */
export const eventText = (description: string) => EVENT_TEXT[description] ?? description
