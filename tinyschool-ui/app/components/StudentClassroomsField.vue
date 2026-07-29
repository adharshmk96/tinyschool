<script setup lang="ts">
import type { StudentClassroom } from '~/types/api'

const model = defineModel<StudentClassroom[]>({ required: true })
const { availableAcademicYears, selectedSchool, selectedYearId } = useSchoolContext()

const classroomOptions = computed(() => selectedSchool.value?.classrooms || [])
const yearOptions = computed(() => availableAcademicYears.value.map(year => ({
  label: year.isCurrent ? `${year.name} (current)` : year.name,
  value: year.id
})))

function usedYearIds(exceptIndex: number) {
  return model.value.filter((_, index) => index !== exceptIndex).map(row => row.academicYearId)
}

function optionsFor(index: number) {
  const taken = usedYearIds(index)
  return yearOptions.value.filter(option => !taken.includes(option.value))
}

function addRow() {
  const taken = model.value.map(row => row.academicYearId)
  const nextYear = yearOptions.value.find(option => !taken.includes(option.value))
  if (!nextYear) return
  model.value = [...model.value, { academicYearId: nextYear.value, classroom: classroomOptions.value[0] || '' }]
}

function removeRow(index: number) {
  model.value = model.value.filter((_, current) => current !== index)
}

onMounted(() => {
  if (!model.value.length && selectedYearId.value)
    model.value = [{ academicYearId: selectedYearId.value, classroom: classroomOptions.value[0] || '' }]
})
</script>

<template>
  <div class="space-y-2">
    <div
      v-for="(row, index) in model"
      :key="index"
      class="grid grid-cols-[1fr_1fr_auto] items-center gap-2"
    >
      <USelect
        v-model="row.academicYearId"
        :items="optionsFor(index)"
        value-key="value"
        class="w-full"
        aria-label="Academic year"
        placeholder="Academic year"
      />
      <USelect
        v-model="row.classroom"
        :items="classroomOptions"
        class="w-full"
        aria-label="Classroom"
        placeholder="Classroom"
      />
      <UButton
        icon="i-lucide-x"
        color="neutral"
        variant="ghost"
        aria-label="Remove classroom"
        @click="removeRow(index)"
      />
    </div>
    <p
      v-if="!model.length"
      class="text-sm text-muted"
    >
      No classroom recorded yet. Add one per academic year.
    </p>
    <UButton
      icon="i-lucide-plus"
      label="Add academic year"
      color="neutral"
      variant="outline"
      size="xs"
      :disabled="!yearOptions.length || model.length >= yearOptions.length"
      @click="addRow"
    />
  </div>
</template>
