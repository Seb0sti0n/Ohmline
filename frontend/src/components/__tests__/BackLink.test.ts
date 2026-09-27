import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import BackLink from '../BackLink.vue'

const stubs = {
  RouterLink: { props: ['to'], template: '<a :href="String(to)" v-bind="$attrs"><slot /></a>' },
}

describe('BackLink', () => {
  it('links to the given route with its label', () => {
    const w = mount(BackLink, { props: { to: '/meters', label: 'Medidores' }, global: { stubs } })
    expect(w.attributes('href')).toBe('/meters')
    expect(w.text()).toBe('Medidores')
  })

  // It must look like the primary "Run AI Analysis" button, not like a text link.
  it('is styled as a primary button', () => {
    const w = mount(BackLink, { props: { to: '/', label: 'Volver' }, global: { stubs } })
    for (const cls of ['bg-brand', 'text-white', 'h-12', 'rounded-md', 'font-semibold'])
      expect(w.classes()).toContain(cls)
    expect(w.classes()).not.toContain('text-brand')
  })
})
