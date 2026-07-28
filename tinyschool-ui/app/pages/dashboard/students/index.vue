<script setup lang="ts">
definePageMeta({ layout: 'dashboard' })
const { selectedSchool } = useSchoolContext()
const fields = computed(() => [
  { key: 'firstName', label: 'First name' },
  { key: 'lastName', label: 'Last name' },
  { key: 'email', label: 'Email', type: 'email' },
  { key: 'phone', label: 'Phone', type: 'tel' },
  { key: 'grade', label: 'Grade', options: selectedSchool.value?.grades || [] },
  { key: 'guardianName', label: 'Guardian name' },
  { key: 'guardianEmail', label: 'Guardian email', type: 'email' },
  { key: 'guardianPhone', label: 'Guardian phone', type: 'tel' },
  { key: 'residentAddress', label: 'Resident address', type: 'textarea' },
  { key: 'permanentAddress', label: 'Permanent address', type: 'textarea' }
])
</script>

<template>
  <DomainListPage
    title="Students"
    singular="Student"
    endpoint="/api/v1/students"
    icon="i-lucide-graduation-cap"
    detail-base="/dashboard/students"
    description="Manage student profiles, guardians, grades and contact details."
    :sort-options="[
      { label: 'Name', value: 'name' },
      { label: 'Grade', value: 'grade' },
      { label: 'Average score', value: 'averageScore' }
    ]"
    :card-fields="[
      { key: 'email', label: 'Email' },
      { key: 'phone', label: 'Phone' },
      { key: 'grade', label: 'Grade' },
      { key: 'guardian.name', label: 'Guardian' }
    ]"
    :fields="fields"
  />
</template>
