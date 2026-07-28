<script setup lang="ts">
definePageMeta({ layout: 'dashboard' })
const fields = computed(() => [
  { key: 'firstName', label: 'First name' },
  { key: 'lastName', label: 'Last name' },
  { key: 'email', label: 'Email', type: 'email' },
  { key: 'phone', label: 'Phone', type: 'tel' },
  { key: 'grades', label: 'Grade by academic year', type: 'grades' },
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
    description="Students stay across academic years — only their grade changes per year."
    :sort-options="[
      { label: 'Name', value: 'name' },
      { label: 'Grade', value: 'grade' },
      { label: 'Average score', value: 'averageScore' }
    ]"
    :card-fields="[
      { key: 'email', label: 'Email' },
      { key: 'phone', label: 'Phone' },
      { key: 'grade', label: 'Grade (current year)' },
      { key: 'guardian.name', label: 'Guardian' }
    ]"
    :fields="fields"
  />
</template>
