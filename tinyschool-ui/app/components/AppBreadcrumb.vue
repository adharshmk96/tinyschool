<script setup lang="ts">
defineProps<{
  items: Array<{ label: string, to?: string, icon?: string }>
}>()
</script>

<template>
  <nav
    class="mb-5"
    aria-label="Breadcrumb"
  >
    <UBreadcrumb
      :items="items"
      :ui="{ link: 'text-xs sm:text-sm', separatorIcon: 'size-3.5' }"
    />
  </nav>
</template>
