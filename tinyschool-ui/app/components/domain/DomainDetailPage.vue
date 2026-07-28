<script setup lang="ts">
type Item = Record<string, unknown>

const props = defineProps<{
  title: string
  endpoint: string
  backTo: string
  backLabel: string
  icon: string
  fields: Array<{ key: string, label: string }>
  editFields?: Array<{ key: string, label: string, type?: string, options?: string[] }>
  tabs?: Array<{ label: string, value: string, icon: string }>
  activeTab?: string
}>()

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { request } = useApi()
const editOpen = ref(false)
const saving = ref(false)
const form = reactive<Record<string, string>>({})
const id = computed(() => String(route.params.id))
const apiPath = computed(() => `${props.endpoint.replace(/^\/api\/v1/, '')}/${id.value}`)
const { data, status, error, refresh } = await useAsyncData(
  `domain-detail-${props.endpoint}-${id.value}`,
  () => request<{ data: Item }>(apiPath.value),
  { watch: [apiPath] }
)
const item = computed<Item>(() => data.value?.data ?? {})
const displayTitle = computed(() => String(item.value.name ?? item.value.fullName ?? (`${item.value.firstName ?? ''} ${item.value.lastName ?? ''}`.trim() || props.title)))

function rawValue(key: string) {
  return key.split('.').reduce<unknown>((current, part) => {
    if (current && typeof current === 'object')
      return (current as Item)[part]
    return undefined
  }, item.value)
}

function value(key: string) {
  const raw = rawValue(key)
  if (Array.isArray(raw))
    return raw.join(', ')
  return raw === undefined || raw === null || raw === '' ? '—' : String(raw)
}

function setTab(tab: string) {
  router.push(`${props.backTo}/${id.value}/${tab}`)
}

function openEdit() {
  for (const field of props.editFields ?? []) {
    const current = rawValue(field.key)
    form[field.key] = current === undefined || current === null ? '' : String(current)
  }
  editOpen.value = true
}

async function saveEdit() {
  const body: Record<string, unknown> = {}
  for (const field of props.editFields ?? [])
    body[field.key] = field.type === 'number' ? Number(form[field.key]) : form[field.key]
  saving.value = true
  try {
    await request(`${props.endpoint.replace(/^\/api\/v1/, '')}/${id.value}`, { method: 'PATCH', body })
    await refresh()
    editOpen.value = false
    toast.add({ title: `${props.title} updated`, color: 'success' })
  } catch (error) {
    toast.add({ title: `Could not update ${props.title.toLowerCase()}`, description: apiErrorMessage(error, 'Try again.'), color: 'error' })
  } finally {
    saving.value = false
  }
}
</script>

<template>
  <div class="page-shell">
    <AppBreadcrumb
      :items="[
        { label: 'Overview', to: '/dashboard', icon: 'i-lucide-layout-dashboard' },
        { label: backLabel, to: backTo, icon },
        { label: displayTitle }
      ]"
    />

    <UAlert
      v-if="error"
      class="mb-6"
      color="error"
      icon="i-lucide-circle-alert"
      title="Could not load details"
      :description="error.message"
      :actions="[{ label: 'Try again', onClick: () => refresh() }]"
    />
    <USkeleton
      v-if="status === 'pending'"
      class="h-56 rounded-xl"
    />
    <UCard v-else>
      <div class="flex flex-col gap-6 sm:flex-row sm:items-start">
        <UAvatar
          :icon="icon"
          size="3xl"
        />
        <div class="min-w-0 flex-1">
          <div class="flex flex-col gap-4 sm:flex-row sm:items-start sm:justify-between">
            <div>
              <h1 class="text-2xl font-bold tracking-tight text-highlighted sm:text-3xl">
                {{ displayTitle }}
              </h1>
              <p class="mt-2 text-sm leading-6 text-muted">
                {{ title }} details for the active academic year.
              </p>
            </div>
            <UButton
              v-if="editFields?.length"
              label="Edit"
              icon="i-lucide-pencil"
              color="neutral"
              variant="outline"
              @click="openEdit"
            />
          </div>
          <dl class="mt-7 grid gap-4 sm:grid-cols-2 lg:grid-cols-4">
            <div
              v-for="field in fields"
              :key="field.key"
              class="rounded-lg bg-elevated/60 px-4 py-3"
            >
              <dt class="text-xs text-muted">
                {{ field.label }}
              </dt>
              <dd class="mt-1 text-sm font-medium text-default">
                {{ value(field.key) }}
              </dd>
            </div>
          </dl>
        </div>
      </div>
    </UCard>

    <div
      v-if="tabs?.length"
      class="mt-7 overflow-x-auto border-b border-default pb-px"
    >
      <UTabs
        :model-value="activeTab"
        :items="tabs"
        value-key="value"
        :content="false"
        @update:model-value="setTab(String($event))"
      />
    </div>

    <div class="mt-7">
      <slot
        :item="item"
        :refresh="refresh"
      />
    </div>

    <UModal
      v-model:open="editOpen"
      :title="`Edit ${title.toLowerCase()}`"
    >
      <template #body>
        <form
          id="detail-edit-form"
          class="space-y-4"
          @submit.prevent="saveEdit"
        >
          <UFormField
            v-for="field in editFields"
            :key="field.key"
            :label="field.label"
          >
            <USelect
              v-if="field.options"
              v-model="form[field.key]"
              :items="field.options"
              class="w-full"
            />
            <UInput
              v-else
              v-model="form[field.key]"
              :type="field.type || 'text'"
              class="w-full"
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
            @click="editOpen = false"
          />
          <UButton
            form="detail-edit-form"
            type="submit"
            label="Save changes"
            :loading="saving"
          />
        </div>
      </template>
    </UModal>
  </div>
</template>
