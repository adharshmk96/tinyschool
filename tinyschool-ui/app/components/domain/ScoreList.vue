<script setup lang="ts">
const props = defineProps<{
  kind: 'assignment' | 'exam' | 'student'
  items?: unknown
  totalScore?: number
}>()

const emit = defineEmits<{
  changed: []
}>()

type Person = {
  id: string
  name: string
  classroom: string
  date: string
  score: number | null
  total: number
  resourceKind: 'assignment' | 'exam'
}

const route = useRoute()
const toast = useToast()
const { request } = useApi()
const scores = reactive<Record<string, string>>({})
const overrides = reactive<Record<string, { score: number | null, date: string }>>({})
const pending = reactive<Record<string, boolean>>({})
const people = computed(() => {
  const source = Array.isArray(props.items) ? props.items : []
  return source.map<Person>((entry, index) => {
    const item = entry as Record<string, unknown>
    const score = item.score === null || item.score === undefined ? null : Number(item.score)
    const total = Number(item.totalScore ?? props.totalScore ?? 100)
    return {
      id: String(item.id ?? index),
      name: String(item.name ?? `Student ${index + 1}`),
      classroom: String(item.classroom ?? ''),
      date: String(item.completedAt ?? item.markedAt ?? ''),
      score,
      total,
      resourceKind: item.kind === 'exam' || props.kind === 'exam' ? 'exam' : 'assignment'
    }
  })
})

function currentScore(person: Person) {
  return person.id in overrides ? overrides[person.id]?.score ?? null : person.score
}

function scoreEndpoint(person: Person) {
  const studentID = props.kind === 'student' ? String(route.params.id) : person.id
  const resourceID = props.kind === 'student' ? person.id : String(route.params.id)
  return `/${person.resourceKind}s/${resourceID}/scores/${studentID}`
}

function scoreDetail(person: Person) {
  const score = currentScore(person)
  const prefix = person.classroom ? `${person.classroom} · ` : ''
  return score === null ? `${prefix}Awaiting score` : `${prefix}${score} / ${person.total}`
}

function scoreDate(person: Person) {
  return overrides[person.id]?.date || person.date || 'Marked just now'
}

async function save(person: Person) {
  const input = String(scores[person.id] ?? '').trim()
  const score = Number(input)
  if (input === '' || !Number.isFinite(score) || score < 0 || score > person.total) {
    toast.add({ title: 'Enter a valid score', description: `Score must be between 0 and ${person.total}.`, color: 'error' })
    return
  }
  pending[person.id] = true
  try {
    await request(scoreEndpoint(person), { method: 'PUT', body: { score } })
    overrides[person.id] = { score, date: new Date().toISOString() }
    scores[person.id] = ''
    emit('changed')
    toast.add({ title: 'Score saved', color: 'success' })
  } catch (error) {
    toast.add({ title: 'Could not save score', description: apiErrorMessage(error, 'Try again.'), color: 'error' })
  } finally {
    pending[person.id] = false
  }
}

async function reset(person: Person) {
  pending[person.id] = true
  try {
    await request(scoreEndpoint(person), { method: 'DELETE' })
    overrides[person.id] = { score: null, date: '' }
    scores[person.id] = ''
    emit('changed')
    toast.add({ title: 'Score reset', color: 'success' })
  } catch (error) {
    toast.add({ title: 'Could not reset score', description: apiErrorMessage(error, 'Try again.'), color: 'error' })
  } finally {
    pending[person.id] = false
  }
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
    <UCard
      v-for="person in people"
      :key="person.id"
    >
      <div class="flex flex-col gap-4 sm:flex-row sm:items-center">
        <UAvatar :alt="person.name" />
        <div class="min-w-0 flex-1">
          <p class="font-medium text-highlighted">
            {{ person.name }}
          </p>
          <p class="text-sm text-muted">
            {{ scoreDetail(person) }}
          </p>
          <p
            v-if="currentScore(person) !== null"
            class="mt-1 text-xs text-muted"
          >
            {{ scoreDate(person) }}
          </p>
        </div>
        <div
          v-if="currentScore(person) !== null"
          class="flex items-center gap-2"
        >
          <UBadge
            color="success"
            variant="subtle"
            label="Completed"
          />
          <UButton
            label="Reset"
            icon="i-lucide-rotate-ccw"
            color="neutral"
            variant="outline"
            :loading="pending[person.id]"
            @click="reset(person)"
          />
        </div>
        <div
          v-else
          class="flex items-center gap-2"
        >
          <UInput
            v-model="scores[person.id]"
            type="number"
            min="0"
            :max="person.total"
            placeholder="Score"
            class="w-28"
          />
          <span class="text-sm text-muted">/ {{ person.total }}</span>
          <UButton
            label="Save"
            :loading="pending[person.id]"
            @click="save(person)"
          />
        </div>
      </div>
    </UCard>
  </div>
</template>
