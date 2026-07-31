<script setup lang="ts">
import type { AcademicYear, ImportSummary, School } from '~/types/api'

definePageMeta({ layout: 'dashboard' })
useSeoMeta({ title: 'Data settings' })

const toast = useToast()
const { request, download, upload, getCollection } = useApi()
const { schools, academicYears, selectedSchoolId, selectedYearId } = useSchoolContext()

const confirmationOpen = ref(false)
const pending = ref(false)
const exporting = ref<'xlsx' | 'csv' | undefined>()
const importFile = ref<File | undefined>()
const importConfirmationOpen = ref(false)
const importing = ref(false)
const lastImport = ref<ImportSummary | undefined>()

const exportFormats = [
  {
    format: 'xlsx' as const,
    label: 'Download Excel workbook',
    icon: 'i-lucide-file-spreadsheet',
    hint: 'One .xlsx file with a sheet per record type.'
  },
  {
    format: 'csv' as const,
    label: 'Download CSV files',
    icon: 'i-lucide-file-text',
    hint: 'A .zip holding one CSV per record type.'
  }
]

const importedCounts = computed(() => {
  const summary = lastImport.value
  if (!summary) return []
  return [
    { label: 'Schools', value: summary.schools },
    { label: 'Academic years', value: summary.academicYears },
    { label: 'Students', value: summary.students },
    { label: 'Classes', value: summary.classes },
    { label: 'Assignments', value: summary.assignments },
    { label: 'Exams', value: summary.exams },
    { label: 'Scores', value: summary.scores },
    { label: 'Behaviour and notes', value: summary.studentLogs }
  ]
})

async function exportData(format: 'xlsx' | 'csv') {
  exporting.value = format
  try {
    const blob = await download(`/me/data/export?format=${format}`)
    const stamp = new Date().toISOString().slice(0, 10)
    saveBlob(blob, `tinyschool-export-${stamp}.${format === 'csv' ? 'zip' : 'xlsx'}`)
  } catch (error) {
    toast.add({
      title: 'Could not export data',
      description: apiErrorMessage(error, 'The export could not be created. Try again.'),
      color: 'error'
    })
  } finally {
    exporting.value = undefined
  }
}

function saveBlob(blob: Blob, filename: string) {
  const url = URL.createObjectURL(blob)
  const link = document.createElement('a')
  link.href = url
  link.download = filename
  link.click()
  URL.revokeObjectURL(url)
}

async function replaceData() {
  const file = importFile.value
  if (!file) return
  importing.value = true
  try {
    const body = new FormData()
    body.append('file', file)
    const response = await upload<ImportSummary>('/me/data/import', body)
    lastImport.value = response.data
    importFile.value = undefined
    toast.add({
      title: 'Data imported',
      description: 'Your workspace now matches the uploaded file.',
      color: 'success'
    })
    // Every record was stored under a new id, so the cached school and year
    // selection has to be rebuilt from what the import created.
    await refreshSchoolContext()
  } catch (error) {
    toast.add({
      title: 'Could not import data',
      description: apiErrorMessage(error, 'Nothing was changed. Check the file and try again.'),
      color: 'error'
    })
  } finally {
    importing.value = false
  }
}

async function refreshSchoolContext() {
  const [schoolResponse, yearResponse] = await Promise.all([
    getCollection<School>('/schools'),
    getCollection<AcademicYear>('/academic-years')
  ])
  schools.value = schoolResponse.data || []
  academicYears.value = yearResponse.data || []
  selectedSchoolId.value = schools.value[0]?.id
  const yearsForSchool = academicYears.value.filter(item => item.schoolId === selectedSchoolId.value)
  selectedYearId.value = yearsForSchool.find(item => item.isCurrent)?.id || yearsForSchool[0]?.id
}

async function clearData() {
  pending.value = true
  try {
    await request('/me/data', { method: 'DELETE' })
    schools.value = []
    academicYears.value = []
    selectedSchoolId.value = undefined
    selectedYearId.value = undefined
    toast.add({
      title: 'Data cleared',
      description: 'Your schools and all related records were deleted.',
      color: 'success'
    })
    await navigateTo('/onboarding')
  } catch (error) {
    toast.add({
      title: 'Could not clear data',
      description: apiErrorMessage(error, 'Your data was not changed. Try again.'),
      color: 'error'
    })
  } finally {
    pending.value = false
  }
}
</script>

<template>
  <SettingsShell>
    <div class="space-y-6">
      <UCard>
        <template #header>
          <h2 class="font-semibold">
            Export data
          </h2>
          <p class="text-sm text-muted">
            Download everything you own — schools, academic years, students, classes, assignments, exams, scores and notes.
          </p>
        </template>

        <div class="flex flex-col gap-3 sm:flex-row">
          <div
            v-for="option in exportFormats"
            :key="option.format"
            class="flex-1"
          >
            <UButton
              block
              color="neutral"
              variant="subtle"
              :icon="option.icon"
              :label="option.label"
              :loading="exporting === option.format"
              :disabled="!!exporting"
              @click="exportData(option.format)"
            />
            <p class="mt-2 text-xs text-muted">
              {{ option.hint }}
            </p>
          </div>
        </div>
      </UCard>

      <UCard>
        <template #header>
          <h2 class="font-semibold">
            Import data
          </h2>
          <p class="text-sm text-muted">
            Upload an exported <code>.xlsx</code> workbook or <code>.zip</code> of CSV files. Because an import
            replaces everything, the file has to describe the whole workspace: rows are linked to each other by
            the ids it carries.
          </p>
        </template>

        <UAlert
          class="mb-5"
          color="warning"
          variant="subtle"
          icon="i-lucide-triangle-alert"
          title="Importing replaces your current data"
          description="Everything you own is deleted and rewritten from the file. Export a copy first if you want a backup."
        />

        <UFileUpload
          v-model="importFile"
          layout="list"
          accept=".xlsx,.zip,.csv"
          label="Drop your export here"
          description="An .xlsx workbook, or a .zip of the exported CSV files"
          class="min-h-40"
        />

        <UButton
          class="mt-4"
          icon="i-lucide-upload"
          label="Import and replace"
          :loading="importing"
          :disabled="!importFile"
          @click="importConfirmationOpen = true"
        />

        <div
          v-if="lastImport"
          class="mt-5 rounded-lg border border-default p-4"
        >
          <p class="mb-3 text-sm font-semibold">
            Last import
          </p>
          <dl class="grid grid-cols-2 gap-3 text-sm sm:grid-cols-4">
            <div
              v-for="entry in importedCounts"
              :key="entry.label"
            >
              <dt class="text-xs text-muted">
                {{ entry.label }}
              </dt>
              <dd class="font-medium">
                {{ entry.value }}
              </dd>
            </div>
          </dl>
        </div>
      </UCard>

      <UCard>
        <template #header>
          <h2 class="font-semibold">
            Clear data
          </h2>
          <p class="text-sm text-muted">
            Delete your schools, academic years, students, classes, assignments, exams and related records.
          </p>
        </template>

        <UAlert
          class="mb-5"
          color="error"
          variant="subtle"
          icon="i-lucide-triangle-alert"
          title="This cannot be undone"
          description="Your account and sign-in details will remain."
        />
        <UButton
          color="error"
          icon="i-lucide-trash-2"
          label="Clear data"
          :loading="pending"
          @click="confirmationOpen = true"
        />
      </UCard>
    </div>

    <ConfirmDialog
      v-model="importConfirmationOpen"
      title="Replace all your data?"
      description="Your current schools, students, classes, assignments, exams and scores are deleted and rebuilt from the uploaded file."
      confirm-label="Import and replace"
      @confirm="replaceData"
    />

    <ConfirmDialog
      v-model="confirmationOpen"
      title="Clear all your data?"
      description="This permanently deletes all workspace data owned by your account. Other users are not affected."
      confirm-label="Clear data"
      @confirm="clearData"
    />
  </SettingsShell>
</template>
