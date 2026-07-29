<script setup lang="ts">
const props = defineProps<{
  form: { name: string, classrooms: string[], classroomsInUse: string[], newClassroom: string }
  autofocus?: boolean
}>()

const toast = useToast()
const form = props.form

function classroomInUse(classroom: string) {
  return form.classroomsInUse.some(item => item.toLowerCase() === classroom.toLowerCase())
}

function addClassroom() {
  const classroom = form.newClassroom.trim()
  if (!classroom) return
  if (form.classrooms.some(item => item.toLowerCase() === classroom.toLowerCase())) {
    toast.add({ title: 'Classroom already added', color: 'error' })
    return
  }
  form.classrooms.push(classroom)
  form.newClassroom = ''
}

function removeClassroom(classroom: string) {
  if (classroomInUse(classroom)) {
    toast.add({ title: 'Classroom is linked', description: 'Remove linked classes or students before deleting this classroom.', color: 'error' })
    return
  }
  form.classrooms = form.classrooms.filter(item => item !== classroom)
}
</script>

<template>
  <div class="space-y-5">
    <UFormField
      label="School name"
      required
    >
      <UInput
        v-model="form.name"
        :autofocus="autofocus"
        class="w-full"
        placeholder="Oakridge Learning Centre"
      />
    </UFormField>
    <UFormField
      label="Classrooms"
      description="Linked classrooms cannot be removed while classes or students use them."
      required
    >
      <div class="space-y-3">
        <div
          v-if="form.classrooms.length"
          class="flex flex-wrap gap-2"
        >
          <div
            v-for="classroom in form.classrooms"
            :key="classroom"
            class="inline-flex items-center gap-1 rounded-md bg-elevated px-2.5 py-1 text-sm"
          >
            <span>{{ classroom }}</span>
            <UButton
              :icon="classroomInUse(classroom) ? 'i-lucide-lock' : 'i-lucide-x'"
              size="xs"
              color="neutral"
              variant="ghost"
              class="-me-1"
              :disabled="classroomInUse(classroom)"
              :aria-label="classroomInUse(classroom) ? `${classroom} is linked` : `Remove ${classroom}`"
              @click.prevent="removeClassroom(classroom)"
            />
          </div>
        </div>
        <div class="flex gap-2">
          <UInput
            v-model="form.newClassroom"
            class="w-full"
            placeholder="10A"
            @keydown.enter.prevent="addClassroom"
          />
          <UButton
            type="button"
            icon="i-lucide-plus"
            label="Add"
            color="neutral"
            variant="outline"
            @click="addClassroom"
          />
        </div>
      </div>
    </UFormField>
  </div>
</template>
