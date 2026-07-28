<script setup lang="ts">
import type { User } from '~/types/api'

const route = useRoute()
const toast = useToast()
const mobileOpen = ref(false)
const logoutPending = ref(false)
const currentUser = useState<User | null>('current-user', () => null)
const { getItem, post } = useApi()
const { schools, academicYears, selectedSchoolId, selectedYearId, selectedSchool, selectedYear, loading, load } = useSchoolContext()

const navigation = [
  { label: 'Overview', to: '/dashboard', icon: 'i-lucide-layout-dashboard' },
  { label: 'Students', to: '/dashboard/students', icon: 'i-lucide-users' },
  { label: 'My Classes', to: '/dashboard/classes', icon: 'i-lucide-panels-top-left' },
  { label: 'Assignments', to: '/dashboard/assignments', icon: 'i-lucide-clipboard-check' },
  { label: 'Exams', to: '/dashboard/exams', icon: 'i-lucide-file-chart-column' }
]

const userItems = [
  [{ label: 'Settings', icon: 'i-lucide-settings', to: '/dashboard/settings/account' }],
  [{ label: 'Log out', icon: 'i-lucide-log-out', onSelect: logout }]
]

function isActive(path: string) {
  return path === '/dashboard' ? route.path === path : route.path.startsWith(path)
}

async function logout() {
  if (logoutPending.value) return
  logoutPending.value = true
  try {
    await post('/auth/logout')
  } catch (error) {
    toast.add({
      title: 'Server logout failed',
      description: apiErrorMessage(error, 'Your local sign-in marker was still cleared.'),
      color: 'error'
    })
  } finally {
    useCookie('tiny-school-authenticated').value = null
    currentUser.value = null
    logoutPending.value = false
    await navigateTo('/login')
  }
}

watch(() => route.fullPath, () => {
  mobileOpen.value = false
})

onMounted(async () => {
  try {
    const [user] = await Promise.all([
      getItem<User>('/me'),
      load()
    ])
    currentUser.value = user.data
  } catch {
    toast.add({
      title: 'Could not load school context',
      description: 'Make sure the local API is running.',
      color: 'error'
    })
  }
})
</script>

<template>
  <div class="min-h-screen bg-elevated/40 lg:pl-(--app-sidebar)">
    <aside class="fixed inset-y-0 left-0 z-30 hidden w-(--app-sidebar) border-r border-default bg-default lg:flex lg:flex-col">
      <div class="flex h-16 items-center border-b border-default px-5">
        <NuxtLink
          to="/dashboard"
          aria-label="Tiny School dashboard"
        >
          <AppLogo />
        </NuxtLink>
      </div>

      <div class="space-y-3 border-b border-default p-4">
        <USelect
          v-model="selectedSchoolId"
          aria-label="Select school"
          value-key="id"
          label-key="name"
          :items="schools"
          :loading="loading"
          icon="i-lucide-school"
          class="w-full"
          placeholder="Select a school"
        />
        <USelect
          v-model="selectedYearId"
          aria-label="Select academic year"
          value-key="id"
          label-key="name"
          :items="academicYears"
          :loading="loading"
          icon="i-lucide-calendar-range"
          class="w-full"
          placeholder="Select a year"
        />
      </div>

      <nav
        class="flex-1 space-y-1 overflow-y-auto p-3"
        aria-label="Main navigation"
      >
        <p class="px-3 pb-1 pt-2 text-xs font-semibold uppercase tracking-wider text-dimmed">
          Manage
        </p>
        <UButton
          v-for="item in navigation"
          :key="item.to"
          :to="item.to"
          :icon="item.icon"
          :label="item.label"
          color="neutral"
          :variant="isActive(item.to) ? 'soft' : 'ghost'"
          class="w-full justify-start"
        />
      </nav>

      <div class="border-t border-default p-3">
        <UDropdownMenu :items="userItems">
          <UButton
            color="neutral"
            variant="ghost"
            class="w-full justify-start"
          >
            <UAvatar
              :text="currentUser?.name.split(/\s+/).map(part => part[0]).join('').slice(0, 2).toUpperCase() || 'TS'"
              size="sm"
            />
            <span class="min-w-0 flex-1 text-left">
              <span class="block truncate text-sm font-semibold">{{ currentUser?.name || 'Tiny School user' }}</span>
              <span class="block truncate text-xs font-normal text-muted">{{ currentUser?.email || 'Loading account…' }}</span>
            </span>
            <UIcon
              name="i-lucide-chevrons-up-down"
              class="size-4 text-muted"
            />
          </UButton>
        </UDropdownMenu>
      </div>
    </aside>

    <header class="sticky top-0 z-20 flex h-16 items-center gap-3 border-b border-default bg-default/90 px-4 backdrop-blur lg:px-6">
      <UButton
        icon="i-lucide-menu"
        color="neutral"
        variant="ghost"
        aria-label="Open navigation"
        class="lg:hidden"
        @click="mobileOpen = true"
      />
      <div class="min-w-0 flex-1">
        <p class="truncate text-sm font-semibold">
          {{ selectedSchool?.name || 'Tiny School' }}
        </p>
        <p class="truncate text-xs text-muted">
          {{ selectedYear?.name || 'No academic year selected' }}
        </p>
      </div>
      <UColorModeButton />
    </header>

    <main>
      <slot />
    </main>

    <USlideover
      v-model:open="mobileOpen"
      side="left"
      title="Navigation"
    >
      <template #body>
        <div class="space-y-3">
          <AppLogo />
          <USelect
            v-model="selectedSchoolId"
            value-key="id"
            label-key="name"
            :items="schools"
            icon="i-lucide-school"
            class="w-full"
            placeholder="Select a school"
          />
          <USelect
            v-model="selectedYearId"
            value-key="id"
            label-key="name"
            :items="academicYears"
            icon="i-lucide-calendar-range"
            class="w-full"
            placeholder="Select a year"
          />
          <nav class="space-y-1 pt-3">
            <UButton
              v-for="item in navigation"
              :key="item.to"
              :to="item.to"
              :icon="item.icon"
              :label="item.label"
              color="neutral"
              :variant="isActive(item.to) ? 'soft' : 'ghost'"
              class="w-full justify-start"
            />
          </nav>
        </div>
      </template>
      <template #footer>
        <div class="w-full space-y-2">
          <div class="flex items-center gap-3 px-2 pb-2">
            <UAvatar
              :text="currentUser?.name.split(/\s+/).map(part => part[0]).join('').slice(0, 2).toUpperCase() || 'TS'"
              size="sm"
            />
            <span class="min-w-0">
              <span class="block truncate text-sm font-semibold">{{ currentUser?.name || 'Tiny School user' }}</span>
              <span class="block truncate text-xs text-muted">{{ currentUser?.email || 'Loading account…' }}</span>
            </span>
          </div>
          <UButton
            to="/dashboard/settings/account"
            icon="i-lucide-settings"
            label="Settings"
            color="neutral"
            variant="ghost"
            class="w-full justify-start"
          />
          <UButton
            icon="i-lucide-log-out"
            label="Log out"
            color="error"
            variant="ghost"
            class="w-full justify-start"
            :loading="logoutPending"
            @click="logout"
          />
        </div>
      </template>
    </USlideover>
  </div>
</template>
