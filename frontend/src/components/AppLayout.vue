<script setup lang="ts">
import AppSidebar from './AppSidebar.vue'
</script>

<template>
  <div class="flex min-h-screen bg-ground">
    <AppSidebar />
    <main class="flex min-w-0 grow flex-col gap-6 px-10 py-8">
      <RouterView />
    </main>
  </div>
</template>
