<script setup lang="ts">
const props = defineProps<{
  kind: 'assignment' | 'exam' | 'student'
  items?: unknown
  totalScore?: number
}>()

const toast = useToast()
const scores = reactive<Record<string, string>>({})
const completed = reactive<Record<string, boolean>>({})
const people = computed(() => {
  const source = Array.isArray(props.items) ? props.items : []
  return source.map((entry, index) => {
    const item = entry as Record<string, unknown>
    const score = item.score
    const total = Number(item.totalScore ?? props.totalScore ?? 100)
    return {
      id: String(item.id ?? index),
      name: String(item.name ?? `Student ${index + 1}`),
      detail: `${String(item.grade ?? '')}${score !== null && score !== undefined ? ` · ${score} / ${total}` : ''}`,
      date: String(item.completedAt ?? item.markedAt ?? ''),
      initialCompleted: score !== null && score !== undefined,
      total
    }
  })
})

function isCompleted(person: { id: string, initialCompleted?: boolean }) {
  return completed[person.id] ?? person.initialCompleted ?? false
}

function save(id: string, total: number) {
  const score = Number(scores[id])
  if (!Number.isFinite(score) || score < 0 || score > total) {
    toast.add({ title: 'Enter a valid score', description: `Score must be between 0 and ${total}.`, color: 'error' })
    return
  }
  completed[id] = true
  toast.add({ title: 'Score saved', description: 'A score log was added.', color: 'success' })
}

function reset(id: string) {
  completed[id] = false
  scores[id] = ''
  toast.add({ title: 'Score reset', color: 'success' })
}
</script>

<template>
  <TabPlaceholder
    v-if="!people.length"
    icon="i-lucide-list-checks"
    title="No data available"
    description="Students and their scores will appear here when they are linked."
  />
  <div
    v-else
    class="space-y-3"
  >
    <UCard v-for="person in people" :key="person.id">
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center">
        <UAvatar :alt="person.name" />
        <div class="min-w-0 flex-1">
          <p class="font-medium text-highlighted">
            {{ person.name }}
          </p>
          <p class="text-sm text-muted">
            {{ isCompleted(person) ? person.detail : kind === 'student' ? 'Awaiting score' : 'Grade 8 · Awaiting score' }}
          </p>
          <p v-if="isCompleted(person)" class="mt-1 text-xs text-muted">
            {{ person.date || 'Marked just now' }}
          </p>
        </div>
        <div v-if="isCompleted(person)" class="flex items-center gap-2">
          <UBadge color="success" variant="subtle" label="Completed" />
          <UButton label="Reset" icon="i-lucide-rotate-ccw" color="neutral" variant="outline" @click="reset(person.id)" />
        </div>
        <div v-else class="flex items-center gap-2">
          <UInput v-model="scores[person.id]" type="number" min="0" max="100" placeholder="Score" class="w-28" />
          <span class="text-sm text-muted">/ {{ 'total' in person ? person.total : totalScore || 100 }}</span>
          <UButton
            label="Save"
            @click="save(person.id, 'total' in person ? person.total : totalScore || 100)"
          />
        </div>
      </div>
    </UCard>
  </div>
</template>
