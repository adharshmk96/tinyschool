<script setup lang="ts">
import type { AcademicYear } from '~/types/api'

definePageMeta({ layout: 'dashboard' })

const route = useRoute()
const toast = useToast()
const { getItem, patchItem } = useApi()
const { academicYears, selectedSchoolId, selectedYearId } = useSchoolContext()
const saving = ref(false)
const id = String(route.params.id)
const { data: year, status, error } = await useAsyncData(`academic-year-edit-${id}`, async () => {
  const local = academicYears.value.find(item => item.id === id)
  if (local) return local
  return (await getItem<AcademicYear>(`/academic-years/${id}`)).data
})

useSeoMeta({ title: () => year.value ? `Edit ${year.value.name}` : 'Edit academic year' })

async function save(updated: AcademicYear) {
  if (saving.value) return
  if (!selectedSchoolId.value) {
    toast.add({ title: 'Select a school first', color: 'error' })
    return
  }
  saving.value = true
  try {
    const schoolId = String((year.value as AcademicYear & { schoolId?: string }).schoolId || selectedSchoolId.value)
    const response = await patchItem<AcademicYear>(`/academic-years/${id}`, {
      schoolId,
      name: updated.name,
      startDate: updated.startDate,
      isCurrent: Boolean(year.value?.isCurrent),
      segments: updated.segments.map(segment => ({
        name: segment.name,
        type: segment.type,
        durationDays: Number(segment.durationDays)
      }))
    })
    const index = academicYears.value.findIndex(item => item.id === response.data.id)
    if (index >= 0) academicYears.value[index] = response.data
    year.value = response.data
    if (response.data.isCurrent) selectedYearId.value = response.data.id
    toast.add({ title: 'Academic year updated', color: 'success' })
    await navigateTo(`/dashboard/settings/academic-years/${response.data.id}`)
  } catch (error) {
    toast.add({ title: 'Could not update academic year', description: apiErrorMessage(error, 'Please try again.'), color: 'error' })
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <SettingsShell>
    <USkeleton
      v-if="status === 'pending'"
      class="h-96 rounded-xl"
    />
    <UAlert
      v-else-if="error || !year"
      color="error"
      variant="subtle"
      title="Academic year not found"
    />
    <template v-else>
      <PageHeading
        :title="`Edit ${year.name}`"
        description="Update the year details, terms and vacations."
        eyebrow="Academic years"
      />
      <div class="mt-6">
        <AcademicYearForm
          :initial="year"
          @save="save"
        />
      </div>
    </template>
  </SettingsShell>
</template>
