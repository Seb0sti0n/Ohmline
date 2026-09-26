<script setup lang="ts">
import type { Severity } from '@/types/api'
import { SEVERITY_BADGE, SEVERITY_LABEL } from '@/utils/labels'

defineProps<{ severity: Severity; long?: boolean }>()
</script>

<template>
  <span
    class="inline-flex h-6 items-center gap-1.5 whitespace-nowrap rounded-pill px-2.5 text-caption"
    :class="SEVERITY_BADGE[severity]"
  >
    <span class="size-[7px] rounded-full bg-current" />
    {{ long ? `Severidad ${SEVERITY_LABEL[severity].toLowerCase()}` : SEVERITY_LABEL[severity] }}
  </span>
</template>
