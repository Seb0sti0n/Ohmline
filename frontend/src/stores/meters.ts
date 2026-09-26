import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { errorMessage, getMeters } from '@/services/api'
import type { MeterFilter, MeterQuery, MeterSort, MeterSummary, SortOrder } from '@/types/api'

export const useMetersStore = defineStore('meters', () => {
  /** Every meter, unfiltered: used for the filter counters. */
  const all = ref<MeterSummary[]>([])
  /** The table rows: filtered, searched and sorted by the API. */
  const rows = ref<MeterSummary[]>([])
  const filter = ref<MeterFilter>('ALL')
  const search = ref('')
  const sort = ref<MeterSort>('severity')
  const order = ref<SortOrder>('desc')
  const loading = ref(false)
  const error = ref<string | null>(null)

  const counts = computed(() => ({
    ALL: all.value.length,
    OK: all.value.filter((m) => m.status === 'OK').length,
    ALERT: all.value.filter((m) => m.status === 'ALERT').length,
    CRITICAL: all.value.filter((m) => m.status === 'CRITICAL').length,
  }))

  const query = computed<MeterQuery>(() => ({
    status: filter.value === 'ALL' ? undefined : filter.value,
    search: search.value.trim() || undefined,
    sort: sort.value,
    order: order.value,
  }))

  // Only the latest request may update the table: a slow response must not overwrite a newer one.
  let seq = 0

  async function fetchRows() {
    const mine = ++seq
    try {
      const data = await getMeters(query.value)
      if (mine !== seq) return
      rows.value = data
      error.value = null
    } catch (e) {
      if (mine !== seq) return
      error.value = errorMessage(e, 'No se pudieron cargar los medidores')
    }
  }

  /** Loads the counters and the table. */
  async function load() {
    loading.value = true
    try {
      const [everything] = await Promise.all([getMeters().catch(() => null), fetchRows()])
      if (everything) all.value = everything
    } finally {
      loading.value = false
    }
  }

  function setFilter(value: MeterFilter) {
    filter.value = value
    return fetchRows()
  }

  function setSearch(value: string) {
    search.value = value
    return fetchRows()
  }

  /** Clicking the active column flips the direction; a new column starts descending. */
  function toggleSort(key: MeterSort) {
    if (sort.value === key) order.value = order.value === 'desc' ? 'asc' : 'desc'
    else {
      sort.value = key
      order.value = 'desc'
    }
    return fetchRows()
  }

  return {
    all,
    rows,
    filter,
    search,
    sort,
    order,
    loading,
    error,
    counts,
    query,
    load,
    setFilter,
    setSearch,
    toggleSort,
  }
})
