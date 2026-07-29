<script setup lang="ts">
import type { DropdownMenuItem } from '@nuxt/ui'

const model = defineModel<string | undefined>({ default: undefined })

const { selectedSchool } = useSchoolContext()

const classrooms = computed(() => selectedSchool.value?.classrooms || [])

const items = computed<DropdownMenuItem[][]>(() => [[
  {
    label: 'Classroom',
    icon: 'i-lucide-door-open',
    filter: { placeholder: 'Search classrooms…', icon: 'i-lucide-search' },
    children: [[
      {
        label: 'All classrooms',
        icon: model.value ? undefined : 'i-lucide-check',
        onSelect: () => { model.value = undefined }
      },
      ...classrooms.value.map(classroom => ({
        label: classroom,
        icon: model.value?.toLowerCase() === classroom.toLowerCase() ? 'i-lucide-check' : undefined,
        onSelect: () => { model.value = classroom }
      }))
    ]]
  }
]])

const buttonLabel = computed(() => model.value ? `Classroom: ${model.value}` : 'Filter')
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
