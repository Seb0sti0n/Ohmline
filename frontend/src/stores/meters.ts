import { computed, ref } from 'vue'
import { defineStore } from 'pinia'
import { errorMessage, getMeters } from '@/services/api'
import type {
  MeterCounts,
  MeterFilter,
  MeterQuery,
  MeterSort,
  MeterSummary,
  SortOrder,
} from '@/types/api'
import { DEFAULT_PAGE_SIZE, pageCount as countPages } from '@/utils/pagination'

const NO_METERS: MeterCounts = { ALL: 0, OK: 0, ALERT: 0, CRITICAL: 0, UNEVALUATED: 0 }

export const useMetersStore = defineStore('meters', () => {
  /** The rows of the current page, already filtered, searched and sorted by the API. */
  const rows = ref<MeterSummary[]>([])
  /** How many meters match the filters (all pages). */
  const total = ref(0)
  /** Meters per status over all meters, for the filter chips. */
  const counts = ref<MeterCounts>({ ...NO_METERS })
  const page = ref(1)
  const pageSize = ref(DEFAULT_PAGE_SIZE)
  const filter = ref<MeterFilter>('ALL')
  const search = ref('')
  const sort = ref<MeterSort>('severity')
  const order = ref<SortOrder>('desc')
  const loading = ref(false)
  const error = ref<string | null>(null)

  const pageCount = computed(() => countPages(total.value, pageSize.value))
  /** True until the first analysis completes: no meter has a status to filter by yet. */
  const unevaluated = computed(
    () => counts.value.ALL > 0 && counts.value.UNEVALUATED === counts.value.ALL,
  )

  const query = computed<MeterQuery>(() => ({
    status: filter.value === 'ALL' ? undefined : filter.value,
    search: search.value.trim() || undefined,
    sort: sort.value,
    order: order.value,
    page: page.value,
    page_size: pageSize.value,
  }))

  // Only the latest request may update the table: a slow response must not overwrite a newer one.
  let seq = 0

  async function fetchRows(): Promise<void> {
    const mine = ++seq
    try {
      const data = await getMeters(query.value)
      if (mine !== seq) return
      counts.value = data.counts
      total.value = data.total
      // The page we asked for no longer exists (fewer meters than before): go to the last one.
      if (data.items.length === 0 && data.total > 0 && page.value > 1) {
        page.value = countPages(data.total, pageSize.value)
        return fetchRows()
      }
      rows.value = data.items
      error.value = null
    } catch (e) {
      if (mine !== seq) return
      error.value = errorMessage(e, 'No se pudieron cargar los medidores')
    }
  }

  async function load() {
    loading.value = true
    try {
      await fetchRows()
      // A status filter left over from before (e.g. the data was reset) would show an empty table.
      if (unevaluated.value && filter.value !== 'ALL') {
        filter.value = 'ALL'
        page.value = 1
        await fetchRows()
      }
    } finally {
      loading.value = false
    }
  }

  // Changing what is shown starts again from the first page.
  function setFilter(value: MeterFilter) {
    filter.value = value
    page.value = 1
    return fetchRows()
  }

  function setSearch(value: string) {
    search.value = value
    page.value = 1
    return fetchRows()
  }

  /** Clicking the active column flips the direction; a new column starts descending. */
  function toggleSort(key: MeterSort) {
    if (sort.value === key) order.value = order.value === 'desc' ? 'asc' : 'desc'
    else {
      sort.value = key
      order.value = 'desc'
    }
    page.value = 1
    return fetchRows()
  }

  function setPage(value: number) {
    const next = Math.min(Math.max(1, Math.trunc(value)), pageCount.value)
    if (next === page.value) return Promise.resolve()
    page.value = next
    return fetchRows()
  }

  function setPageSize(value: number) {
    pageSize.value = value
    page.value = 1
    return fetchRows()
  }

  return {
    rows,
    total,
    counts,
    page,
    pageSize,
    pageCount,
    filter,
    search,
    sort,
    order,
    loading,
    error,
    unevaluated,
    query,
    load,
    setFilter,
    setSearch,
    toggleSort,
    setPage,
    setPageSize,
  }
})
