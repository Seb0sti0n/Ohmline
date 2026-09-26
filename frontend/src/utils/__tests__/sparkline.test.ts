import { describe, expect, it } from 'vitest'
import { sparkline } from '../sparkline'

const ys = (path: string) => path.split(' ').map((p) => Number(p.split(',')[1]))

describe('sparkline', () => {
  it('draws one point per value across the width', () => {
    const { path } = sparkline([1, 2, 3, 4], 2.5)
    expect(path.startsWith('M0.0,')).toBe(true)
    expect(path.split(' ')).toHaveLength(4)
    expect(path).toContain('L96.0,')
  })

  it('keeps a healthy meter flat and lets a step change stand out', () => {
    const flat = ys(
      sparkline(
        Array(14)
          .fill(700)
          .map((v, i) => v + (i % 2) * 5),
        700,
      ).path,
    )
    const step = ys(sparkline([...Array(10).fill(1000), ...Array(4).fill(2200)], 1000).path)
    expect(Math.max(...flat) - Math.min(...flat)).toBeLessThan(2)
    expect(Math.max(...step) - Math.min(...step)).toBeGreaterThan(15)
    expect(step[13]).toBeLessThan(step[0]) // SVG y grows downwards: higher consumption = smaller y
  })

  it('always places the baseline inside the drawing area', () => {
    for (const values of [
      [10, 10, 10],
      [5, 9, 30],
      [100, 1, 1],
    ]) {
      const { baselineY } = sparkline(values, 10)
      expect(baselineY).toBeGreaterThanOrEqual(0)
      expect(baselineY).toBeLessThanOrEqual(28)
    }
  })

  it('handles an empty series and a single point', () => {
    expect(sparkline([], 10).path).toBe('')
    expect(sparkline([5], 5).path).toMatch(/^M0\.0,/)
  })
})
