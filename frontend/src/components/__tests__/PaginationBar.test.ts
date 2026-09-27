import { describe, expect, it } from 'vitest'
import { mount } from '@vue/test-utils'
import PaginationBar from '../PaginationBar.vue'

const bar = (
  props: Partial<{ page: number; pageCount: number; pageSize: number; total: number }> = {},
) => mount(PaginationBar, { props: { page: 1, pageCount: 2, pageSize: 10, total: 12, ...props } })

const buttons = (w: ReturnType<typeof bar>) => w.findAll('nav button')

describe('PaginationBar', () => {
  it('says which meters are shown', () => {
    expect(bar().text()).toContain('Mostrando 1–10 de 12 medidores')
    expect(bar({ page: 2 }).text()).toContain('Mostrando 11–12 de 12 medidores')
  })

  it('offers previous, the pages and next, and marks the current page', () => {
    const w = bar({ page: 2 })
    expect(buttons(w).map((b) => b.attributes('aria-label'))).toEqual([
      'Página anterior',
      'Página 1',
      'Página 2',
      'Página siguiente',
    ])
    expect(w.find('[aria-current=page]').text()).toBe('2')
    expect(w.findAll('[aria-current]')).toHaveLength(1)
  })

  it('disables previous on the first page and next on the last', () => {
    const first = buttons(bar({ page: 1 }))
    expect(first[0].attributes('disabled')).toBeDefined()
    expect(first.at(-1)!.attributes('disabled')).toBeUndefined()
    const last = buttons(bar({ page: 2 }))
    expect(last[0].attributes('disabled')).toBeUndefined()
    expect(last.at(-1)!.attributes('disabled')).toBeDefined()
  })

  it('emits the requested page', async () => {
    const w = bar({ page: 2, pageCount: 3, total: 25 })
    await w.find('[aria-label="Página siguiente"]').trigger('click')
    await w.find('[aria-label="Página anterior"]').trigger('click')
    await w.find('[aria-label="Página 1"]').trigger('click')
    expect(w.emitted('page')).toEqual([[3], [1], [1]])
  })

  it('collapses a long list of pages with an ellipsis', () => {
    const w = bar({ page: 10, pageCount: 20, total: 200 })
    expect(
      w
        .findAll('nav button[aria-label^="Página "]')
        .map((b) => b.text())
        .filter((t) => /^\d+$/.test(t)),
    ).toEqual(['1', '9', '10', '11', '20'])
    expect(w.find('nav').text()).toContain('…')
  })

  it('has no page buttons when everything fits on one page, but still shows the count and the size', () => {
    const w = bar({ pageCount: 1, total: 8 })
    expect(w.find('nav').exists()).toBe(false)
    expect(w.text()).toContain('Mostrando 1–8 de 8 medidores')
    expect(w.find('select').exists()).toBe(true)
  })

  it('lets the user choose how many rows per page', async () => {
    const w = bar()
    const options = w.findAll('select option').map((o) => o.text())
    expect(options).toEqual(['10', '25', '50'])
    expect((w.find('select').element as HTMLSelectElement).value).toBe('10')
    await w.find('select').setValue('25')
    expect(w.emitted('size')).toEqual([[25]])
  })
})
