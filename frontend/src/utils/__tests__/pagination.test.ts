import { describe, expect, it } from 'vitest'
import { pageCount, pageWindow, rangeLabel } from '../pagination'

describe('pageCount', () => {
  it.each([
    [0, 10, 1],
    [1, 10, 1],
    [10, 10, 1],
    [11, 10, 2],
    [12, 10, 2],
    [100, 25, 4],
    [101, 25, 5],
  ])('%s meters at %s per page = %s page(s)', (total, size, want) =>
    expect(pageCount(total, size)).toBe(want),
  )
})

describe('pageWindow', () => {
  it('shows every page when there are few', () => {
    expect(pageWindow(1, 1)).toEqual([1])
    expect(pageWindow(2, 3)).toEqual([1, 2, 3])
    expect(pageWindow(4, 7)).toEqual([1, 2, 3, 4, 5, 6, 7])
  })

  it.each([
    [1, 20, [1, 2, '…', 20]],
    [2, 20, [1, 2, 3, '…', 20]],
    [3, 20, [1, 2, 3, 4, '…', 20]],
    [10, 20, [1, '…', 9, 10, 11, '…', 20]],
    [18, 20, [1, '…', 17, 18, 19, 20]],
    [19, 20, [1, '…', 18, 19, 20]],
    [20, 20, [1, '…', 19, 20]],
  ])('page %s of %s', (current, total, want) => expect(pageWindow(current, total)).toEqual(want))

  it('never hides a single page behind an ellipsis, and never repeats or skips the current one', () => {
    for (let total = 8; total <= 30; total++) {
      for (let current = 1; current <= total; current++) {
        const w = pageWindow(current, total)
        expect(w).toContain(current)
        expect(w[0]).toBe(1)
        expect(w.at(-1)).toBe(total)
        const nums = w.filter((x): x is number => x !== '…')
        expect(nums).toEqual([...new Set(nums)].sort((a, b) => a - b)) // ascending, no duplicates
        w.forEach((x, i) => {
          if (x !== '…') return
          // an ellipsis hides at least two pages
          expect((w[i + 1] as number) - (w[i - 1] as number)).toBeGreaterThan(2)
        })
      }
    }
  })
})

describe('rangeLabel', () => {
  it.each([
    [1, 10, 12, '1–10 de 12'],
    [2, 10, 12, '11–12 de 12'],
    [1, 10, 3, '1–3 de 3'],
    [3, 50, 1234, '101–150 de 1.234'],
    [1, 10, 0, '0'],
  ])('page %s, %s per page, %s total', (page, size, total, want) =>
    expect(rangeLabel(page, size, total)).toBe(want),
  )
})
