export interface Sparkline {
  path: string
  baselineY: number
}

/**
 * SVG path of a series drawn against its baseline. The vertical scale always includes the baseline
 * plus some margin, so a healthy meter looks flat and a step change stands out.
 */
export function sparkline(
  values: number[],
  baseline: number,
  width = 96,
  height = 28,
  pad = 2,
): Sparkline {
  if (values.length === 0) return { path: '', baselineY: height / 2 }
  const margin = Math.abs(baseline) * 0.1
  const lo = Math.min(baseline - margin, ...values)
  const hi = Math.max(baseline + margin, ...values)
  const y = (v: number) => pad + (1 - (v - lo) / (hi - lo || 1)) * (height - 2 * pad)
  const step = values.length > 1 ? width / (values.length - 1) : 0
  const path = values
    .map((v, i) => `${i === 0 ? 'M' : 'L'}${(i * step).toFixed(1)},${y(v).toFixed(1)}`)
    .join(' ')
  return { path, baselineY: Number(y(baseline).toFixed(1)) }
}
