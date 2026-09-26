import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import ActionCard from '../investigation/ActionCard.vue'
import { anomaly } from '@/utils/__tests__/fixtures'
import type { AnomalyStatus } from '@/types/api'

const mountCard = (
  status: AnomalyStatus,
  extra: { updating?: boolean; error?: string | null } = {},
) =>
  mount(ActionCard, {
    props: { anomaly: anomaly({ status }), updating: false, error: null, ...extra },
  })

const labels = (w: ReturnType<typeof mountCard>) => w.findAll('button').map((b) => b.text())

describe('ActionCard', () => {
  it('shows the recommended action', () => {
    expect(mountCard('OPEN').text()).toContain('Investigar medidor e instalación.')
  })

  it('offers to investigate or resolve an open anomaly', async () => {
    const w = mountCard('OPEN')
    expect(labels(w)).toEqual(['Marcar en investigación', 'Resolver'])
    await w.findAll('button')[0].trigger('click')
    await w.findAll('button')[1].trigger('click')
    expect(w.emitted('set')).toEqual([['ACKNOWLEDGED'], ['RESOLVED']])
  })

  it('shows an anomaly under investigation as such, and can only resolve it', async () => {
    const w = mountCard('ACKNOWLEDGED')
    expect(labels(w)).toEqual(['En investigación', 'Resolver'])
    expect(w.findAll('button')[0].attributes('disabled')).toBeDefined()
    await w.findAll('button')[1].trigger('click')
    expect(w.emitted('set')).toEqual([['RESOLVED']])
  })

  it('lets a resolved anomaly be reopened and hides Resolver', async () => {
    const w = mountCard('RESOLVED')
    expect(labels(w)).toEqual(['Reabrir'])
    await w.find('button').trigger('click')
    expect(w.emitted('set')).toEqual([['OPEN']])
  })

  it('disables the buttons while updating and shows an error', () => {
    const w = mountCard('OPEN', { updating: true, error: 'No se pudo actualizar el estado' })
    expect(w.findAll('button').every((b) => b.attributes('disabled') !== undefined)).toBe(true)
    expect(w.find('[role=alert]').text()).toBe('No se pudo actualizar el estado')
  })
})
