<script setup lang="ts">
const props = defineProps<{
  kind: 'behaviour' | 'notes'
  items?: unknown
}>()

const toast = useToast()
const note = ref('')
const sentiment = ref('positive')
const entries = ref<Array<{ id: string | number, tone: string, note: string, date: string }>>([])

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

function addEntry() {
  const text = note.value.trim()
  if (!text) {
    toast.add({ title: 'Add a note first', color: 'warning' })
    return
  }
  entries.value.unshift({ id: Date.now(), tone: sentiment.value, note: text, date: 'Just now' })
  note.value = ''
  toast.add({ title: props.kind === 'notes' ? 'Note added' : 'Behaviour logged', color: 'success' })
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
      <UCard v-for="entry in entries" :key="entry.id">
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
          <UButton icon="i-lucide-pencil" color="neutral" variant="ghost" aria-label="Edit log" />
          <UButton icon="i-lucide-trash-2" color="error" variant="ghost" aria-label="Delete log" />
        </div>
      </UCard>
    </div>
    <UCard>
      <h2 class="font-semibold">
        Add {{ kind === 'notes' ? 'note' : 'behaviour log' }}
      </h2>
      <div v-if="kind === 'behaviour'" class="mt-4">
        <USelect
          v-model="sentiment"
          :items="[
            { label: 'Positive', value: 'positive' },
            { label: 'Needs attention', value: 'attention' },
            { label: 'Incident', value: 'incident' }
          ]"
          value-key="value"
          class="w-full"
        />
      </div>
      <UTextarea v-model="note" placeholder="Write a short note…" class="mt-4 w-full" :rows="5" />
      <UButton label="Add log" icon="i-lucide-plus" class="mt-4 w-full justify-center" @click="addEntry" />
    </UCard>
  </div>
</template>
