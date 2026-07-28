<script setup lang="ts">
defineProps<{
  icon: string
  title: string
  description: string
}>()
</script>

<template>
  <div class="flex min-h-64 flex-col items-center justify-center rounded-xl border border-dashed border-default p-8 text-center">
    <div class="mb-4 grid size-12 place-items-center rounded-full bg-elevated">
      <UIcon
        :name="icon"
        class="size-6 text-muted"
      />
    </div>
    <h2 class="font-semibold">
      {{ title }}
    </h2>
    <p class="mt-1 max-w-md text-sm text-muted">
      {{ description }}
    </p>
    <div
      v-if="$slots.default"
      class="mt-5"
    >
      <slot />
    </div>
  </div>
</template>
