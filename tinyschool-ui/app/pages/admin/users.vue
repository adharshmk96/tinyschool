<script setup lang="ts">
import type { AdminUser } from '~/types/api'

definePageMeta({ layout: 'admin' })

const route = useRoute()
const router = useRouter()
const toast = useToast()
const { request } = useAdminApi()

const sortOptions = [
  { label: 'Name', value: 'name' },
  { label: 'Email', value: 'email' },
  { label: 'Joined', value: 'createdAt' },
  { label: 'Role', value: 'role' }
]

const pageSize = 10
const search = ref(String(route.query.search || ''))
const sort = ref(String(route.query.sort || 'name'))
const order = ref(route.query.order === 'desc' ? 'desc' : 'asc')
const page = ref(Math.max(1, Number(route.query.page) || 1))
const pendingIds = ref<string[]>([])
const confirmOpen = ref(false)
const target = ref<AdminUser | null>(null)
const deleteOpen = ref(false)
const deleting = ref<AdminUser | null>(null)
const currentAdmin = useState<AdminUser | null>('current-admin', () => null)
const adminOpen = ref(false)
const adminSaving = ref(false)
const adminForm = reactive({ name: '', email: '', password: '' })

const listPath = computed(() => {
  const parameters = new URLSearchParams({
    sort: sort.value,
    order: order.value,
    page: String(page.value),
    pageSize: String(pageSize)
  })
  if (search.value) parameters.set('search', search.value)
  return `/users?${parameters.toString()}`
})

const { data, status, error, refresh } = await useAsyncData(
  'admin-users',
  () => request<{ data: AdminUser[], meta?: { total?: number } }>(listPath.value),
  { watch: [listPath] }
)

const users = computed(() => data.value?.data ?? [])
const total = computed(() => Number(data.value?.meta?.total ?? users.value.length))

watch([search, sort, order], () => {
  page.value = 1
})

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

function initials(user: AdminUser) {
  return user.name.split(/\s+/).map(part => part[0]).join('').slice(0, 2).toUpperCase() || 'TS'
}

function joined(user: AdminUser) {
  if (!user.createdAt) return '—'
  return new Date(user.createdAt).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' })
}

function confirmToggle(user: AdminUser) {
  target.value = user
  confirmOpen.value = true
}

async function toggleBlocked() {
  const user = target.value
  if (!user) return
  const action = user.blocked ? 'unblock' : 'block'
  pendingIds.value.push(user.id)
  try {
    await request(`/users/${user.id}/${action}`, { method: 'POST' })
    await refresh()
    toast.add({
      title: user.blocked ? 'User unblocked' : 'User blocked',
      description: user.blocked
        ? `${user.name} can sign in again.`
        : `${user.name} was signed out and can no longer sign in.`,
      color: 'success'
    })
  } catch (requestError) {
    toast.add({
      title: `Could not ${action} the user`,
      description: apiErrorMessage(requestError, 'Please try again.'),
      color: 'error'
    })
  } finally {
    pendingIds.value = pendingIds.value.filter(id => id !== user.id)
    target.value = null
  }
}

// The API refuses to delete you or the last administrator; hide the button in
// those cases so the console never offers an action it will reject.
function canDelete(user: AdminUser) {
  if (user.id === currentAdmin.value?.id) return false
  if (user.role !== 'admin') return true
  return users.value.filter(item => item.role === 'admin').length > 1
}

function confirmDelete(user: AdminUser) {
  deleting.value = user
  deleteOpen.value = true
}

async function removeUser() {
  const user = deleting.value
  if (!user) return
  pendingIds.value.push(user.id)
  try {
    await request(`/users/${user.id}`, { method: 'DELETE' })
    await refresh()
    toast.add({
      title: 'User deleted',
      description: `${user.name} and all of their schools, students and records were removed.`,
      color: 'success'
    })
  } catch (requestError) {
    toast.add({
      title: 'Could not delete the user',
      description: apiErrorMessage(requestError, 'Please try again.'),
      color: 'error'
    })
  } finally {
    pendingIds.value = pendingIds.value.filter(id => id !== user.id)
    deleting.value = null
  }
}

function openAddAdmin() {
  adminForm.name = ''
  adminForm.email = ''
  adminForm.password = ''
  adminOpen.value = true
}

// Only a signed-in administrator can reach this endpoint; the public setup page
// stays closed once the first administrator exists.
async function addAdmin() {
  if (!adminForm.email || adminForm.password.length < 8) {
    toast.add({
      title: 'Check the details',
      description: 'An email and a password of at least 8 characters are required.',
      color: 'error'
    })
    return
  }
  adminSaving.value = true
  try {
    await request('/admins', {
      method: 'POST',
      body: { name: adminForm.name.trim(), email: adminForm.email.trim(), password: adminForm.password }
    })
    adminOpen.value = false
    await refresh()
    toast.add({ title: 'Administrator added', description: `${adminForm.email} can now sign in to the back office.`, color: 'success' })
  } catch (requestError) {
    toast.add({
      title: 'Could not add the administrator',
      description: apiErrorMessage(requestError, 'Please try again.'),
      color: 'error'
    })
  } finally {
    adminSaving.value = false
  }
}

useSeoMeta({ title: 'Admin users' })
</script>

<template>
  <div class="page-shell">
    <PageHeading
      eyebrow="Back office"
      title="Users"
      description="Every account on this instance. Search, sort and control who can sign in."
    >
      <template #actions>
        <UButton
          icon="i-lucide-shield-plus"
          label="Add administrator"
          @click="openAddAdmin"
        />
      </template>
    </PageHeading>

    <UCard class="mt-8">
      <div class="grid gap-3 md:grid-cols-[minmax(240px,1fr)_190px_auto]">
        <UInput
          v-model="search"
          icon="i-lucide-search"
          placeholder="Search by name or email…"
          aria-label="Search users"
        />
        <USelect
          v-model="sort"
          :items="sortOptions"
          value-key="value"
          aria-label="Sort by"
        />
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
      title="Could not load users"
      :description="error.message"
      :actions="[{ label: 'Try again', onClick: () => refresh() }]"
    />

    <div
      v-else-if="status === 'pending'"
      class="mt-5 space-y-3"
    >
      <USkeleton
        v-for="index in 5"
        :key="index"
        class="h-20 rounded-xl"
      />
    </div>

    <div
      v-else-if="users.length"
      class="mt-5 space-y-3"
    >
      <UCard
        v-for="user in users"
        :key="user.id"
      >
        <div class="flex flex-wrap items-center gap-4">
          <UAvatar
            :text="initials(user)"
            size="lg"
          />
          <div class="min-w-0 flex-1">
            <div class="flex flex-wrap items-center gap-2">
              <h2 class="truncate font-semibold text-highlighted">
                {{ user.name }}
              </h2>
              <UBadge
                v-if="user.role === 'admin'"
                label="Admin"
                color="warning"
                variant="subtle"
                size="sm"
              />
              <UBadge
                v-if="user.blocked"
                label="Blocked"
                color="error"
                variant="subtle"
                size="sm"
              />
            </div>
            <p class="truncate text-sm text-muted">
              {{ user.email }}
            </p>
            <p class="mt-1 text-xs text-dimmed">
              Joined {{ joined(user) }}
            </p>
          </div>
          <div class="flex flex-wrap gap-2">
            <UButton
              v-if="user.role !== 'admin'"
              :icon="user.blocked ? 'i-lucide-user-check' : 'i-lucide-user-x'"
              :label="user.blocked ? 'Unblock' : 'Block'"
              :color="user.blocked ? 'neutral' : 'error'"
              variant="outline"
              :loading="pendingIds.includes(user.id)"
              @click="confirmToggle(user)"
            />
            <UButton
              v-if="canDelete(user)"
              icon="i-lucide-trash-2"
              color="error"
              variant="ghost"
              :aria-label="`Delete ${user.name}`"
              :loading="pendingIds.includes(user.id)"
              @click="confirmDelete(user)"
            />
          </div>
        </div>
      </UCard>
    </div>

    <EmptyState
      v-else
      class="mt-5"
      icon="i-lucide-users"
      title="No users found"
      description="Try a different search term."
    />

    <div
      v-if="total > pageSize"
      class="mt-8 flex justify-center"
    >
      <UPagination
        v-model:page="page"
        :total="total"
        :items-per-page="pageSize"
      />
    </div>

    <ConfirmDialog
      v-model="deleteOpen"
      title="Delete this user?"
      :description="`${deleting?.name} (${deleting?.email}) and every school, academic year, student, class, assignment and exam they own will be permanently deleted.`"
      confirm-label="Delete user"
      @confirm="removeUser"
    />

    <UModal
      v-model:open="adminOpen"
      title="Add administrator"
      description="The new account can sign in to this back office immediately."
    >
      <template #body>
        <form
          class="space-y-4"
          @submit.prevent="addAdmin"
        >
          <UFormField label="Name">
            <UInput
              v-model="adminForm.name"
              placeholder="Administrator"
              class="w-full"
            />
          </UFormField>
          <UFormField
            label="Email"
            required
          >
            <UInput
              v-model="adminForm.email"
              type="email"
              class="w-full"
            />
          </UFormField>
          <UFormField
            label="Password"
            required
          >
            <UInput
              v-model="adminForm.password"
              type="password"
              autocomplete="new-password"
              class="w-full"
            />
            <template #help>
              At least 8 characters.
            </template>
          </UFormField>
          <div class="flex justify-end gap-2 pt-2">
            <UButton
              label="Cancel"
              color="neutral"
              variant="outline"
              @click="adminOpen = false"
            />
            <UButton
              type="submit"
              label="Add administrator"
              :loading="adminSaving"
            />
          </div>
        </form>
      </template>
    </UModal>

    <!-- Blocking is reversible, so this uses its own dialog rather than the
         delete-flavoured ConfirmDialog. -->
    <UModal
      v-model:open="confirmOpen"
      :title="target?.blocked ? 'Unblock this user?' : 'Block this user?'"
    >
      <template #body>
        <p class="text-sm text-muted">
          {{ target?.blocked
            ? `${target?.name} will be able to sign in again.`
            : `${target?.name} will be signed out immediately and denied sign-in. You can unblock them at any time.` }}
        </p>
      </template>
      <template #footer>
        <div class="flex w-full justify-end gap-2">
          <UButton
            color="neutral"
            variant="ghost"
            label="Cancel"
            @click="confirmOpen = false"
          />
          <UButton
            :color="target?.blocked ? 'primary' : 'error'"
            :label="target?.blocked ? 'Unblock' : 'Block'"
            @click="confirmOpen = false; toggleBlocked()"
          />
        </div>
      </template>
    </UModal>
  </div>
</template>
