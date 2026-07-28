<script setup lang="ts">
import type { AcademicYear } from '~/types/api'

definePageMeta({ layout: 'dashboard' })
useSeoMeta({ title: 'Create academic year' })

const toast = useToast()
const { postItem } = useApi()
const { academicYears, selectedSchoolId, selectedYearId } = useSchoolContext()
const saving = ref(false)

async function save(year: AcademicYear) {
  if (saving.value) return
  if (!selectedSchoolId.value) {
    toast.add({ title: 'Select a school first', color: 'error' })
    return
  }
  saving.value = true
  try {
    const response = await postItem<AcademicYear>('/academic-years', {
      schoolId: selectedSchoolId.value,
      name: year.name,
      startDate: year.startDate,
      isCurrent: true,
      segments: year.segments.map(segment => ({
        name: segment.name,
        type: segment.type,
        durationDays: Number(segment.durationDays)
      }))
    })
    academicYears.value = academicYears.value.map(item => ({ ...item, isCurrent: false }))
    academicYears.value.push(response.data)
    selectedYearId.value = response.data.id
    toast.add({ title: 'Academic year created', description: `${response.data.name} is now active.`, color: 'success' })
    await navigateTo(`/dashboard/settings/academic-years/${response.data.id}`)
  } catch (error) {
    toast.add({ title: 'Could not create academic year', description: apiErrorMessage(error, 'Please try again.'), color: 'error' })
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <SettingsShell>
    <PageHeading
      title="Create academic year"
      description="Build a clear calendar from terms and vacations."
      :eyebrow="$route.query.onboarding === '1' ? 'Setup · Step 2 of 2' : 'Academic years'"
    />
    <div class="mt-6">
      <AcademicYearForm @save="save" />
    </div>
  </SettingsShell>
</template>
