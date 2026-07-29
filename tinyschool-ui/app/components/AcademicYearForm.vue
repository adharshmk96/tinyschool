<script setup lang="ts">
import type { AcademicSegment, AcademicYear } from '~/types/api'

const props = defineProps<{
  initial?: AcademicYear
  /** Render without the built-in footer so a host (e.g. the onboarding wizard) can drive submission. */
  hideActions?: boolean
  /** Form id to target from an external submit button when `hideActions` is set. */
  formId?: string
  /** Drop the card chrome so the form blends into a host container. */
  flat?: boolean
}>()

const cardUi = computed(() => props.flat
  ? { root: 'bg-transparent ring-0 shadow-none', header: 'px-0 pt-0', body: 'px-0', footer: 'px-0' }
  : undefined)
const emit = defineEmits<{ save: [value: AcademicYear] }>()
const toast = useToast()

const form = reactive({
  name: props.initial?.name || '',
  startDate: props.initial?.startDate || new Date().toISOString().slice(0, 10),
  segments: (props.initial?.segments?.length
    ? props.initial.segments
    : [
        { name: 'Term 1', type: 'term', durationDays: 90 },
        { name: 'Summer break', type: 'vacation', durationDays: 30 },
        { name: 'Term 2', type: 'term', durationDays: 90 }
      ]).map(item => ({ ...item })) as AcademicSegment[]
})

const durationDays = computed(() => form.segments.reduce((sum, segment) => sum + Number(segment.durationDays || 0), 0))

function addSegment() {
  form.segments.push({ name: `Term ${form.segments.filter(item => item.type === 'term').length + 1}`, type: 'term', durationDays: 30 })
}

function removeSegment(index: number) {
  form.segments.splice(index, 1)
}

function submit() {
  if (!form.name.trim() || !form.startDate || !form.segments.length) {
    toast.add({ title: 'Complete the academic year', description: 'Name, start date and at least one segment are required.', color: 'error' })
    return
  }
  if (form.segments.some(segment => !segment.name.trim() || Number(segment.durationDays) < 1)) {
    toast.add({ title: 'Check the segments', description: 'Every segment needs a name and positive duration.', color: 'error' })
    return
  }
  emit('save', {
    id: props.initial?.id || `year-${Date.now()}`,
    name: form.name.trim(),
    startDate: form.startDate,
    durationDays: durationDays.value,
    isCurrent: props.initial?.isCurrent,
    segments: form.segments.map(segment => ({ ...segment, durationDays: Number(segment.durationDays) }))
  })
}
</script>

<template>
  <form
    :id="formId"
    class="space-y-6"
    @submit.prevent="submit"
  >
    <UCard :ui="cardUi">
      <template #header>
        <h2 class="font-semibold">
          Academic year details
        </h2><p class="text-sm text-muted">
          Give this calendar a clear name and start date.
        </p>
      </template>
      <div class="grid gap-5 sm:grid-cols-2">
        <UFormField
          label="Name"
          required
        >
          <UInput
            v-model="form.name"
            class="w-full"
            placeholder="2026–27"
          />
        </UFormField>
        <UFormField
          label="Start date"
          required
        >
          <UInput
            v-model="form.startDate"
            type="date"
            class="w-full"
          />
        </UFormField>
      </div>
    </UCard>

    <UCard :ui="cardUi">
      <template #header>
        <div class="flex items-center justify-between gap-4">
          <div>
            <h2 class="font-semibold">
              Segments
            </h2><p class="text-sm text-muted">
              Build the year from terms and vacations.
            </p>
          </div>
          <UButton
            type="button"
            icon="i-lucide-plus"
            color="neutral"
            variant="outline"
            label="Add segment"
            @click="addSegment"
          />
        </div>
      </template>
      <div class="space-y-3">
        <div
          v-for="(segment, index) in form.segments"
          :key="index"
          class="grid items-end gap-3 rounded-lg border border-default p-4 sm:grid-cols-[1fr_15rem_8rem_auto]"
        >
          <UFormField
            :label="`Segment ${index + 1}`"
            required
          >
            <UInput
              v-model="segment.name"
              class="w-full"
            />
          </UFormField>
          <UFormField
            label="Type"
            required
          >
            <div class="grid grid-cols-2 rounded-lg bg-elevated p-1">
              <UButton
                type="button"
                label="Term"
                icon="i-lucide-book-open"
                :variant="segment.type === 'term' ? 'solid' : 'ghost'"
                :color="segment.type === 'term' ? 'primary' : 'neutral'"
                class="justify-center"
                @click="segment.type = 'term'"
              />
              <UButton
                type="button"
                label="Vacation"
                icon="i-lucide-palmtree"
                :variant="segment.type === 'vacation' ? 'solid' : 'ghost'"
                :color="segment.type === 'vacation' ? 'primary' : 'neutral'"
                class="justify-center"
                @click="segment.type = 'vacation'"
              />
            </div>
          </UFormField>
          <UFormField
            label="Duration (days)"
            required
          >
            <UInputNumber
              v-model="segment.durationDays"
              :min="1"
              class="w-full"
            />
          </UFormField>
          <UButton
            type="button"
            icon="i-lucide-trash-2"
            color="error"
            variant="ghost"
            :aria-label="`Remove ${segment.name}`"
            @click="removeSegment(index)"
          />
        </div>
      </div>
    </UCard>

    <UCard :ui="cardUi">
      <template #header>
        <h2 class="font-semibold">
          Timeline preview
        </h2><p class="text-sm text-muted">
          {{ durationDays }} days across {{ form.segments.length }} segments.
        </p>
      </template>
      <AcademicTimeline
        :start-date="form.startDate"
        :segments="form.segments"
      />
    </UCard>

    <div
      v-if="!hideActions"
      class="flex justify-end gap-2"
    >
      <UButton
        type="button"
        to="/dashboard/settings/academic-years"
        color="neutral"
        variant="ghost"
        label="Cancel"
      />
      <UButton
        type="submit"
        :label="initial ? 'Save changes' : 'Create academic year'"
      />
    </div>
  </form>
</template>
