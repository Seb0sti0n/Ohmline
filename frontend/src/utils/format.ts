// Number and date formatting for the UI: es-CO conventions from design/DESIGN.md
// (1.048 kWh, +110,7%, "12 sep, 14:00").

const int = new Intl.NumberFormat('es-CO', { maximumFractionDigits: 0 })
const one = new Intl.NumberFormat('es-CO', { minimumFractionDigits: 1, maximumFractionDigits: 1 })
const two = new Intl.NumberFormat('es-CO', { minimumFractionDigits: 2, maximumFractionDigits: 2 })
const signedOne = new Intl.NumberFormat('es-CO', {
  minimumFractionDigits: 1,
  maximumFractionDigits: 1,
  signDisplay: 'exceptZero',
})

export const formatInt = (v: number) => int.format(v)
export const formatDecimal = (v: number, digits: 1 | 2 = 1) => (digits === 1 ? one : two).format(v)
export const formatKwh = (v: number) => `${int.format(v)} kWh`
/** Variation with sign: +110,7% */
export const formatPct = (v: number) => `${signedOne.format(v)}%`

const MONTHS = ['ene', 'feb', 'mar', 'abr', 'may', 'jun', 'jul', 'ago', 'sep', 'oct', 'nov', 'dic']
const MONTHS_LONG = [
  'enero',
  'febrero',
  'marzo',
  'abril',
  'mayo',
  'junio',
  'julio',
  'agosto',
  'septiembre',
  'octubre',
  'noviembre',
  'diciembre',
]
const pad = (n: number) => String(n).padStart(2, '0')

/** Timestamps in the data are UTC and the engine writes them in UTC: "12 sep, 14:00". */
export function formatDataTime(iso: string): string {
  const d = new Date(iso)
  return `${d.getUTCDate()} ${MONTHS[d.getUTCMonth()]}, ${pad(d.getUTCHours())}:${pad(d.getUTCMinutes())}`
}

/** "1 de septiembre" style range for the data period, from two YYYY-MM-DD dates. */
export function formatPeriod(from: string, to: string): string {
  const a = new Date(`${from}T00:00:00Z`)
  const b = new Date(`${to}T00:00:00Z`)
  const month = MONTHS_LONG[b.getUTCMonth()]
  if (a.getUTCMonth() === b.getUTCMonth() && a.getUTCFullYear() === b.getUTCFullYear()) {
    return `del ${a.getUTCDate()} al ${b.getUTCDate()} de ${month} de ${b.getUTCFullYear()}`
  }
  return `del ${a.getUTCDate()} de ${MONTHS_LONG[a.getUTCMonth()]} al ${b.getUTCDate()} de ${month} de ${b.getUTCFullYear()}`
}

export const formatDay = (isoDate: string) => Number(isoDate.slice(8, 10))

/** "12 sep" from a YYYY-MM-DD date. */
export const formatDayMonth = (isoDate: string) =>
  `${formatDay(isoDate)} ${MONTHS[Number(isoDate.slice(5, 7)) - 1]}`

/** Real clock time of an event that happened now (an analysis run), in the browser's time zone. */
export function formatLocalTime(
  iso: string,
  now: Date = new Date(),
): { day: string; time: string } {
  const d = new Date(iso)
  const time = `${pad(d.getHours())}:${pad(d.getMinutes())}`
  const sameDay = d.toDateString() === now.toDateString()
  return { day: sameDay ? 'hoy' : `${d.getDate()} ${MONTHS[d.getMonth()]}`, time }
}

export const formatSeconds = (ms: number) => `${one.format(ms / 1000)} s`

/** "M-109 y M-112", "A, B y C" */
export function joinList(items: string[]): string {
  if (items.length <= 1) return items.join('')
  return `${items.slice(0, -1).join(', ')} y ${items[items.length - 1]}`
}

export const plural = (n: number, one: string, many: string) =>
  `${formatInt(n)} ${n === 1 ? one : many}`
