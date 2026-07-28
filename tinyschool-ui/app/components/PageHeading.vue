<script setup lang="ts">
defineProps<{
  title: string
  description?: string
  eyebrow?: string
}>()
</script>

<template>
  <header class="flex flex-col gap-5 sm:flex-row sm:items-end sm:justify-between">
    <div class="min-w-0">
      <p
        v-if="eyebrow"
        class="mb-1 text-xs font-semibold uppercase tracking-widest text-primary"
      >
        {{ eyebrow }}
      </p>
      <h1 class="text-2xl font-bold tracking-tight text-highlighted sm:text-3xl">
        {{ title }}
      </h1>
      <p
        v-if="description"
        class="mt-2 max-w-2xl text-sm leading-6 text-muted sm:text-base"
      >
        {{ description }}
      </p>
    </div>
    <div
      v-if="$slots.actions"
      class="flex shrink-0 flex-wrap gap-2"
    >
      <slot name="actions" />
    </div>
  </header>
</template>
