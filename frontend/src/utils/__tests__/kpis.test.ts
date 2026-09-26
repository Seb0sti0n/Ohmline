import { describe, expect, it } from 'vitest'
import { buildKpis, typeBreakdown } from '../kpis'
import { run, summary } from './fixtures'

const now = new Date(2026, 8, 25, 18, 0)
const byLabel = (ks: ReturnType<typeof buildKpis>) =>
  Object.fromEntries(ks.map((k) => [k.label, k]))

describe('buildKpis', () => {
  it('shows dashes and an invitation before the first analysis', () => {
    const k = byLabel(
      buildKpis(
        summary({
          has_results: false,
          anomalies_count: 0,
          high_priority_count: 0,
          avg_confidence: null,
          anomalies_by_type: {},
          high_priority_meters: [],
          last_analysis: null,
        }),
        null,
        now,
      ),
    )
    expect(k['Medidores']).toMatchObject({ value: '12', sub: '2 ubicaciones' })
    expect(k['Consumo del periodo']).toMatchObject({
      value: '155.251',
      unit: 'kWh',
      sub: '14 días',
    })
    for (const label of ['Anomalías IA', 'Alta prioridad', 'Confianza IA media']) {
      expect(k[label]).toMatchObject({ value: '—', sub: 'Pendiente de análisis' })
    }
    expect(k['Alta prioridad'].tone).toBeUndefined()
    expect(k['Último análisis']).toMatchObject({
      value: 'Sin ejecutar',
      sub: 'Pulsa Run AI Analysis',
    })
  })

  it('fills the analysis KPIs after a completed run', () => {
    const finished = new Date(2026, 8, 25, 9, 42).toISOString()
    const k = byLabel(buildKpis(summary(), run({ finished_at: finished }), now))
    expect(k['Anomalías IA']).toMatchObject({
      value: '4',
      sub: '1 real, 1 de datos, 1 explicable, 1 descartada',
    })
    expect(k['Alta prioridad']).toMatchObject({
      value: '2',
      sub: 'M-109 y M-112',
      tone: 'critical',
    })
    expect(k['Confianza IA media']).toMatchObject({ value: '0,90', sub: 'Media de 4 hallazgos' })
    expect(k['Último análisis']).toMatchObject({ value: 'Hoy, 09:42', sub: 'Completado' })
  })

  it('reflects a run in progress, a failed run and a clean result', () => {
    expect(
      byLabel(buildKpis(summary(), run({ status: 'RUNNING', summary: null }), now))[
        'Último análisis'
      ],
    ).toMatchObject({ value: 'En curso', sub: 'Analizando lecturas' })
    expect(
      byLabel(buildKpis(summary(), run({ status: 'FAILED' }), now))['Último análisis'].value,
    ).toBe('Falló')
    const clean = byLabel(
      buildKpis(
        summary({
          anomalies_count: 0,
          high_priority_count: 0,
          high_priority_meters: [],
          avg_confidence: null,
          anomalies_by_type: {},
        }),
        run(),
        now,
      ),
    )
    expect(clean['Anomalías IA']).toMatchObject({ value: '0', sub: 'Sin hallazgos' })
    expect(clean['Alta prioridad']).toMatchObject({ value: '0', sub: 'Ninguna' })
    expect(clean['Alta prioridad'].tone).toBeUndefined()
    expect(clean['Confianza IA media']).toMatchObject({ value: '—', sub: 'Sin hallazgos' })
  })
})

describe('typeBreakdown', () => {
  it('pluralizes and skips zeros', () => {
    expect(typeBreakdown({ REAL_ANOMALY: 2, FALSE_POSITIVE: 1 })).toBe('2 reales, 1 descartada')
    expect(typeBreakdown({ DATA_QUALITY: 3 })).toBe('3 de datos')
    expect(typeBreakdown({})).toBe('Sin hallazgos')
  })
})
