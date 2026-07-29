<script setup lang="ts">
import type { AcademicYear, School } from '~/types/api'

definePageMeta({ layout: 'default' })
useSeoMeta({ title: 'Set up your workspace' })

const toast = useToast()
const { postItem } = useApi()
const { schools, academicYears, availableAcademicYears, selectedSchoolId, selectedYearId, load } = useSchoolContext()

const steps = [
  { key: 'school', label: 'School', description: 'Name and classrooms', icon: 'i-lucide-school' },
  { key: 'year', label: 'Academic year', description: 'Terms and vacations', icon: 'i-lucide-calendar-range' },
  { key: 'done', label: 'Finish', description: 'Start teaching', icon: 'i-lucide-party-popper' }
] as const

const stepIndex = ref(0)
const saving = ref(false)
const ready = ref(false)
const schoolForm = reactive({ name: '', classrooms: [] as string[], classroomsInUse: [] as string[], newClassroom: '' })

const copy = computed(() => [
  {
    title: 'Create your school',
    description: 'Add the school you teach at and the classrooms you handle. You can add more later in settings.'
  },
  {
    title: 'Set up your academic year',
    description: 'Build the calendar from terms and vacations. Everything you track is grouped by this year.'
  },
  {
    title: 'You are all set',
    description: 'Your workspace is ready. Add students and classes whenever you are.'
  }
][stepIndex.value]!)

async function saveSchool() {
  if (saving.value) return
  const name = schoolForm.name.trim()
  const classrooms = schoolForm.classrooms.map(item => item.trim()).filter(Boolean)
  if (!name || !classrooms.length) {
    toast.add({ title: 'School name and classrooms are required', color: 'error' })
    return
  }
  saving.value = true
  try {
    const response = await postItem<School>('/schools', { name, classrooms, isActive: true })
    schools.value.push(response.data)
    selectedSchoolId.value = response.data.id
    stepIndex.value = 1
  } catch (error) {
    toast.add({ title: 'Could not create school', description: apiErrorMessage(error, 'Please try again.'), color: 'error' })
  } finally {
    saving.value = false
  }
}

async function saveYear(year: AcademicYear) {
  if (saving.value) return
  if (!selectedSchoolId.value) {
    toast.add({ title: 'Create a school first', color: 'error' })
    stepIndex.value = 0
    return
  }
  saving.value = true
  try {
    const response = await postItem<AcademicYear>('/academic-years', {
      schoolId: selectedSchoolId.value,
      name: year.name,
      startDate: year.startDate,
      isCurrent: true,
      segments: year.segments.map(segment => ({
        name: segment.name,
        type: segment.type,
        durationDays: Number(segment.durationDays)
      }))
    })
    academicYears.value = academicYears.value.map(item => ({ ...item, isCurrent: false }))
    academicYears.value.push(response.data)
    selectedYearId.value = response.data.id
    stepIndex.value = 2
  } catch (error) {
    toast.add({ title: 'Could not create academic year', description: apiErrorMessage(error, 'Please try again.'), color: 'error' })
  } finally {
    saving.value = false
  }
}

onMounted(async () => {
  try {
    await load()
  } catch {
    toast.add({ title: 'Could not load your workspace', description: 'Make sure the local API is running.', color: 'error' })
  }
  if (schools.value.length && availableAcademicYears.value.length) {
    await navigateTo('/dashboard')
    return
  }
  stepIndex.value = schools.value.length ? 1 : 0
  ready.value = true
})
</script>

<template>
  <OnboardingShell
    :steps="steps"
    :active-index="stepIndex"
    :title="copy.title"
    :description="copy.description"
  >
    <USkeleton
      v-if="!ready"
      class="h-64 rounded-lg"
    />

    <form
      v-else-if="stepIndex === 0"
      id="onboarding-school-form"
      @submit.prevent="saveSchool"
    >
      <SchoolFields
        :form="schoolForm"
        autofocus
      />
    </form>

    <AcademicYearForm
      v-else-if="stepIndex === 1"
      form-id="onboarding-year-form"
      hide-actions
      flat
      @save="saveYear"
    />

    <div
      v-else
      class="flex flex-col items-center gap-4 py-6 text-center"
    >
      <div class="grid size-14 place-items-center rounded-full bg-primary/10 text-primary">
        <UIcon
          name="i-lucide-check"
          class="size-7"
        />
      </div>
      <div>
        <p class="font-semibold text-highlighted">
          {{ schools.find(item => item.id === selectedSchoolId)?.name }}
        </p>
        <p class="text-sm text-muted">
          {{ availableAcademicYears.find(item => item.id === selectedYearId)?.name }} is your active academic year.
        </p>
      </div>
    </div>

    <template #footer>
      <UButton
        v-if="stepIndex === 0"
        form="onboarding-school-form"
        type="submit"
        :loading="saving"
        label="Create school"
        trailing-icon="i-lucide-arrow-right"
      />
      <template v-else-if="stepIndex === 1">
        <UButton
          form="onboarding-year-form"
          type="submit"
          :loading="saving"
          label="Create academic year"
          trailing-icon="i-lucide-arrow-right"
        />
      </template>
      <UButton
        v-else
        to="/dashboard"
        label="Go to dashboard"
        trailing-icon="i-lucide-arrow-right"
      />
    </template>
  </OnboardingShell>
</template>
