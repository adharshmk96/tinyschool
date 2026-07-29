<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'

const model = defineModel<string | undefined>({ default: undefined })

const { selectedSchool } = useSchoolContext()

const grades = computed(() => selectedSchool.value?.grades || [])

const items = computed<DropdownMenuItem[][]>(() => [[
  {
    label: 'Grade',
    icon: 'i-lucide-graduation-cap',
    filter: { placeholder: 'Search grades…', icon: 'i-lucide-search' },
    children: [[
      {
        label: 'All grades',
        icon: model.value ? undefined : 'i-lucide-check',
        onSelect: () => { model.value = undefined }
      },
      ...grades.value.map(grade => ({
        label: grade,
        icon: model.value?.toLowerCase() === grade.toLowerCase() ? 'i-lucide-check' : undefined,
        onSelect: () => { model.value = grade }
      }))
    ]]
  }
]])

const buttonLabel = computed(() => model.value ? `Grade: ${model.value}` : 'Filter')
</script>

<template>
  <UDropdownMenu
    :items="items"
    :content="{ align: 'start' }"
    :ui="{ content: 'min-w-44' }"
  >
    <UButton
      color="neutral"
      variant="outline"
      icon="i-lucide-list-filter"
      :label="buttonLabel"
      trailing-icon="i-lucide-chevron-down"
    />
  </UDropdownMenu>
</template>
