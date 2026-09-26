<script setup lang="ts">
import { computed } from 'vue'
import { splitRule } from '@/utils/anomalies'

const props = defineProps<{ rules: string[] }>()
const items = computed(() => props.rules.map(splitRule))
</script>

<template>
  <section class="flex flex-col rounded-lg border border-line bg-surface px-6 pb-2 pt-[18px]">
    <h2 class="mb-1.5 text-heading">Evidencia</h2>
    <ul class="m-0 list-none p-0">
      <li v-for="item in items" :key="item.title" class="flex gap-3 border-t border-line py-3">
        <span class="flex size-6 shrink-0 items-center justify-center rounded-full bg-brand-soft">
          <svg
            width="14"
            height="14"
            viewBox="0 0 24 24"
            fill="none"
            stroke="var(--color-brand)"
            stroke-width="2.6"
            stroke-linecap="round"
            stroke-linejoin="round"
            aria-hidden="true"
          >
            <path d="m5 12 5 5 9-10" />
          </svg>
        </span>
        <span class="flex flex-col gap-[3px]">
          <span class="text-body font-[650]">{{ item.title }}</span>
          <span v-if="item.detail" class="text-label font-normal leading-[1.45] text-ink-muted">{{
            item.detail
          }}</span>
        </span>
      </li>
    </ul>
  </section>
</template>
