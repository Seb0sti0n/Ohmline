/** Reads a design token (a CSS variable generated from design/tokens.json), e.g. cssColor('brand'). */
export function cssColor(token: string, fallback = '#000000'): string {
  const v = getComputedStyle(document.documentElement).getPropertyValue(`--color-${token}`).trim()
  return v || fallback
}
