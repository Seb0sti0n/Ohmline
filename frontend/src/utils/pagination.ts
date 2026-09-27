import { formatInt } from './format'

export const PAGE_SIZES = [10, 25, 50] as const
export const DEFAULT_PAGE_SIZE = 10

export const pageCount = (total: number, pageSize: number) =>
  Math.max(1, Math.ceil(total / pageSize))

/**
 * The page numbers to show: the first and last page, the current one and its neighbours, with "…"
 * for the gaps (a gap of a single page is shown as that page: 1 2 3 … is never "1 … 3").
 * `1 … 4 5 6 … 20`
 */
export function pageWindow(current: number, total: number): (number | '…')[] {
  if (total <= 7) return Array.from({ length: total }, (_, i) => i + 1)
  const wanted = [...new Set([1, total, current - 1, current, current + 1])]
    .filter((p) => p >= 1 && p <= total)
    .sort((a, b) => a - b)
  const out: (number | '…')[] = []
  wanted.forEach((p, i) => {
    const prev = wanted[i - 1]
    if (prev !== undefined && p - prev === 2) out.push(prev + 1)
    else if (prev !== undefined && p - prev > 2) out.push('…')
    out.push(p)
  })
  return out
}

/** "1–10 de 12" for the rows shown on a page. */
export function rangeLabel(page: number, pageSize: number, total: number): string {
  if (total === 0) return '0'
  const start = (page - 1) * pageSize + 1
  const end = Math.min(page * pageSize, total)
  return `${formatInt(start)}–${formatInt(end)} de ${formatInt(total)}`
}
