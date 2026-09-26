import { describe, expect, it } from 'vitest'
import {
  changeRows,
  confidenceBase,
  confidenceLabel,
  confidenceParts,
  investigationTitle,
  ordinal,
  qualityRows,
  signed,
  signedPct,
  splitRule,
  summarizeSeries,
  windowSentence,
} from '../anomalies'
import { anomaly, evidence, points } from './fixtures'

describe('formatting', () => {
  it.each([
    [0.96, 'Alta'],
    [0.85, 'Alta'],
    [0.84, 'Media'],
    [0.65, 'Media'],
    [0.64, 'Baja'],
  ])('confidence %s is %s', (c, want) => expect(confidenceLabel(c)).toBe(want))

  it('uses the design minus sign and no sign for zero', () => {
    expect(signed(-0.2, 2)).toBe('−0,20')
    expect(signed(2.9, 1)).toBe('+2,9')
    expect(signed(0, 1)).toBe('0,0')
    expect(signed(-0.04, 1)).toBe('0,0') // rounds to zero: no misleading sign
    expect(signed(2825, 0)).toBe('+2.825') // zero decimals really means none
    expect(signed(-110.19, 0)).toBe('−110')
    expect(signedPct(-1.3)).toBe('−1,3%')
    expect(signedPct(110.5)).toBe('+110,5%')
  })

  it('writes the rank', () => expect(ordinal(1, 4)).toBe('1.ª de 4'))
})

describe('investigationTitle', () => {
  it.each([
    [{ type: 'REAL_ANOMALY' as const }, 'Consumo duplicado sin causa conocida'],
    [
      {
        type: 'REAL_ANOMALY' as const,
        evidence: evidence({ metrics: { ...evidence().metrics, window_variation_pct: 35 } }),
      },
      'Consumo +35,0% sin causa conocida',
    ],
    [
      {
        type: 'REAL_ANOMALY' as const,
        evidence: evidence({ metrics: { ...evidence().metrics, window_variation_pct: -40 } }),
      },
      'Consumo −40,0% sin causa conocida',
    ],
    [
      {
        type: 'EXPLAINABLE_ANOMALY' as const,
        evidence: evidence({ metrics: { ...evidence().metrics, window_variation_pct: 46.5 } }),
      },
      'Consumo +46,5% por un cambio operativo',
    ],
    [
      {
        type: 'FALSE_POSITIVE' as const,
        evidence: evidence({ metrics: { ...evidence().metrics, window_variation_pct: -79.8 } }),
      },
      'Caída de consumo por parada programada',
    ],
    [{ type: 'DATA_QUALITY' as const }, 'Lecturas eléctricas inconsistentes'],
  ])('%j', (over, want) => expect(investigationTitle(anomaly(over))).toBe(want))

  it('still gives a title without evidence', () => {
    expect(investigationTitle(anomaly({ evidence: undefined }))).toBe(
      'Consumo anómalo sin causa conocida',
    )
  })
})

describe('windowSentence', () => {
  it('describes a window that is still active and one that ended', () => {
    expect(windowSentence(evidence())).toBe('activa desde el 12 sep a las 14:00')
    const ended = evidence({
      window: {
        start: '2026-09-08T00:00:00Z',
        end: '2026-09-08T11:00:00Z',
        hours: 12,
        ongoing: false,
      },
    })
    expect(windowSentence(ended)).toBe('del 8 sep, 00:00 al 8 sep, 11:00')
  })
})

describe('confidence breakdown', () => {
  it('shows each component as a fraction of its maximum', () => {
    const parts = confidenceParts(evidence(), 'REAL_ANOMALY')
    expect(parts.map((p) => p.label)).toEqual([
      'Magnitud de la señal',
      'Señales que coinciden',
      'Ausencia de evento',
    ])
    expect(parts.map((p) => Number(p.fraction.toFixed(2)))).toEqual([1, 1, 0.9])
    expect(confidenceBase(evidence())).toBe(0.4)
  })

  it('names the last component by what it means for the type', () => {
    expect(confidenceParts(evidence(), 'DATA_QUALITY')[2].label).toBe('Claridad de la explicación')
    expect(confidenceParts(evidence(), 'EXPLAINABLE_ANOMALY')[2].label).toBe(
      'Claridad de la explicación',
    )
  })

  it('never exceeds 1 and skips components it has no weight for', () => {
    const e = evidence({
      confidence_breakdown: { base: 0.4, signal_strength: 0.9, independent_signals: 0.2 },
      confidence_weights: { base: 0.4, signal_strength: 0.25, independent_signals: 0.2 },
    })
    const parts = confidenceParts(e, 'REAL_ANOMALY')
    expect(parts.map((p) => p.key)).toEqual(['signal_strength', 'independent_signals'])
    expect(parts[0].fraction).toBe(1)
  })
})

describe('changeRows', () => {
  it('formats every variable like the design', () => {
    const rows = changeRows(evidence())
    expect(rows.map((r) => r.label)).toEqual([
      'Consumo medio',
      'Corriente media',
      'Factor de potencia',
      'Voltaje medio',
      'Variabilidad de voltaje',
    ])
    expect(rows[0]).toMatchObject({
      before: '44,1 kWh/h',
      after: '92,8 kWh/h',
      change: '+110,5%',
      significant: true,
    })
    expect(rows[1]).toMatchObject({ before: '202 A', after: '425 A', change: '+110,2%' })
    expect(rows[2]).toMatchObject({ before: '0,94', after: '0,74', change: '−0,20' }) // PF change is absolute, not a percentage
    expect(rows[3]).toMatchObject({
      before: '219,8 V',
      after: '216,9 V',
      change: '−1,3%',
      significant: false,
    })
    expect(rows[4]).toMatchObject({ before: '±1,2 V', after: '±1,3 V', change: '+9,1%' })
  })
})

describe('qualityRows', () => {
  it('lists only the checks that fired', () => {
    const e = evidence({
      kind: 'data_quality',
      quality: {
        total_flags: 82,
        by_kind: { voltage_jump: 32, power_factor_jump: 24, voltage_out_of_range: 0 },
        recurrent: true,
        first_flag: '',
        last_flag: '',
        max_flags_in_24h: 16,
        flagged_hours: 32,
        window_hours: 47,
      },
    })
    expect(qualityRows(e)).toEqual([
      { label: 'Saltos de voltaje de más de 10 V', count: 32 },
      { label: 'Saltos bruscos del factor de potencia', count: 24 },
    ])
    expect(qualityRows(evidence())).toEqual([])
  })
})

describe('splitRule', () => {
  it('splits a rule into title and detail, capitalizing the detail', () => {
    expect(splitRule('Cambio persistente: 58 h consecutivas')).toEqual({
      title: 'Cambio persistente',
      detail: '58 h consecutivas.',
    })
    expect(splitRule('Deterioro eléctrico: cambió el FP')).toEqual({
      title: 'Deterioro eléctrico',
      detail: 'Cambió el FP.',
    })
    expect(splitRule('Sin evento compatible que explique el cambio')).toEqual({
      title: 'Sin evento compatible que explique el cambio',
      detail: '',
    })
  })
})

describe('summarizeSeries', () => {
  const win = { start: '2026-09-02T00:00:00Z', end: '2026-09-02T23:00:00Z' }
  const week = 14 * 24

  it('reports a power factor drop inside the window', () => {
    const p = points(week, (i) => (i >= 24 && i < 48 ? { power_factor: 0.74 } : {}))
    expect(summarizeSeries(p, 'pf', win)).toEqual({
      delta: '−0,20',
      sub: 'Media 0,94 → 0,74',
      alert: true,
    })
  })

  it('reports a current increase as a percentage', () => {
    const p = points(week, (i) => (i >= 24 && i < 48 ? { current_a: 420 } : {}))
    expect(summarizeSeries(p, 'current', win)).toEqual({
      delta: '+110%',
      sub: 'Media 200 → 420 A',
      alert: true,
    })
  })

  it('flags a voltage that became unstable even if its mean did not move', () => {
    const p = points(week, (i) =>
      i >= 24 && i < 48 ? { voltage_v: i % 2 ? 240 : 200 } : { voltage_v: 220 + (i % 2) * 0.2 },
    )
    const s = summarizeSeries(p, 'voltage', win)
    expect(s.sub).toContain('más inestable')
    expect(s.alert).toBe(true)
  })

  it('does not raise an alert for ordinary variation', () => {
    const p = points(week, (i) => ({ voltage_v: 220 + (i % 2 ? 0.3 : -0.3), current_a: 201 }))
    expect(summarizeSeries(p, 'voltage', win).alert).toBe(false)
    expect(summarizeSeries(p, 'current', win).alert).toBe(false)
    expect(summarizeSeries(p, 'pf', win)).toMatchObject({ delta: '0,00', alert: false })
  })

  it('uses the last 24 hours when there is no anomalous window', () => {
    const p = points(week, (i) => (i >= week - 24 ? { current_a: 300 } : {}))
    expect(summarizeSeries(p, 'current', null).sub).toBe('Media 200 → 300 A')
  })

  it('returns nothing when the window has no data', () => {
    expect(
      summarizeSeries(points(24), 'pf', {
        start: '2027-01-01T00:00:00Z',
        end: '2027-01-02T00:00:00Z',
      }),
    ).toEqual({ delta: '', sub: '', alert: false })
    expect(summarizeSeries([], 'pf', null)).toEqual({ delta: '', sub: '', alert: false })
  })
})
