<script setup lang="ts">
withDefaults(defineProps<{
  icon?: string
  title: string
  description: string
  badge?: string
}>(), {
  icon: 'i-lucide-construction',
  badge: 'Preview'
})
</script>

<template>
  <UCard class="overflow-hidden">
    <div class="flex min-h-56 flex-col items-center justify-center px-5 py-10 text-center sm:px-8">
      <div class="grid size-12 place-items-center rounded-xl bg-primary/10 text-primary ring-1 ring-primary/15">
        <UIcon
          :name="icon"
          class="size-6"
        />
      </div>
      <UBadge
        :label="badge"
        color="neutral"
        variant="subtle"
        class="mt-5"
      />
      <h2 class="mt-3 text-lg font-semibold text-highlighted">
        {{ title }}
      </h2>
      <p class="mt-2 max-w-lg text-sm leading-6 text-muted">
        {{ description }}
      </p>
      <slot />
    </div>
  </UCard>
</template>
