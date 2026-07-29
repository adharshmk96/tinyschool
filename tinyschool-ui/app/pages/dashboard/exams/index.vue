<script setup lang="ts">
definePageMeta({ layout: 'dashboard' })
</script>

<template>
  <DomainListPage
    title="Exams"
    singular="Exam"
    endpoint="/api/v1/exams"
    icon="i-lucide-file-check-2"
    detail-base="/dashboard/exams"
    description="Schedule examinations and track marking progress by class."
    academic-year-filter
    grade-filter
    :sort-options="[
      { label: 'Exam date', value: 'examDate' },
      { label: 'Name', value: 'name' },
      { label: 'Marked', value: 'markedCount' }
    ]"
    :card-fields="[
      { key: 'class.name', label: 'Class' },
      { key: 'examDate', label: 'Exam date' },
      { key: 'totalScore', label: 'Total score' },
      { key: 'markedCount', label: 'Marked' }
    ]"
    :fields="[
      { key: 'name', label: 'Exam name' },
      { key: 'examDate', label: 'Exam date', type: 'date' },
      { key: 'totalScore', label: 'Total score', type: 'number' },
      { key: 'class', label: 'Class', options: ['Grade 8 Mathematics', 'Grade 8 Science', 'Grade 9 English'] }
    ]"
  />
</template>
