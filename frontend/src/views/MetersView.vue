<script setup lang="ts">
import { onBeforeUnmount, onMounted, ref } from 'vue'
import ErrorState from '@/components/ErrorState.vue'
import MetersTable from '@/components/MetersTable.vue'
import PaginationBar from '@/components/PaginationBar.vue'
import { useMetersStore } from '@/stores/meters'
import type { MeterFilter } from '@/types/api'

const meters = useMetersStore()

const filters: { key: MeterFilter; label: string }[] = [
  { key: 'ALL', label: 'Todos' },
  { key: 'OK', label: 'Normales' },
  { key: 'ALERT', label: 'Alertas' },
  { key: 'CRITICAL', label: 'Críticos' },
]

const searchText = ref(meters.search)
let searchTimer: ReturnType<typeof setTimeout> | undefined

// Wait for a pause in typing before asking the API.
function onSearch() {
  clearTimeout(searchTimer)
  searchTimer = setTimeout(() => void meters.setSearch(searchText.value), 250)
}

// Before the first analysis no meter has a status, so the status filters have nothing to filter.
const statusLocked = (key: MeterFilter) => key !== 'ALL' && meters.unevaluated

onMounted(() => void meters.load())
onBeforeUnmount(() => clearTimeout(searchTimer))
</script>

<template>
  <header class="flex flex-col gap-1.5">
    <h1 class="text-display tracking-[-0.01em] [font-stretch:112%]">Medidores</h1>
    <p class="text-body text-ink-muted">
      Consumo del último día frente a su baseline horario de la primera semana.
    </p>
  </header>

  <div class="flex flex-wrap items-center justify-between gap-4">
    <div role="group" aria-label="Filtrar por estado" class="flex gap-2">
      <button
        v-for="f in filters"
        :key="f.key"
        type="button"
        :aria-pressed="meters.filter === f.key"
        :disabled="statusLocked(f.key)"
        :title="
          statusLocked(f.key) ? 'Ejecuta el análisis IA para poder filtrar por estado' : undefined
        "
        class="inline-flex h-10 cursor-pointer items-center gap-2 rounded-full border px-4 text-[15px] font-semibold leading-5 disabled:cursor-not-allowed disabled:opacity-50"
        :class="
          meters.filter === f.key
            ? 'border-ink bg-ink text-white'
            : 'border-line bg-surface text-ink hover:bg-surface-muted'
        "
        @click="meters.setFilter(f.key)"
      >
        {{ f.label
        }}<span v-if="meters.counts.ALL && !statusLocked(f.key)" class="font-medium opacity-80">{{
          meters.counts[f.key]
        }}</span>
      </button>
    </div>

    <label
      class="search-field flex h-11 w-[300px] items-center gap-2.5 rounded-md border border-line bg-surface px-3.5 text-ink-muted"
    >
      <svg
        width="18"
        height="18"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        stroke-width="1.8"
        stroke-linecap="round"
        stroke-linejoin="round"
        aria-hidden="true"
      >
        <circle cx="11" cy="11" r="7" />
        <path d="m20 20-4-4" />
      </svg>
      <span class="sr-only">Buscar por código de medidor</span>
      <input
        v-model="searchText"
        type="search"
        placeholder="Buscar por código de medidor"
        class="h-full min-w-0 grow border-0 bg-transparent p-0 text-[15px] leading-5 text-ink outline-none"
        @input="onSearch"
      />
    </label>
  </div>

  <p v-if="meters.unevaluated" class="text-label font-normal text-ink-muted">
    Los medidores aún no se han evaluado.
    <RouterLink to="/" class="font-semibold text-brand no-underline hover:text-brand-strong"
      >Ejecuta el análisis IA</RouterLink
    >
    para ver el estado de cada uno.
  </p>

  <ErrorState
    v-if="meters.error && meters.rows.length === 0"
    :message="meters.error"
    @retry="meters.load()"
  />
  <p v-else-if="meters.loading && meters.rows.length === 0" class="text-body text-ink-muted">
    Cargando medidores…
  </p>

  <section v-else class="overflow-x-auto rounded-lg border border-line bg-surface">
    <MetersTable
      :rows="meters.rows"
      :sort="meters.sort"
      :order="meters.order"
      @sort="meters.toggleSort"
    />
    <p v-if="meters.rows.length === 0" class="px-7 py-7 text-body text-ink-muted">
      Ningún medidor coincide con la búsqueda. Revisa el código o cambia el filtro.
    </p>
    <PaginationBar
      v-if="meters.total > 0"
      :page="meters.page"
      :page-count="meters.pageCount"
      :page-size="meters.pageSize"
      :total="meters.total"
      @page="meters.setPage"
      @size="meters.setPageSize"
    />
  </section>
</template>

<style scoped>
/* The visible search box is this whole pill (icon + input), not just the <input> inside it, so the
   focus ring belongs on the pill. Without this the browser's default focus outline draws around the
   bare input only, which looks like a stray, partial highlight instead of one field lighting up. */
.search-field:focus-within {
  outline: 2px solid var(--color-brand);
  outline-offset: 2px;
}
.search-field input:focus-visible {
  outline: none;
}
</style>
