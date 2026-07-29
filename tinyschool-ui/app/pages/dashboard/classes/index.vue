<script setup lang="ts">
definePageMeta({ layout: 'dashboard' })
const { selectedSchool } = useSchoolContext()
const fields = computed(() => [
  { key: 'name', label: 'Class name' },
  { key: 'subject', label: 'Subject' },
  { key: 'grade', label: 'Grade', options: selectedSchool.value?.grades || [] },
  { key: 'description', label: 'Description', type: 'textarea' }
])
</script>

<template>
  <DomainListPage
    title="My Classes"
    singular="Class"
    endpoint="/api/v1/classes"
    icon="i-lucide-presentation"
    detail-base="/dashboard/classes"
    description="Organize classes by subject and grade for the active academic year."
    academic-year-filter
    grade-filter
    :sort-options="[
      { label: 'Name', value: 'name' },
      { label: 'Subject', value: 'subject' },
      { label: 'Grade', value: 'grade' }
    ]"
    :card-fields="[
      { key: 'subject', label: 'Subject' },
      { key: 'grade', label: 'Grade' },
      { key: 'studentCount', label: 'Students' },
      { key: 'description', label: 'Description' }
    ]"
    :fields="fields"
  />
</template>
