<script setup lang="ts">
const props = defineProps<{
  kind: 'behaviour' | 'notes'
  items?: unknown
}>()

const emit = defineEmits<{
  changed: []
}>()

type Entry = { id: string, tone: string, note: string, date: string }

const route = useRoute()
const toast = useToast()
const { request } = useApi()
const note = ref('')
const sentiment = ref('positive')
const entries = ref<Entry[]>([])
const adding = ref(false)
const deleting = reactive<Record<string, boolean>>({})
const deleteOpen = ref(false)
const deletingEntry = ref<Entry | null>(null)

watchEffect(() => {
  const source = Array.isArray(props.items) ? props.items : []
  entries.value = source.length
    ? source.map((entry, index) => {
        const item = entry as Record<string, unknown>
        return {
          id: String(item.id ?? index),
          tone: String(item.type ?? 'positive'),
          note: String(item.note ?? ''),
          date: String(item.createdAt ?? '')
        }
      })
    : []
})

async function addEntry() {
  const text = note.value.trim()
  if (!text) {
    toast.add({ title: 'Add a note first', color: 'warning' })
    return
  }
  const body = props.kind === 'notes'
    ? { note: text }
    : { type: sentiment.value, note: text }
  adding.value = true
  try {
    const response = await request<{ data: Record<string, unknown> }>(
      `/students/${String(route.params.id)}/${props.kind}`,
      { method: 'POST', body }
    )
    entries.value.unshift({
      id: String(response.data.id),
      tone: String(response.data.type ?? sentiment.value),
      note: String(response.data.note ?? text),
      date: String(response.data.createdAt ?? 'Just now')
    })
    note.value = ''
    emit('changed')
    toast.add({ title: props.kind === 'notes' ? 'Note added' : 'Behaviour logged', color: 'success' })
  } catch (error) {
    toast.add({
      title: props.kind === 'notes' ? 'Could not add note' : 'Could not log behaviour',
      description: apiErrorMessage(error, 'Try again.'),
      color: 'error'
    })
  } finally {
    adding.value = false
  }
}

function requestDelete(entry: Entry) {
  deletingEntry.value = entry
  deleteOpen.value = true
}

async function deleteEntry() {
  const entry = deletingEntry.value
  if (!entry)
    return

  deleting[entry.id] = true
  try {
    await request(`/students/${String(route.params.id)}/${props.kind}/${entry.id}`, { method: 'DELETE' })
    entries.value = entries.value.filter(item => item.id !== entry.id)
    deletingEntry.value = null
    emit('changed')
    toast.add({ title: props.kind === 'notes' ? 'Note deleted' : 'Behaviour log deleted', color: 'success' })
  } catch (error) {
    toast.add({ title: 'Could not delete log', description: apiErrorMessage(error, 'Try again.'), color: 'error' })
  } finally {
    deleting[entry.id] = false
  }
}
</script>

<template>
  <div class="grid gap-4 lg:grid-cols-[minmax(0,2fr)_minmax(260px,1fr)]">
    <div class="space-y-3">
      <TabPlaceholder
        v-if="!entries.length"
        :icon="kind === 'notes' ? 'i-lucide-notebook-pen' : 'i-lucide-heart-handshake'"
        :title="kind === 'notes' ? 'No notes available' : 'No behaviour logs available'"
        :description="kind === 'notes'
          ? 'Student notes will appear here after the first note is added.'
          : 'Positive moments, concerns and incidents will appear here after the first log is added.'"
      />
      <UCard
        v-for="entry in entries"
        :key="entry.id"
      >
        <div class="flex items-start gap-3">
          <UIcon
            :name="entry.tone === 'positive' ? 'i-lucide-circle-check' : 'i-lucide-circle-alert'"
            :class="entry.tone === 'positive' ? 'text-success' : 'text-warning'"
            class="mt-0.5 size-5 shrink-0"
          />
          <div class="min-w-0 flex-1">
            <p class="text-sm text-default">
              {{ entry.note }}
            </p>
            <p class="mt-2 text-xs text-muted">
              {{ entry.date }}
            </p>
          </div>
          <UButton
            icon="i-lucide-trash-2"
            color="error"
            variant="ghost"
            aria-label="Delete log"
            :loading="deleting[entry.id]"
            @click="requestDelete(entry)"
          />
        </div>
      </UCard>
    </div>
    <UCard>
      <h2 class="font-semibold">
        Add {{ kind === 'notes' ? 'note' : 'behaviour log' }}
      </h2>
      <div
        v-if="kind === 'behaviour'"
        class="mt-4"
      >
        <USelect
          v-model="sentiment"
          :items="[
            { label: 'Positive', value: 'positive' },
            { label: 'Needs attention', value: 'need_attention' },
            { label: 'Incident', value: 'incident' }
          ]"
          value-key="value"
          class="w-full"
        />
      </div>
      <UTextarea
        v-model="note"
        placeholder="Write a short note…"
        class="mt-4 w-full"
        :rows="5"
      />
      <UButton
        label="Add log"
        icon="i-lucide-plus"
        class="mt-4 w-full justify-center"
        :loading="adding"
        @click="addEntry"
      />
    </UCard>
    <ConfirmDialog
      v-model="deleteOpen"
      :title="kind === 'notes' ? 'Delete note?' : 'Delete behaviour log?'"
      :description="kind === 'notes' ? 'Delete this student note?' : 'Delete this behaviour log?'"
      @confirm="deleteEntry"
    />
  </div>
</template>
