<script setup lang="ts">
import type { School } from '~/types/api'

definePageMeta({ layout: 'dashboard' })
useSeoMeta({ title: 'School settings' })

const toast = useToast()
const { schools, selectedSchoolId, load } = useSchoolContext()
const modalOpen = ref(false)
const deleteOpen = ref(false)
const editingId = ref<string | null>(null)
const deleting = ref<School | null>(null)
const form = reactive({ name: '', grades: '' })

function openCreate() {
  editingId.value = null
  form.name = ''
  form.grades = ''
  modalOpen.value = true
}

function openEdit(school: School) {
  editingId.value = school.id
  form.name = school.name
  form.grades = school.grades.join(', ')
  modalOpen.value = true
}

function save() {
  const name = form.name.trim()
  const grades = form.grades.split(',').map(item => item.trim()).filter(Boolean)
  if (!name || !grades.length) {
    toast.add({ title: 'School name and grades are required', color: 'error' })
    return
  }
  if (editingId.value) {
    const school = schools.value.find(item => item.id === editingId.value)
    if (school) Object.assign(school, { name, grades })
  } else {
    const id = `school-${Date.now()}`
    schools.value.push({ id, name, grades })
    selectedSchoolId.value = id
  }
  modalOpen.value = false
  toast.add({ title: editingId.value ? 'School updated' : 'School created', color: 'success' })
}

function requestDelete(school: School) {
  deleting.value = school
  deleteOpen.value = true
}

function remove() {
  if (!deleting.value) return
  if (schools.value.length <= 1) {
    toast.add({ title: 'Keep at least one school', description: 'Create another school before deleting this one.', color: 'error' })
    return
  }
  schools.value = schools.value.filter(item => item.id !== deleting.value?.id)
  if (selectedSchoolId.value === deleting.value.id) selectedSchoolId.value = schools.value[0]?.id
  toast.add({ title: 'School deleted', color: 'success' })
}

onMounted(async () => {
  try {
    await load()
  } catch {
    toast.add({ title: 'Could not load schools', color: 'error' })
  }
})
</script>

<template>
  <SettingsShell>
    <PageHeading
      title="My schools"
      description="Manage schools and the grades taught in each."
    >
      <template #actions>
        <UButton
          icon="i-lucide-plus"
          label="Create school"
          @click="openCreate"
        />
      </template>
    </PageHeading>

    <div class="mt-6 space-y-3">
      <UCard
        v-for="school in schools"
        :key="school.id"
      >
        <div class="flex flex-col gap-4 sm:flex-row sm:items-center">
          <div class="grid size-11 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
            <UIcon
              name="i-lucide-school"
              class="size-5"
            />
          </div>
          <div class="min-w-0 flex-1">
            <div class="flex items-center gap-2">
              <h2 class="truncate font-semibold">
                {{ school.name }}
              </h2>
              <UBadge
                v-if="school.id === selectedSchoolId"
                label="Active"
                size="sm"
                variant="subtle"
              />
            </div>
            <p class="mt-1 text-sm text-muted">
              Grades: {{ school.grades.join(', ') }}
            </p>
          </div>
          <div class="flex gap-1 self-end sm:self-auto">
            <UButton
              icon="i-lucide-pencil"
              color="neutral"
              variant="ghost"
              :aria-label="`Edit ${school.name}`"
              @click="openEdit(school)"
            />
            <UButton
              icon="i-lucide-trash-2"
              color="error"
              variant="ghost"
              :aria-label="`Delete ${school.name}`"
              @click="requestDelete(school)"
            />
          </div>
        </div>
      </UCard>
      <EmptyState
        v-if="!schools.length"
        icon="i-lucide-school"
        title="Create your first school"
        description="A school is needed before you can manage students and academic years."
      >
        <UButton
          icon="i-lucide-plus"
          label="Create school"
          @click="openCreate"
        />
      </EmptyState>
    </div>

    <UModal
      v-model:open="modalOpen"
      :title="editingId ? 'Edit school' : 'Create school'"
      description="Add the school name and comma-separated grades."
    >
      <template #body>
        <form
          id="school-form"
          class="space-y-5"
          @submit.prevent="save"
        >
          <UFormField
            label="School name"
            required
          >
            <UInput
              v-model="form.name"
              autofocus
              class="w-full"
              placeholder="Oakridge Learning Centre"
            />
          </UFormField>
          <UFormField
            label="Grades"
            description="Separate each grade with a comma."
            required
          >
            <UInput
              v-model="form.grades"
              class="w-full"
              placeholder="Grade 6, Grade 7, Grade 8"
            />
          </UFormField>
        </form>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton
            label="Cancel"
            color="neutral"
            variant="ghost"
            @click="modalOpen = false"
          />
          <UButton
            form="school-form"
            type="submit"
            :label="editingId ? 'Save changes' : 'Create school'"
          />
        </div>
      </template>
    </UModal>

    <ConfirmDialog
      v-model="deleteOpen"
      title="Delete school?"
      :description="`Delete ${deleting?.name || 'this school'} and its placeholder data?`"
      @confirm="remove"
    />
  </SettingsShell>
</template>
