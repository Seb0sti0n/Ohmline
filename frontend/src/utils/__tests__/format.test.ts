import { describe, expect, it } from 'vitest'
import {
  formatDataTime,
  formatDay,
  formatDayMonth,
  formatDecimal,
  formatInt,
  formatKwh,
  formatLocalTime,
  formatPct,
  formatPeriod,
  formatSeconds,
  joinList,
  plural,
} from '../format'

describe('numbers (es-CO)', () => {
  it.each([
    [731, '731'],
    [1013, '1.013'],
    [155251, '155.251'],
    [0, '0'],
    [1047.7, '1.048'],
  ])('formatInt(%s) = %s', (v, want) => expect(formatInt(v)).toBe(want))

  it('formats kWh', () => expect(formatKwh(2207.6)).toBe('2.208 kWh'))

  it.each([
    [110.7, '+110,7%'],
    [-0.3, '-0,3%'],
    [0, '0,0%'],
    [47.44, '+47,4%'],
    [1.25, '+1,3%'],
  ])('formatPct(%s) = %s', (v, want) => expect(formatPct(v)).toBe(want))

  it('formats decimals', () => {
    expect(formatDecimal(0.9, 2)).toBe('0,90')
    expect(formatDecimal(4.6)).toBe('4,6')
  })

  it('formats seconds', () => expect(formatSeconds(4600)).toBe('4,6 s'))
})

describe('dates', () => {
  it('formats data timestamps in UTC like the engine texts', () => {
    expect(formatDataTime('2026-09-12T14:00:00Z')).toBe('12 sep, 14:00')
    expect(formatDataTime('2026-09-08T00:05:00Z')).toBe('8 sep, 00:05')
  })

  it('formats the data period', () => {
    expect(formatPeriod('2026-09-01', '2026-09-14')).toBe('del 1 al 14 de septiembre de 2026')
    expect(formatPeriod('2026-08-28', '2026-09-10')).toBe(
      'del 28 de agosto al 10 de septiembre de 2026',
    )
  })

  it('reads a day from a date string', () => {
    expect(formatDay('2026-09-03')).toBe(3)
    expect(formatDayMonth('2026-09-12')).toBe('12 sep')
  })

  it('says "hoy" only for the same calendar day', () => {
    const now = new Date(2026, 8, 25, 18, 0)
    expect(formatLocalTime(new Date(2026, 8, 25, 9, 42).toISOString(), now)).toEqual({
      day: 'hoy',
      time: '09:42',
    })
    expect(formatLocalTime(new Date(2026, 8, 24, 9, 42).toISOString(), now)).toEqual({
      day: '24 sep',
      time: '09:42',
    })
  })
})

describe('text helpers', () => {
  it.each([
    [[], ''],
    [['M-109'], 'M-109'],
    [['M-109', 'M-112'], 'M-109 y M-112'],
    [['A', 'B', 'C'], 'A, B y C'],
  ])('joinList(%j) = %s', (items, want) => expect(joinList(items)).toBe(want))

  it('pluralizes', () => {
    expect(plural(1, 'día', 'días')).toBe('1 día')
    expect(plural(14, 'día', 'días')).toBe('14 días')
    expect(plural(0, 'día', 'días')).toBe('0 días')
  })
})
