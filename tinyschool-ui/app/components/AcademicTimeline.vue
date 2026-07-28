<script setup lang="ts">
import type { AcademicSegment } from '~/types/api'

const props = defineProps<{
  startDate: string
  segments: AcademicSegment[]
}>()

const total = computed(() => Math.max(1, props.segments.reduce((sum, segment) => sum + Number(segment.durationDays || 0), 0)))

function width(days: number) {
  return `${Math.max(4, (days / total.value) * 100)}%`
}
</script>

<template>
  <div>
    <div class="flex min-h-12 w-full overflow-hidden rounded-lg border border-default bg-elevated">
      <div
        v-for="(segment, index) in segments"
        :key="`${segment.name}-${index}`"
        class="flex min-w-0 items-center justify-center px-2 text-center text-xs font-semibold"
        :class="segment.type === 'vacation' ? 'bg-amber-100 text-amber-800 dark:bg-amber-950 dark:text-amber-200' : 'bg-primary/12 text-primary'"
        :style="{ width: width(segment.durationDays) }"
        :title="`${segment.name}: ${segment.durationDays} days`"
      >
        <span class="truncate">{{ segment.name }}</span>
      </div>
      <div
        v-if="!segments.length"
        class="grid w-full place-items-center text-xs text-muted"
      >
        No segments yet
      </div>
    </div>
    <div class="mt-2 flex justify-between text-xs text-muted">
      <span>{{ startDate || 'Choose a start date' }}</span>
      <span>{{ total }} days</span>
    </div>
  </div>
</template>
