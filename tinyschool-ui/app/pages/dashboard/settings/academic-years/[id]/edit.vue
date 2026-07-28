<script setup lang="ts">
import type { AcademicYear } from '~/types/api'

definePageMeta({ layout: 'dashboard' })

const route = useRoute()
const toast = useToast()
const { getItem } = useApi()
const { academicYears } = useSchoolContext()
const id = String(route.params.id)
const { data: year, status, error } = await useAsyncData(`academic-year-edit-${id}`, async () => {
  const local = academicYears.value.find(item => item.id === id)
  if (local) return local
  return (await getItem<AcademicYear>(`/academic-years/${id}`)).data
})

useSeoMeta({ title: () => year.value ? `Edit ${year.value.name}` : 'Edit academic year' })

async function save(updated: AcademicYear) {
  const index = academicYears.value.findIndex(item => item.id === updated.id)
  if (index >= 0) academicYears.value[index] = updated
  year.value = updated
  toast.add({ title: 'Academic year updated', color: 'success' })
  await navigateTo(`/dashboard/settings/academic-years/${updated.id}`)
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
