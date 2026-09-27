import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import SpikesNote from '../meter/SpikesNote.vue'

const text = (count: number) => mount(SpikesNote, { props: { count } }).text()

describe('SpikesNote', () => {
  it('shows the count and says it is not treated as an anomaly', () => {
    const t = text(23)
    expect(t).toContain('23 picos aislados')
    expect(t).toContain('no se trata como anomalía')
    expect(t).toContain('no es un cambio persistente')
  })

  it('uses the singular for one spike', () => {
    expect(text(1)).toContain('1 pico aislado:')
    expect(text(1)).not.toContain('picos aislados')
  })

  it('formats large counts with the thousands separator', () =>
    expect(text(1234)).toContain('1.234 picos aislados'))

  it('says so when there are none', () => {
    const t = text(0)
    expect(t).toContain('Sin picos aislados')
    expect(t).not.toContain('no se trata como anomalía')
  })
})
