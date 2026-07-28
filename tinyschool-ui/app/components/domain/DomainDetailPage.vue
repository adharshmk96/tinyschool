<script setup lang="ts">
type Item = Record<string, unknown>

const props = defineProps<{
  title: string
  endpoint: string
  backTo: string
  backLabel: string
  icon: string
  fields: Array<{ key: string, label: string }>
  tabs?: Array<{ label: string, value: string, icon: string }>
  activeTab?: string
}>()

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { baseURL } = useApi()
const id = computed(() => String(route.params.id))
const apiURL = computed(() => `${baseURL}${props.endpoint.replace(/^\/api\/v1/, '')}/${id.value}`)
const { data, status, error, refresh } = await useFetch<{ data: Item }>(apiURL)
const item = computed<Item>(() => data.value?.data ?? {})
const displayTitle = computed(() => String(item.value.name ?? item.value.fullName ?? (`${item.value.firstName ?? ''} ${item.value.lastName ?? ''}`.trim() || props.title)))

function value(key: string) {
  const raw = key.split('.').reduce<unknown>((current, part) => {
    if (current && typeof current === 'object')
      return (current as Item)[part]
    return undefined
  }, item.value)
  if (Array.isArray(raw))
    return raw.join(', ')
  return raw === undefined || raw === null || raw === '' ? '—' : String(raw)
}

function setTab(tab: string) {
  router.push(`${props.backTo}/${id.value}/${tab}`)
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
        <UAvatar :icon="icon" size="3xl" />
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
              label="Edit"
              icon="i-lucide-pencil"
              color="neutral"
              variant="outline"
              @click="toast.add({ title: `Edit ${title}`, description: 'Placeholder form opened.', color: 'success' })"
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
      <slot :item="item" />
    </div>
  </div>
</template>
