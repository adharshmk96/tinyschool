<script setup lang="ts">
const props = defineProps<{
  subject?: string
  data?: unknown
}>()

const performance = computed(() => {
  if (!props.data || typeof props.data !== 'object')
    return null

  const value = props.data as Record<string, unknown>
  return {
    averageScore: Number(value.averageScore ?? 0),
    classAverage: Number(value.classAverage ?? 0),
    completionRate: Number(value.completionRate ?? 0),
    completed: Number(value.completed ?? 0),
    total: Number(value.total ?? 0),
    standing: String(value.standing ?? 'Not ranked'),
    trend: Array.isArray(value.trend) ? value.trend.map(Number).filter(Number.isFinite) : []
  }
})
</script>

<template>
  <TabPlaceholder
    v-if="!performance"
    icon="i-lucide-chart-no-axes-combined"
    title="No performance data available"
    description="Scores and completion insights will appear here after work has been marked."
  />
  <section v-else>
    <div class="mb-5 flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between">
      <div>
        <div class="flex items-center gap-2">
          <span class="grid size-9 place-items-center rounded-lg bg-primary/10 text-primary">
            <UIcon
              name="i-lucide-chart-no-axes-combined"
              class="size-4"
            />
          </span>
          <h2 class="text-lg font-semibold text-highlighted">
            Performance overview
          </h2>
        </div>
        <p class="mt-2 text-sm leading-6 text-muted">
          Current-year scores, completion and progress over time.
        </p>
      </div>
      <UBadge
        label="Current year"
        color="neutral"
        variant="subtle"
        icon="i-lucide-calendar-check"
        class="self-start"
      />
    </div>

    <div class="grid gap-5 lg:grid-cols-3">
      <UCard>
        <div class="flex items-start justify-between gap-3">
          <div>
            <p class="text-sm text-muted">
              Average score
            </p>
            <p class="mt-2 text-3xl font-semibold text-highlighted">
              {{ performance.averageScore }}%
            </p>
          </div>
          <UIcon
            name="i-lucide-gauge"
            class="size-5 text-primary"
          />
        </div>
        <UBadge
          class="mt-3"
          color="success"
          variant="subtle"
          :label="performance.classAverage ? `Class average ${performance.classAverage}%` : 'Current average'"
        />
      </UCard>
      <UCard>
        <p class="text-sm text-muted">
          {{ subject || 'Work' }} completed
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.completed }} / {{ performance.total }}
        </p>
        <UProgress
          :model-value="performance.completionRate"
          class="mt-4"
        />
      </UCard>
      <UCard>
        <p class="text-sm text-muted">
          Class standing
        </p>
        <p class="mt-2 text-3xl font-semibold text-highlighted">
          {{ performance.standing }}
        </p>
        <p class="mt-3 text-sm text-muted">
          Based on marked work from the current academic year.
        </p>
      </UCard>
      <UCard class="lg:col-span-3">
        <div
          v-if="performance.trend.length"
          class="flex h-40 items-end gap-3 pt-5"
        >
          <div
            v-for="(height, index) in performance.trend"
            :key="index"
            class="flex-1 rounded-t bg-primary/80"
            :style="{ height: `${height}%` }"
            :title="`${height}%`"
          />
        </div>
        <div
          v-if="performance.trend.length"
          class="mt-3 flex justify-between text-xs text-muted"
        >
          <span>Term start</span><span>Current</span>
        </div>
        <TabPlaceholder
          v-else
          icon="i-lucide-chart-spline"
          title="No score trend yet"
          description="The trend chart will fill in as more scores are recorded."
          badge="Waiting for scores"
        />
      </UCard>
    </div>
  </section>
</template>
