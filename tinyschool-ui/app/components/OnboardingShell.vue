<script setup lang="ts">
defineProps<{
  steps: ReadonlyArray<{ key: string, label: string, description: string, icon: string }>
  activeIndex: number
  title: string
  description: string
}>()
</script>

<template>
  <div class="min-h-screen bg-elevated/40">
    <header class="flex h-16 items-center justify-between px-6 lg:px-10">
      <AppLogo />
      <UColorModeButton />
    </header>

    <main class="mx-auto w-full max-w-3xl px-6 pb-16">
      <ol class="mb-8 flex flex-col gap-3 sm:flex-row sm:items-center">
        <li
          v-for="(step, index) in steps"
          :key="step.key"
          class="flex flex-1 items-center gap-3"
        >
          <div
            class="grid size-9 shrink-0 place-items-center rounded-full border text-sm font-semibold"
            :class="index < activeIndex
              ? 'border-primary bg-primary text-inverted'
              : index === activeIndex
                ? 'border-primary text-primary'
                : 'border-default text-dimmed'"
          >
            <UIcon
              :name="index < activeIndex ? 'i-lucide-check' : step.icon"
              class="size-4"
            />
          </div>
          <div class="min-w-0">
            <p
              class="truncate text-sm font-semibold"
              :class="index <= activeIndex ? 'text-highlighted' : 'text-dimmed'"
            >
              {{ step.label }}
            </p>
            <p class="truncate text-xs text-muted">
              {{ step.description }}
            </p>
          </div>
          <div
            v-if="index < steps.length - 1"
            class="hidden h-px flex-1 bg-accented sm:block"
          />
        </li>
      </ol>

      <UCard>
        <template #header>
          <p class="mb-1 text-xs font-semibold uppercase tracking-widest text-primary">
            Step {{ activeIndex + 1 }} of {{ steps.length }}
          </p>
          <h1 class="text-xl font-bold tracking-tight text-highlighted sm:text-2xl">
            {{ title }}
          </h1>
          <p class="mt-1 text-sm text-muted">
            {{ description }}
          </p>
        </template>

        <slot />

        <template
          v-if="$slots.footer"
          #footer
        >
          <div class="flex w-full flex-wrap items-center justify-end gap-2">
            <slot name="footer" />
          </div>
        </template>
      </UCard>
    </main>
  </div>
</template>
