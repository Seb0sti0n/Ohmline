import type { Anomaly, AnomalyType, Evidence, ReadingPoint } from '@/types/api'
import { formatDataTime, formatDecimal, formatInt, formatPct } from './format'

/** Alta / Media / Baja, from a 0-1 confidence. */
export function confidenceLabel(c: number): 'Alta' | 'Media' | 'Baja' {
  if (c >= 0.85) return 'Alta'
  if (c >= 0.65) return 'Media'
  return 'Baja'
}

const MINUS = '−' // typographic minus, as in the design ("−0,20")

/** A signed number with the design's minus sign. `digits` decimals; zero has no sign. */
export function signed(v: number, digits: 0 | 1 | 2 = 1): string {
  const body = digits === 0 ? formatInt(Math.abs(v)) : formatDecimal(Math.abs(v), digits)
  const rounded = Number(Math.abs(v).toFixed(digits))
  if (rounded === 0) return body
  return `${v < 0 ? MINUS : '+'}${body}`
}

export const signedPct = (v: number) => formatPct(v).replace('-', MINUS)

/** The page title of an investigation, e.g. "Consumo duplicado sin causa conocida". */
export function investigationTitle(a: Anomaly): string {
  const pct = a.evidence?.metrics.window_variation_pct
  switch (a.type) {
    case 'REAL_ANOMALY':
      if (pct === undefined) return 'Consumo anómalo sin causa conocida'
      return `${pct >= 90 ? 'Consumo duplicado' : `Consumo ${signedPct(pct)}`} sin causa conocida`
    case 'EXPLAINABLE_ANOMALY':
      return `Consumo ${pct === undefined ? 'distinto' : signedPct(pct)} por un cambio operativo`
    case 'FALSE_POSITIVE':
      return pct !== undefined && pct > 0
        ? 'Desviación explicada por un evento programado'
        : 'Caída de consumo por parada programada'
    default:
      return 'Lecturas eléctricas inconsistentes'
  }
}

/** "activa desde el 12 sep a las 14:00" or "del 8 sep, 00:00 al 8 sep, 11:00". */
export function windowSentence(evidence: Evidence): string {
  const { start, end, ongoing } = evidence.window
  if (ongoing) {
    const [day, time] = formatDataTime(start).split(', ')
    return `activa desde el ${day} a las ${time}`
  }
  return `del ${formatDataTime(start)} al ${formatDataTime(end)}`
}

export const ordinal = (n: number, total: number) => `${n}.ª de ${total}`

// ---- confidence breakdown ----

export interface ConfidencePart {
  key: string
  label: string
  /** 0-1: how much of the component's maximum was earned. */
  fraction: number
  /** What it added to the confidence. */
  contribution: number
}

const CONFIDENCE_LABELS: Record<string, string> = {
  signal_strength: 'Magnitud de la señal',
  independent_signals: 'Señales que coinciden',
  explanation_clarity: 'Claridad de la explicación',
}

/** The confidence components (except the fixed base) as fractions of their maximum. */
export function confidenceParts(evidence: Evidence, type: AnomalyType): ConfidencePart[] {
  const labels: Record<string, string> = {
    ...CONFIDENCE_LABELS,
    explanation_clarity:
      type === 'REAL_ANOMALY' ? 'Ausencia de evento' : CONFIDENCE_LABELS.explanation_clarity,
  }
  return Object.keys(CONFIDENCE_LABELS).flatMap((key) => {
    const max = evidence.confidence_weights[key]
    const contribution = evidence.confidence_breakdown[key]
    if (!max || contribution === undefined) return []
    return [{ key, label: labels[key], fraction: Math.min(1, contribution / max), contribution }]
  })
}

export const confidenceBase = (evidence: Evidence) => evidence.confidence_breakdown.base ?? 0

// ---- variables that changed ----

export interface ChangeRow {
  label: string
  before: string
  after: string
  change: string
  significant: boolean
}

const VARIABLE_LABEL: Record<string, string> = {
  consumption_kwh: 'Consumo medio',
  current_a: 'Corriente media',
  power_factor: 'Factor de potencia',
  voltage_v: 'Voltaje medio',
  voltage_std: 'Variabilidad de voltaje',
}

/** Rows of the "Variables que cambiaron" table, formatted as in the design. */
export function changeRows(evidence: Evidence): ChangeRow[] {
  return evidence.changed_variables.map((c) => {
    const one = (v: number) => formatDecimal(v, 1)
    let before: string, after: string, change: string
    switch (c.variable) {
      case 'consumption_kwh':
        ;[before, after, change] = [
          `${one(c.baseline)} kWh/h`,
          `${one(c.observed)} kWh/h`,
          signedPct(c.delta_pct),
        ]
        break
      case 'current_a':
        ;[before, after, change] = [
          `${formatInt(c.baseline)} A`,
          `${formatInt(c.observed)} A`,
          signedPct(c.delta_pct),
        ]
        break
      case 'power_factor':
        ;[before, after, change] = [
          formatDecimal(c.baseline, 2),
          formatDecimal(c.observed, 2),
          signed(c.delta, 2),
        ]
        break
      case 'voltage_v':
        ;[before, after, change] = [
          `${one(c.baseline)} V`,
          `${one(c.observed)} V`,
          signedPct(c.delta_pct),
        ]
        break
      default:
        ;[before, after, change] = [
          `±${one(c.baseline)} V`,
          `±${one(c.observed)} V`,
          signedPct(c.delta_pct),
        ]
    }
    return {
      label: VARIABLE_LABEL[c.variable] ?? c.variable,
      before,
      after,
      change,
      significant: c.significant,
    }
  })
}

const QUALITY_LABEL: Record<string, string> = {
  voltage_jump: 'Saltos de voltaje de más de 10 V',
  voltage_out_of_range: 'Voltaje fuera de ±5% del nominal',
  power_factor_jump: 'Saltos bruscos del factor de potencia',
  kwh_incoherent_with_v_i_pf: 'kWh que no cuadra con V·I·FP',
}

/** Rows for a data-quality anomaly: how many readings tripped each check. */
export function qualityRows(evidence: Evidence): { label: string; count: number }[] {
  const by = evidence.quality?.by_kind ?? {}
  return Object.keys(QUALITY_LABEL)
    .filter((k) => (by[k] ?? 0) > 0)
    .map((k) => ({ label: QUALITY_LABEL[k], count: by[k] }))
}

/** "Cambio persistente: 58 h..." → title and detail; a rule without a colon is all title. */
export function splitRule(rule: string): { title: string; detail: string } {
  const i = rule.indexOf(': ')
  if (i < 0) return { title: rule, detail: '' }
  const detail = rule.slice(i + 2)
  return { title: rule.slice(0, i), detail: `${detail.charAt(0).toUpperCase()}${detail.slice(1)}.` }
}

// ---- small series charts on the meter detail ----

export interface SeriesSummary {
  delta: string
  sub: string
  alert: boolean
}

type Key = 'voltage' | 'current' | 'pf'

const std = (xs: number[]) => {
  if (xs.length < 2) return 0
  const m = xs.reduce((a, b) => a + b, 0) / xs.length
  return Math.sqrt(xs.reduce((a, b) => a + (b - m) ** 2, 0) / (xs.length - 1))
}
const mean = (xs: number[]) => (xs.length ? xs.reduce((a, b) => a + b, 0) / xs.length : 0)

/**
 * Compares a variable inside the anomalous window (or the last 24 h if there is none) with its
 * baseline for the same hours. `alert` uses the same thresholds as the engine (config/engine.go)
 * and only decides the color of the number.
 */
export function summarizeSeries(
  points: ReadingPoint[],
  key: Key,
  window: { start: string; end: string } | null,
): SeriesSummary {
  const value = (p: ReadingPoint) =>
    ({ voltage: p.voltage_v, current: p.current_a, pf: p.power_factor })[key]
  const base = (p: ReadingPoint) =>
    ({ voltage: p.baseline_voltage_v, current: p.baseline_current_a, pf: p.baseline_power_factor })[
      key
    ]

  const inWindow = window
    ? points.filter((p) => p.timestamp >= window.start && p.timestamp <= window.end)
    : points.slice(-24)
  if (inWindow.length === 0) return { delta: '', sub: '', alert: false }
  const obs = mean(inWindow.map(value))
  const exp = mean(inWindow.map(base))

  switch (key) {
    case 'voltage': {
      const reference = std(points.slice(0, 168).map(value)) // the first week is the baseline
      const unstable = reference > 0 && std(inWindow.map(value)) / reference > 1.5
      const d = obs - exp
      return {
        delta: `${signed(d, 1)} V`,
        sub: `Media ${formatDecimal(exp, 1)} → ${formatDecimal(obs, 1)} V${unstable ? ', más inestable' : ''}`,
        alert: Math.abs(d) > 1.5 || unstable,
      }
    }
    case 'current': {
      const pct = exp === 0 ? 0 : (obs / exp - 1) * 100
      return {
        delta: `${signed(pct, 0)}%`,
        sub: `Media ${formatInt(exp)} → ${formatInt(obs)} A`,
        alert: Math.abs(pct) > 15,
      }
    }
    default: {
      const d = obs - exp
      return {
        delta: signed(d, 2),
        sub: `Media ${formatDecimal(exp, 2)} → ${formatDecimal(obs, 2)}`,
        alert: Math.abs(d) > 0.05,
      }
    }
  }
}

/** Sentence of the dark bar at the top of the meter detail. */
export function verdict(
  a: Anomaly,
  label: string,
  severity: string,
): { strong: string; rest: string } {
  return {
    strong: `${label}, severidad ${severity}, confianza ${formatDecimal(a.confidence, 2)}.`,
    rest: a.reason,
  }
}
