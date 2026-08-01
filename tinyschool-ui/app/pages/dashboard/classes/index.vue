<script setup lang="ts">
definePageMeta({ layout: 'dashboard' })
const { selectedSchool } = useSchoolContext()
const fields = computed(() => [
  { key: 'name', label: 'Class name' },
  { key: 'subject', label: 'Subject' },
  { key: 'classrooms', label: 'Classrooms', options: selectedSchool.value?.classrooms || [], multiple: true },
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
    description="Organize reusable classes by subject and classroom."
    classroom-filter
    :sort-options="[
      { label: 'Name', value: 'name' },
      { label: 'Subject', value: 'subject' },
      { label: 'Classroom', value: 'classroom' }
    ]"
    :card-fields="[
      { key: 'subject', label: 'Subject' },
      { key: 'classrooms', label: 'Classrooms' },
      { key: 'studentCount', label: 'Students' },
      { key: 'description', label: 'Description' }
    ]"
    :fields="fields"
  />
</template>
