<script setup lang="ts">
defineProps<{
  label: string
  value: number
  icon: string
  to: string
  detail: string
}>()
</script>

<template>
  <NuxtLink
    :to="to"
    class="group rounded-xl border border-default bg-default p-5 shadow-xs transition hover:-translate-y-0.5 hover:border-primary/50 hover:shadow-md"
  >
    <div class="flex items-start justify-between">
      <span class="grid size-10 place-items-center rounded-lg bg-primary/10 text-primary">
        <UIcon
          :name="icon"
          class="size-5"
        />
      </span>
      <UIcon
        name="i-lucide-arrow-up-right"
        class="size-4 text-dimmed transition group-hover:text-primary"
      />
    </div>
    <p class="mt-5 text-3xl font-bold tabular-nums">{{ value }}</p>
    <p class="mt-1 font-medium">{{ label }}</p>
    <p class="mt-1 text-xs text-muted">{{ detail }}</p>
  </NuxtLink>
</template>
