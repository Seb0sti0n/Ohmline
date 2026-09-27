<script setup lang="ts">
import { computed } from 'vue'
import { PAGE_SIZES, pageWindow, rangeLabel } from '@/utils/pagination'

const props = defineProps<{ page: number; pageCount: number; pageSize: number; total: number }>()
const emit = defineEmits<{ page: [value: number]; size: [value: number] }>()

const pages = computed(() => pageWindow(props.page, props.pageCount))
const range = computed(() => rangeLabel(props.page, props.pageSize, props.total))

const control =
  'inline-flex h-10 min-w-10 cursor-pointer items-center justify-center rounded-md border px-3 text-[15px] leading-5 font-semibold'
</script>

<template>
  <div class="flex flex-wrap items-center justify-between gap-3 border-t border-line px-4 py-3">
    <p class="text-label font-normal text-ink-muted" aria-live="polite">
      Mostrando <span class="font-semibold text-ink tabular-nums">{{ range }}</span> medidores
    </p>

    <div class="flex flex-wrap items-center gap-4">
      <label class="flex items-center gap-2 text-label font-normal text-ink-muted">
        Filas por página
        <select
          :value="pageSize"
          class="h-10 cursor-pointer rounded-md border border-line bg-surface px-2 text-[15px] leading-5 font-semibold text-ink"
          @change="emit('size', Number(($event.target as HTMLSelectElement).value))"
        >
          <option v-for="n in PAGE_SIZES" :key="n" :value="n">{{ n }}</option>
        </select>
      </label>

      <nav v-if="pageCount > 1" aria-label="Paginación" class="flex items-center gap-1.5">
        <button
          type="button"
          :class="[
            control,
            'border-line bg-surface text-ink hover:bg-surface-muted disabled:cursor-not-allowed disabled:opacity-50',
          ]"
          :disabled="page <= 1"
          aria-label="Página anterior"
          @click="emit('page', page - 1)"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <path d="M15 5l-7 7 7 7" />
          </svg>
        </button>
        <template v-for="(p, i) in pages" :key="`${p}-${i}`">
          <span v-if="p === '…'" class="px-1 text-ink-muted" aria-hidden="true">…</span>
          <button
            v-else
            type="button"
            :class="[
              control,
              p === page
                ? 'border-ink bg-ink text-white'
                : 'border-line bg-surface text-ink hover:bg-surface-muted',
            ]"
            :aria-current="p === page ? 'page' : undefined"
            :aria-label="`Página ${p}`"
            @click="emit('page', p)"
          >
            {{ p }}
          </button>
        </template>
        <button
          type="button"
          :class="[
            control,
            'border-line bg-surface text-ink hover:bg-surface-muted disabled:cursor-not-allowed disabled:opacity-50',
          ]"
          :disabled="page >= pageCount"
          aria-label="Página siguiente"
          @click="emit('page', page + 1)"
        >
          <svg
            width="16"
            height="16"
            viewBox="0 0 24 24"
            fill="none"
            stroke="currentColor"
            stroke-width="2"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <path d="M9 5l7 7-7 7" />
          </svg>
        </button>
      </nav>
    </div>
  </div>
</template>
