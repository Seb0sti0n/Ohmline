<script setup lang="ts">
defineProps<{ message: string }>()
defineEmits<{ retry: [] }>()
</script>

<template>
  <div
    role="alert"
    class="flex flex-col items-start gap-3 rounded-lg border border-line bg-surface px-[26px] py-[22px]"
  >
    <p class="text-body font-semibold text-status-critical">{{ message }}</p>
    <button
      type="button"
      class="h-11 cursor-pointer rounded-md border border-line bg-surface px-4 text-[15px] font-semibold hover:bg-surface-muted"
      @click="$emit('retry')"
    >
      Reintentar
    </button>
  </div>
</template>
