<script setup lang="ts">
import { plural } from '@/utils/format'

defineProps<{ count: number }>()
</script>

<template>
  <p class="text-label font-normal text-ink-muted">
    <template v-if="count === 0"
      >Sin picos aislados: ninguna lectura se salió de la banda esperada por poco tiempo.</template
    >
    <template v-else>
      <span class="font-semibold text-ink tabular-nums">{{
        plural(count, 'pico aislado', 'picos aislados')
      }}</span
      >: lecturas que se salieron de la banda esperada durante pocas horas seguidas. Es ruido normal
      y no se trata como anomalía porque no es un cambio persistente.
    </template>
  </p>
</template>
