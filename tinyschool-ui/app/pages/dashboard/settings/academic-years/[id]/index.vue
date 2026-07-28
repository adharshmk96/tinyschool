<script setup lang="ts">
import type { AcademicYear } from '~/types/api'

definePageMeta({ layout: 'dashboard' })

const route = useRoute()
const { getItem } = useApi()
const id = computed(() => String(route.params.id))
const { academicYears } = useSchoolContext()

const { data: fetched, status, error } = await useAsyncData(`academic-year-${id.value}`, async () => {
  const local = academicYears.value.find(item => item.id === id.value)
  if (local) return local
  return (await getItem<AcademicYear>(`/academic-years/${id.value}`)).data
})

const year = computed(() => fetched.value || academicYears.value.find(item => item.id === id.value))
useSeoMeta({ title: () => year.value?.name || 'Academic year' })

function endDate(startDate: string, duration: number) {
  const date = new Date(`${startDate}T00:00:00`)
  date.setDate(date.getDate() + Math.max(0, duration - 1))
  return new Intl.DateTimeFormat('en', { dateStyle: 'medium' }).format(date)
}
</script>

<template>
  <SettingsShell>
    <div
      v-if="status === 'pending'"
      class="space-y-4"
    >
      <USkeleton class="h-12 w-72" />
      <USkeleton class="h-72 rounded-xl" />
    </div>
    <UAlert
      v-else-if="error || !year"
      color="error"
      variant="subtle"
      icon="i-lucide-circle-alert"
      title="Academic year not found"
      description="Return to the academic years list and choose another calendar."
    >
      <template #actions>
        <UButton
          to="/dashboard/settings/academic-years"
          label="Back to academic years"
          color="neutral"
          variant="outline"
        />
      </template>
    </UAlert>
    <template v-else>
      <PageHeading
        :title="year.name"
        :description="`${year.durationDays} days from ${year.startDate} to ${endDate(year.startDate, year.durationDays)}`"
        eyebrow="Academic year"
      >
        <template #actions>
          <UButton
            :to="`/dashboard/settings/academic-years/${year.id}/edit`"
            icon="i-lucide-pencil"
            color="neutral"
            variant="outline"
            label="Edit"
          />
        </template>
      </PageHeading>

      <div class="mt-6 space-y-6">
        <UCard>
          <template #header>
            <h2 class="font-semibold">
              Year timeline
            </h2><p class="text-sm text-muted">
              Terms and breaks shown in proportion to their duration.
            </p>
          </template>
          <AcademicTimeline
            :start-date="year.startDate"
            :segments="year.segments || []"
          />
        </UCard>

        <UCard>
          <template #header>
            <h2 class="font-semibold">
              Segments
            </h2><p class="text-sm text-muted">
              {{ year.segments?.length || 0 }} parts in this academic year.
            </p>
          </template>
          <div class="divide-y divide-default">
            <div
              v-for="(segment, index) in year.segments"
              :key="segment.id || index"
              class="flex items-center gap-3 py-4 first:pt-0 last:pb-0"
            >
              <span
                class="grid size-9 place-items-center rounded-lg"
                :class="segment.type === 'vacation' ? 'bg-amber-100 text-amber-700 dark:bg-amber-950 dark:text-amber-300' : 'bg-primary/10 text-primary'"
              >
                <UIcon
                  :name="segment.type === 'vacation' ? 'i-lucide-palmtree' : 'i-lucide-book-open'"
                  class="size-4"
                />
              </span>
              <div class="min-w-0 flex-1">
                <p class="truncate text-sm font-semibold">
                  {{ segment.name }}
                </p><p class="text-xs capitalize text-muted">
                  {{ segment.type }}
                </p>
              </div>
              <p class="text-sm tabular-nums text-muted">
                {{ segment.durationDays }} days
              </p>
            </div>
          </div>
        </UCard>
      </div>
    </template>
  </SettingsShell>
</template>
