import { describe, expect, it } from 'vitest'
import { buildSteps, resultText, statusText } from '../analysis'
import { run, steps } from './fixtures'

const byType = { REAL_ANOMALY: 1, DATA_QUALITY: 1, EXPLAINABLE_ANOMALY: 1, FALSE_POSITIVE: 1 }

describe('buildSteps', () => {
  it('shows seven pending steps before any analysis', () => {
    const s = buildSteps(null)
    expect(s.map((x) => x.label)).toEqual([
      'Lecturas',
      'Baseline',
      'Detección',
      'Correlación',
      'Eventos',
      'Explicación',
      'Recomendación',
    ])
    expect(s.every((x) => x.state === 'pending' && x.detail === '')).toBe(true)
    expect(s.map((x) => x.number)).toEqual([1, 2, 3, 4, 5, 6, 7])
  })

  it('marks done, active and pending steps while running, with durations for the done ones', () => {
    const r = run({
      status: 'RUNNING',
      summary: null,
      finished_at: null,
      steps: steps(['DONE', 'DONE', 'RUNNING'], 600),
    })
    const s = buildSteps(r)
    expect(s.map((x) => x.state)).toEqual([
      'done',
      'done',
      'active',
      'pending',
      'pending',
      'pending',
      'pending',
    ])
    expect(s[0].detail).toBe('0,6 s')
    expect(s[2].detail).toBe('En curso')
    expect(s[3].detail).toBe('')
  })

  it('shows what each stage found once completed', () => {
    const s = buildSteps(run(), byType)
    expect(s.every((x) => x.state === 'done')).toBe(true)
    expect(s.map((x) => x.detail)).toEqual([
      '4.032 lecturas',
      '288 perfiles horarios',
      '3 cambios, 1 serie irregular',
      'Corriente, FP, voltaje',
      '4 eventos cruzados',
      '4 explicaciones',
      '4 acciones',
    ])
  })

  it('handles singular forms and no findings', () => {
    const one = run({ summary: { ...run().summary!, anomalies: 1, events_count: 1 } })
    const d = buildSteps(one, { REAL_ANOMALY: 1 }).map((x) => x.detail)
    expect(d[2]).toBe('1 cambio')
    expect(d[4]).toBe('1 evento cruzado')
    expect(d[5]).toBe('1 explicación')
    expect(d[6]).toBe('1 acción')
    expect(buildSteps(run(), {})[2].detail).toBe('Sin hallazgos')
  })
})

describe('statusText', () => {
  const now = new Date(2026, 8, 25, 18, 0)
  it('describes every state', () => {
    expect(statusText(null)).toBe('Sin ejecutar')
    expect(statusText(run({ status: 'FAILED' }))).toBe('El análisis falló')
    expect(statusText(run({ status: 'RUNNING', steps: steps(['DONE', 'DONE', 'RUNNING']) }))).toBe(
      'Paso 3 de 7',
    )
    expect(statusText(run({ status: 'RUNNING', steps: steps(['DONE', 'DONE', 'DONE']) }))).toBe(
      'Paso 4 de 7',
    )
    const finished = new Date(2026, 8, 25, 9, 42).toISOString()
    expect(statusText(run({ finished_at: finished }), now)).toBe(
      'Completado hoy a las 09:42 en 4,6 s',
    )
    const earlier = new Date(2026, 8, 20, 9, 42).toISOString()
    expect(statusText(run({ finished_at: earlier }), now)).toBe(
      'Completado el 20 sep a las 09:42 en 4,6 s',
    )
  })
})

describe('resultText', () => {
  it('matches the wording of the brief', () => {
    expect(resultText(run())).toBe('4 anomalías detectadas, 2 requieren atención prioritaria')
  })
  it('handles singular, none and not finished', () => {
    const s = run().summary!
    expect(resultText(run({ summary: { ...s, anomalies: 1, high_priority: 1 } }))).toBe(
      '1 anomalía detectada, 1 requiere atención prioritaria',
    )
    expect(resultText(run({ summary: { ...s, anomalies: 3, high_priority: 0 } }))).toBe(
      '3 anomalías detectadas, ninguna requiere atención prioritaria',
    )
    expect(resultText(run({ summary: { ...s, anomalies: 0, high_priority: 0 } }))).toBe(
      'No se detectaron anomalías',
    )
    expect(resultText(run({ status: 'RUNNING', summary: null }))).toBeNull()
    expect(resultText(null)).toBeNull()
  })
})
