<script setup lang="ts">
import type { AcademicYear } from '~/types/api'

definePageMeta({ layout: 'dashboard' })
useSeoMeta({ title: 'Academic years' })

const toast = useToast()
const { academicYears, selectedYearId, load } = useSchoolContext()
const deleteOpen = ref(false)
const deleting = ref<AcademicYear | null>(null)

function requestDelete(year: AcademicYear) {
  deleting.value = year
  deleteOpen.value = true
}

function remove() {
  if (!deleting.value) return
  academicYears.value = academicYears.value.filter(item => item.id !== deleting.value?.id)
  if (selectedYearId.value === deleting.value.id) {
    selectedYearId.value = academicYears.value.find(item => item.isCurrent)?.id || academicYears.value[0]?.id
  }
  toast.add({ title: 'Academic year deleted', color: 'success' })
}

onMounted(async () => {
  try {
    await load()
  } catch {
    toast.add({ title: 'Could not load academic years', color: 'error' })
  }
})
</script>

<template>
  <SettingsShell>
    <PageHeading
      title="Academic years"
      description="Plan terms, breaks and the calendar used across your school."
    >
      <template #actions>
        <UButton
          to="/dashboard/settings/academic-years/create"
          icon="i-lucide-plus"
          label="Create academic year"
        />
      </template>
    </PageHeading>

    <div class="mt-6 space-y-4">
      <NuxtLink
        v-for="year in academicYears"
        :key="year.id"
        :to="`/dashboard/settings/academic-years/${year.id}`"
        class="group block rounded-xl border border-default bg-default p-5 shadow-xs transition hover:border-primary/50 hover:shadow-md"
      >
        <div class="flex flex-col gap-5">
          <div class="flex items-start gap-3">
            <div class="grid size-10 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
              <UIcon
                name="i-lucide-calendar-range"
                class="size-5"
              />
            </div>
            <div class="min-w-0 flex-1">
              <div class="flex flex-wrap items-center gap-2">
                <h2 class="font-semibold">{{ year.name }}</h2>
                <UBadge
                  v-if="year.isCurrent || year.id === selectedYearId"
                  label="Current"
                  size="sm"
                  variant="subtle"
                />
              </div>
              <p class="mt-1 text-sm text-muted">Starts {{ year.startDate }} · {{ year.durationDays }} days</p>
            </div>
            <div class="flex shrink-0 gap-1">
              <UButton
                :to="`/dashboard/settings/academic-years/${year.id}/edit`"
                icon="i-lucide-pencil"
                color="neutral"
                variant="ghost"
                :aria-label="`Edit ${year.name}`"
                @click.stop
              />
              <UButton
                icon="i-lucide-trash-2"
                color="error"
                variant="ghost"
                :aria-label="`Delete ${year.name}`"
                @click.prevent.stop="requestDelete(year)"
              />
            </div>
          </div>
          <AcademicTimeline
            :start-date="year.startDate"
            :segments="year.segments || []"
          />
        </div>
      </NuxtLink>

      <EmptyState
        v-if="!academicYears.length"
        icon="i-lucide-calendar-plus"
        title="Create your first academic year"
        description="Add terms and vacations to start organizing students, classes and assessments."
      >
        <UButton
          to="/dashboard/settings/academic-years/create"
          icon="i-lucide-plus"
          label="Create academic year"
        />
      </EmptyState>
    </div>

    <ConfirmDialog
      v-model="deleteOpen"
      title="Delete academic year?"
      :description="`Delete ${deleting?.name || 'this academic year'} and its placeholder calendar?`"
      @confirm="remove"
    />
  </SettingsShell>
</template>
