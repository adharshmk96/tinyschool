<script setup lang="ts">
import type { AcademicYear } from '~/types/api'

definePageMeta({ layout: 'dashboard' })
useSeoMeta({ title: 'Create academic year' })

const toast = useToast()
const { academicYears, selectedYearId } = useSchoolContext()

async function save(year: AcademicYear) {
  academicYears.value.push(year)
  selectedYearId.value = year.id
  toast.add({ title: 'Academic year created', description: `${year.name} is now active.`, color: 'success' })
  await navigateTo(`/dashboard/settings/academic-years/${year.id}`)
}
</script>

<template>
  <SettingsShell>
    <PageHeading
      title="Create academic year"
      description="Build a clear calendar from terms and vacations."
      eyebrow="Academic years"
    />
    <div class="mt-6">
      <AcademicYearForm @save="save" />
    </div>
  </SettingsShell>
</template>
