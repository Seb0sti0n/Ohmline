import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import AnomaliesTable from '../AnomaliesTable.vue'
import { anomaly } from '@/utils/__tests__/fixtures'

const stubs = { RouterLink: { props: ['to'], template: '<a :href="String(to)"><slot /></a>' } }

const items = [
  anomaly({
    id: 1,
    meter_id: 'M-109',
    type: 'REAL_ANOMALY',
    severity: 'HIGH',
    confidence: 0.94,
    priority_score: 100,
    reason: 'Motivo real.',
  }),
  anomaly({
    id: 2,
    meter_id: 'M-112',
    type: 'DATA_QUALITY',
    severity: 'HIGH',
    confidence: 0.88,
    priority_score: 71.8,
    status: 'ACKNOWLEDGED',
  }),
  anomaly({
    id: 3,
    meter_id: 'M-104',
    type: 'EXPLAINABLE_ANOMALY',
    severity: 'MEDIUM',
    confidence: 0.7,
    priority_score: 46.6,
  }),
  anomaly({
    id: 4,
    meter_id: 'M-106',
    type: 'FALSE_POSITIVE',
    severity: 'LOW',
    confidence: 0.5,
    priority_score: 19,
  }),
]

describe('AnomaliesTable', () => {
  const rows = () =>
    mount(AnomaliesTable, { props: { items }, global: { stubs } }).findAll('tbody tr')

  it('numbers the rows by priority and shows the score', () => {
    const r = rows()
    expect(r).toHaveLength(4)
    expect(r[0].find('td').text()).toContain('1')
    expect(r[0].find('td').text()).toContain('100/100')
    expect(r[1].find('td').text()).toContain('72/100')
    expect(r[3].find('td').text()).toContain('19/100')
  })

  it('shows type, severity, confidence label and value, and the reason', () => {
    const first = rows()[0].text()
    for (const text of [
      'M-109',
      'Tablero principal B',
      'Anomalía real',
      'Alta',
      '0,94',
      'Motivo real.',
    ])
      expect(first).toContain(text)
    expect(rows()[2].text()).toContain('Media') // severity MEDIUM…
    expect(rows()[2].text()).toContain('0,70') // …and a confidence that is only "Media"
    expect(rows()[3].text()).toContain('Baja')
  })

  it('names the action after the type and links to the investigation', () => {
    const links = rows().map((r) => r.findAll('a').at(-1)!)
    expect(links.map((a) => a.text())).toEqual([
      'Investigar',
      'Validar medidor',
      'Validar operación',
      'No escalar',
    ])
    expect(links.map((a) => a.attributes('href'))).toEqual([
      '/anomalies/1',
      '/anomalies/2',
      '/anomalies/3',
      '/anomalies/4',
    ])
  })

  it('highlights the action only for high severity', () => {
    const links = rows().map((r) => r.findAll('a').at(-1)!)
    expect(links[0].classes()).toContain('bg-brand')
    expect(links[1].classes()).toContain('bg-brand')
    expect(links[2].classes()).not.toContain('bg-brand')
    expect(links[3].classes()).not.toContain('bg-brand')
  })

  it('shows the workflow status only once it has moved past "open"', () => {
    const r = rows()
    expect(r[0].text()).not.toContain('Abierta')
    expect(r[1].text()).toContain('En investigación')
  })
})
