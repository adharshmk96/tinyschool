<script setup lang="ts">
const props = defineProps<{
  items?: unknown
  kind: string
}>()

const normalized = computed(() => {
  if (!Array.isArray(props.items))
    return []
  return props.items.map((entry, index) => {
    const item = entry as Record<string, unknown>
    return {
      id: String(item.id ?? index),
      name: String(item.name ?? `Item ${index + 1}`)
    }
  })
})

const base = computed(() => `/dashboard/${props.kind}`)
const icon = computed(() => props.kind === 'students'
  ? 'i-lucide-graduation-cap'
  : props.kind === 'exams' ? 'i-lucide-file-check-2' : 'i-lucide-clipboard-check')
</script>

<template>
  <div v-if="normalized.length" class="grid gap-4 lg:grid-cols-2">
    <UCard v-for="item in normalized" :key="item.id" class="relative">
      <NuxtLink :to="`${base}/${item.id}`" class="absolute inset-0 rounded-xl" :aria-label="`Open ${item.name}`" />
      <div class="relative pointer-events-none flex items-center gap-3">
        <UAvatar :icon="icon" />
        <div class="flex-1">
          <p class="font-medium text-highlighted">
            {{ item.name }}
          </p>
          <p class="text-sm text-muted">
            Linked to this class
          </p>
        </div>
        <UIcon name="i-lucide-arrow-up-right" class="size-4 text-muted" />
      </div>
    </UCard>
  </div>
  <TabPlaceholder
    v-else
    :icon="icon"
    :title="`No linked ${kind} yet`"
    :description="`Linked ${kind} will appear here with quick access to their details.`"
  />
</template>
