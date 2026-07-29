<script setup lang="ts">
import type { School } from '~/types/api'

definePageMeta({ layout: 'dashboard' })
useSeoMeta({ title: 'School settings' })

const toast = useToast()
const { postItem, patchItem, request } = useApi()
const route = useRoute()
const { schools, selectedSchoolId, load } = useSchoolContext()
const modalOpen = ref(false)
const deleteOpen = ref(false)
const editingId = ref<string | null>(null)
const deleting = ref<School | null>(null)
const form = reactive({ name: '', classrooms: [] as string[], classroomsInUse: [] as string[], newClassroom: '' })
const saving = ref(false)
const deletingSchool = ref(false)

function openCreate() {
  editingId.value = null
  form.name = ''
  form.classrooms = []
  form.classroomsInUse = []
  form.newClassroom = ''
  modalOpen.value = true
}

function openEdit(school: School) {
  editingId.value = school.id
  form.name = school.name
  form.classrooms = [...school.classrooms]
  form.classroomsInUse = [...(school.classroomsInUse || [])]
  form.newClassroom = ''
  modalOpen.value = true
}

async function save() {
  const name = form.name.trim()
  const classrooms = form.classrooms.map(item => item.trim()).filter(Boolean)
  if (!name || !classrooms.length) {
    toast.add({ title: 'School name and classrooms are required', color: 'error' })
    return
  }
  saving.value = true
  try {
    const existing = schools.value.find(item => item.id === editingId.value) as (School & { isActive?: boolean }) | undefined
    const body = { name, classrooms, isActive: existing?.isActive ?? true }
    const response = editingId.value
      ? await patchItem<School>(`/schools/${editingId.value}`, body)
      : await postItem<School>('/schools', body)
    if (editingId.value) {
      const index = schools.value.findIndex(item => item.id === editingId.value)
      if (index >= 0) schools.value[index] = response.data
    } else {
      schools.value.push(response.data)
      selectedSchoolId.value = response.data.id
    }
    modalOpen.value = false
    toast.add({ title: editingId.value ? 'School updated' : 'School created', color: 'success' })
  } catch (error) {
    toast.add({ title: 'Could not save school', description: apiErrorMessage(error, 'Please try again.'), color: 'error' })
  } finally {
    saving.value = false
  }
}

function requestDelete(school: School) {
  deleting.value = school
  deleteOpen.value = true
}

async function remove() {
  if (!deleting.value) return
  deletingSchool.value = true
  try {
    const deletedId = deleting.value.id
    await request(`/schools/${deletedId}`, { method: 'DELETE' })
    if (selectedSchoolId.value === deletedId)
      selectedSchoolId.value = undefined
    schools.value = schools.value.filter(item => item.id !== deletedId)
    selectedSchoolId.value ||= schools.value[0]?.id
    deleteOpen.value = false
    deleting.value = null
    toast.add({ title: 'School deleted', color: 'success' })
  } catch (error) {
    toast.add({ title: 'Could not delete school', description: apiErrorMessage(error, 'Please try again.'), color: 'error' })
  } finally {
    deletingSchool.value = false
  }
}

onMounted(async () => {
  try {
    await load()
    if (route.query.create === '1')
      openCreate()
  } catch {
    toast.add({ title: 'Could not load schools', color: 'error' })
  }
})
</script>

<template>
  <SettingsShell>
    <PageHeading
      title="My schools"
      description="Manage schools and the classrooms taught in each."
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
              Classrooms: {{ school.classrooms.join(', ') }}
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
      description="Add the school name and classrooms taught."
    >
      <template #body>
        <form
          id="school-form"
          @submit.prevent="save"
        >
          <SchoolFields
            :form="form"
            autofocus
          />
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
            :loading="saving"
            :label="editingId ? 'Save changes' : 'Create school'"
          />
        </div>
      </template>
    </UModal>

    <ConfirmDialog
      v-model="deleteOpen"
      title="Delete school?"
      :description="`Delete ${deleting?.name || 'this school'} and its related data?`"
      @confirm="remove"
    />
  </SettingsShell>
</template>
