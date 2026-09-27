import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import MetersTable from '../MetersTable.vue'
import { meter } from '@/utils/__tests__/fixtures'
import type { Anomaly } from '@/types/api'

const stubs = { RouterLink: { props: ['to'], template: '<a :href="String(to)"><slot /></a>' } }
const anomaly = { id: 1, type: 'REAL_ANOMALY' } as Anomaly

const rows = [
  meter('M-109', {
    name: 'Tablero principal B',
    status: 'CRITICAL',
    consumption_kwh: 2207.6,
    baseline_kwh: 1047.7,
    variation_pct: 110.7,
    anomaly,
  }),
  meter('M-101', { name: 'Compresores A', variation_pct: -0.3, consumption_kwh: 729 }),
]

const mountTable = (
  sort: 'severity' | 'consumption' | 'variation' = 'severity',
  order: 'asc' | 'desc' = 'desc',
) => mount(MetersTable, { props: { rows, sort, order }, global: { stubs } })

describe('MetersTable', () => {
  it('renders each meter with formatted numbers, status and anomaly', () => {
    const w = mountTable()
    const first = w.findAll('tbody tr')[0].text()
    expect(first).toContain('M-109')
    expect(first).toContain('Tablero principal B')
    expect(first).toContain('2.208 kWh')
    expect(first).toContain('1.048 kWh')
    expect(first).toContain('+110,7%')
    expect(first).toContain('Crítico')
    expect(first).toContain('Anomalía real')
    const second = w.findAll('tbody tr')[1].text()
    expect(second).toContain('-0,3%')
    expect(second).toContain('Normal')
    expect(second).toContain('—') // no anomaly
  })

  it('highlights large variations only', () => {
    const cells = mountTable()
      .findAll('tbody tr')
      .map((r) => r.findAll('td')[4])
    expect(cells[0].classes()).toContain('text-status-critical')
    expect(cells[1].classes()).not.toContain('text-status-critical')
  })

  it('links each meter to its detail', () => {
    expect(mountTable().find('tbody a').attributes('href')).toBe('/meters/M-109')
  })

  it('emits the clicked column and marks the active one with an arrow and aria-sort', async () => {
    const w = mountTable('variation', 'asc')
    const buttons = w.findAll('thead button')
    expect(buttons.map((b) => b.text())).toEqual(['Consumo', 'Variación↑', 'Estado'])
    expect(w.findAll('thead th')[4].attributes('aria-sort')).toBe('ascending')
    expect(w.findAll('thead th')[2].attributes('aria-sort')).toBe('none')
    await buttons[0].trigger('click')
    await buttons[2].trigger('click')
    expect(w.emitted('sort')).toEqual([['consumption'], ['severity']])
  })

  it('draws a sparkline per row', () => {
    expect(mountTable().findAll('tbody svg')).toHaveLength(2)
  })

  it('says "Sin evaluar" (not "Normal") for a meter that has not been analysed', () => {
    const w = mount(MetersTable, {
      props: { rows: [meter('M-101', { status: 'UNEVALUATED' })], sort: 'severity', order: 'desc' },
      global: { stubs },
    })
    const row = w.find('tbody tr').text()
    expect(row).toContain('Sin evaluar')
    expect(row).not.toContain('Normal')
  })
})
