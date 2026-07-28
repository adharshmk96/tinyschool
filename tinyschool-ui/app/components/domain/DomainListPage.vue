<script setup lang="ts">
type Item = Record<string, unknown>

const props = defineProps<{
  title: string
  singular: string
  endpoint: string
  icon: string
  detailBase: string
  description: string
  fields: Array<{ key: string, label: string, type?: string, options?: string[] }>
  cardFields: Array<{ key: string, label: string }>
  sortOptions: Array<{ label: string, value: string }>
}>()

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { baseURL } = useApi()
const modalOpen = ref(false)
const editing = ref<Item | null>(null)
const form = reactive<Record<string, string>>({})

const search = ref(String(route.query.search || ''))
const sort = ref(String(route.query.sort || props.sortOptions[0]?.value || 'name'))
const order = ref(route.query.order === 'asc' ? 'asc' : 'desc')
const page = ref(Math.max(1, Number(route.query.page) || 1))
const pageSize = 8

const query = computed(() => ({
  search: search.value || undefined,
  sort: sort.value,
  order: order.value,
  page: page.value,
  pageSize
}))
const apiURL = computed(() => `${baseURL}${props.endpoint.replace(/^\/api\/v1/, '')}`)

const { data, status, error, refresh } = await useFetch<{ data: Item[], meta?: { total?: number } }>(
  apiURL,
  { query, watch: [query] }
)

const items = computed(() => Array.isArray(data.value?.data) ? data.value.data : [])
const total = computed(() => Number(data.value?.meta?.total ?? items.value.length))

watch([search, sort, order, page], () => {
  router.replace({
    query: {
      ...(search.value ? { search: search.value } : {}),
      sort: sort.value,
      order: order.value,
      ...(page.value > 1 ? { page: String(page.value) } : {})
    }
  })
})

watch([search, sort, order], () => {
  page.value = 1
})

function value(item: Item, key: string) {
  const raw = key.split('.').reduce<unknown>((current, part) => {
    if (current && typeof current === 'object')
      return (current as Item)[part]
    return undefined
  }, item)
  if (Array.isArray(raw))
    return raw.join(', ')
  if (raw === null || raw === undefined || raw === '')
    return '—'
  return String(raw)
}

function itemId(item: Item) {
  return String(item.id ?? item.slug ?? '')
}

function itemTitle(item: Item) {
  return String(item.name ?? item.fullName ?? (`${item.firstName ?? ''} ${item.lastName ?? ''}`.trim() || props.singular))
}

function openCreate() {
  editing.value = null
  for (const field of props.fields)
    form[field.key] = ''
  modalOpen.value = true
}

function openEdit(item: Item) {
  editing.value = item
  for (const field of props.fields)
    form[field.key] = value(item, field.key) === '—' ? '' : value(item, field.key)
  modalOpen.value = true
}

function save() {
  modalOpen.value = false
  toast.add({
    title: editing.value ? `${props.singular} updated` : `${props.singular} created`,
    description: 'Saved in this placeholder UI session.',
    color: 'success'
  })
}

function remove(item: Item) {
  if (!window.confirm(`Delete ${itemTitle(item)}? This action cannot be undone.`))
    return
  toast.add({ title: `${props.singular} deleted`, description: 'Placeholder action completed.', color: 'success' })
}
</script>

<template>
  <div class="page-shell">
    <AppBreadcrumb
      :items="[
        { label: 'Overview', to: '/dashboard', icon: 'i-lucide-layout-dashboard' },
        { label: title, icon }
      ]"
    />
    <div class="flex flex-col gap-5 sm:flex-row sm:items-start sm:justify-between">
      <div>
        <h1 class="text-2xl font-bold tracking-tight text-highlighted sm:text-3xl">
          {{ title }}
        </h1>
        <p class="mt-2 max-w-2xl text-sm leading-6 text-muted sm:text-base">
          {{ description }}
        </p>
      </div>
      <UButton
        icon="i-lucide-plus"
        :label="`Create ${singular}`"
        class="shrink-0"
        @click="openCreate"
      />
    </div>

    <UCard class="mt-8">
      <div class="grid gap-3 md:grid-cols-[minmax(240px,1fr)_190px_auto]">
        <UInput
          v-model="search"
          icon="i-lucide-search"
          :placeholder="`Search ${title.toLowerCase()}…`"
          aria-label="Search"
        />
        <USelect v-model="sort" :items="sortOptions" value-key="value" aria-label="Sort by" />
        <UButton
          color="neutral"
          variant="outline"
          :icon="order === 'asc' ? 'i-lucide-arrow-up' : 'i-lucide-arrow-down'"
          :label="order === 'asc' ? 'Ascending' : 'Descending'"
          @click="order = order === 'asc' ? 'desc' : 'asc'"
        />
      </div>
    </UCard>

    <UAlert
      v-if="error"
      class="mt-5"
      color="error"
      icon="i-lucide-circle-alert"
      title="Could not load data"
      :description="error.message"
      :actions="[{ label: 'Try again', onClick: () => refresh() }]"
    />

    <div
      v-if="status === 'pending'"
      class="mt-5 grid gap-5 lg:grid-cols-2"
    >
      <USkeleton v-for="index in 4" :key="index" class="h-48 rounded-xl" />
    </div>

    <div
      v-else-if="items.length"
      class="mt-5 grid gap-5 lg:grid-cols-2"
    >
      <UCard
        v-for="item in items"
        :key="itemId(item)"
        class="relative transition hover:-translate-y-0.5 hover:shadow-md hover:ring-1 hover:ring-primary/40"
      >
        <NuxtLink
          :to="`${detailBase}/${itemId(item)}`"
          class="absolute inset-0 rounded-xl"
          :aria-label="`Open ${itemTitle(item)}`"
        />
        <div class="relative pointer-events-none flex items-start gap-5">
          <UAvatar :icon="icon" size="lg" class="shrink-0" />
          <div class="min-w-0 flex-1">
            <h2 class="truncate font-semibold text-highlighted">
              {{ itemTitle(item) }}
            </h2>
            <dl class="mt-3 grid grid-cols-2 gap-x-4 gap-y-3 text-sm">
              <div v-for="field in cardFields" :key="field.key">
                <dt class="text-xs text-muted">
                  {{ field.label }}
                </dt>
                <dd class="mt-0.5 truncate text-default">
                  {{ value(item, field.key) }}
                </dd>
              </div>
            </dl>
          </div>
          <div class="pointer-events-auto relative flex gap-1">
            <UButton icon="i-lucide-pencil" color="neutral" variant="ghost" aria-label="Edit" @click.stop="openEdit(item)" />
            <UButton icon="i-lucide-trash-2" color="error" variant="ghost" aria-label="Delete" @click.stop="remove(item)" />
          </div>
        </div>
      </UCard>
    </div>

    <UCard
      v-else
      class="mt-5 py-12 text-center"
    >
      <UIcon :name="icon" class="mx-auto size-9 text-dimmed" />
      <h2 class="mt-3 font-medium">
        No {{ title.toLowerCase() }} found
      </h2>
      <p class="mt-1 text-sm text-muted">
        Try another search or create the first {{ singular.toLowerCase() }}.
      </p>
    </UCard>

    <div
      v-if="total > pageSize"
      class="mt-8 flex justify-center"
    >
      <UPagination v-model:page="page" :total="total" :items-per-page="pageSize" />
    </div>

    <UModal v-model:open="modalOpen" :title="`${editing ? 'Edit' : 'Create'} ${singular}`">
      <template #body>
        <form class="space-y-4" @submit.prevent="save">
          <UFormField v-for="field in fields" :key="field.key" :label="field.label" required>
            <UTextarea v-if="field.type === 'textarea'" v-model="form[field.key]" class="w-full" />
            <USelect
              v-else-if="field.options"
              v-model="form[field.key]"
              :items="field.options"
              class="w-full"
            />
            <UInput v-else v-model="form[field.key]" :type="field.type || 'text'" class="w-full" />
          </UFormField>
          <div class="flex justify-end gap-2 pt-2">
            <UButton label="Cancel" color="neutral" variant="outline" @click="modalOpen = false" />
            <UButton type="submit" :label="editing ? 'Save changes' : `Create ${singular}`" />
          </div>
        </form>
      </template>
    </UModal>
  </div>
</template>
