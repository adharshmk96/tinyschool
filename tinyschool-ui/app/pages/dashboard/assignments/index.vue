<script setup lang="ts">
definePageMeta({ layout: 'dashboard' })
</script>

<template>
  <DomainListPage
    title="Assignments"
    singular="Assignment"
    endpoint="/api/v1/assignments"
    icon="i-lucide-clipboard-check"
    detail-base="/dashboard/assignments"
    description="Track class and individual work, due dates, completion and scores."
    academic-year-filter
    grade-filter
    :sort-options="[
      { label: 'Due date', value: 'dueDate' },
      { label: 'Name', value: 'name' },
      { label: 'Type', value: 'type' }
    ]"
    :card-fields="[
      { key: 'type', label: 'Type' },
      { key: 'class.name', label: 'Class' },
      { key: 'dueDate', label: 'Due date' },
      { key: 'totalScore', label: 'Total score' },
      { key: 'completion', label: 'Completion' }
    ]"
    :fields="[
      { key: 'name', label: 'Assignment name' },
      { key: 'type', label: 'Type', options: ['Class', 'Individual'] },
      { key: 'dueDate', label: 'Due date', type: 'date' },
      { key: 'totalScore', label: 'Total score', type: 'number' },
      { key: 'assignees', label: 'Class' }
    ]"
  />
</template>
