<script setup lang="ts">
definePageMeta({ layout: 'dashboard' })
useSeoMeta({ title: 'Data settings' })

const toast = useToast()
const { request } = useApi()
const { schools, academicYears, selectedSchoolId, selectedYearId } = useSchoolContext()
const confirmationOpen = ref(false)
const pending = ref(false)

async function clearData() {
  pending.value = true
  try {
    await request('/me/data', { method: 'DELETE' })
    schools.value = []
    academicYears.value = []
    selectedSchoolId.value = undefined
    selectedYearId.value = undefined
    toast.add({
      title: 'Data cleared',
      description: 'Your schools and all related records were deleted.',
      color: 'success'
    })
    await navigateTo('/dashboard/settings/schools?create=1&onboarding=1')
  } catch (error) {
    toast.add({
      title: 'Could not clear data',
      description: apiErrorMessage(error, 'Your data was not changed. Try again.'),
      color: 'error'
    })
  } finally {
    pending.value = false
  }
}
</script>

<template>
  <SettingsShell>
    <UCard>
      <template #header>
        <h2 class="font-semibold">
          Clear data
        </h2>
        <p class="text-sm text-muted">
          Delete your schools, academic years, students, classes, assignments, exams and related records.
        </p>
      </template>

      <UAlert
        class="mb-5"
        color="error"
        variant="subtle"
        icon="i-lucide-triangle-alert"
        title="This cannot be undone"
        description="Your account and sign-in details will remain."
      />
      <UButton
        color="error"
        icon="i-lucide-trash-2"
        label="Clear data"
        :loading="pending"
        @click="confirmationOpen = true"
      />
    </UCard>

    <ConfirmDialog
      v-model="confirmationOpen"
      title="Clear all your data?"
      description="This permanently deletes all workspace data owned by your account. Other users are not affected."
      confirm-label="Clear data"
      @confirm="clearData"
    />
  </SettingsShell>
</template>
