<script setup lang="ts">
import type { AdminUser } from '~/types/api'

const route = useRoute()
const toast = useToast()
const { getItem, request } = useAdminApi()
const currentAdmin = useState<AdminUser | null>('current-admin', () => null)
const mobileOpen = ref(false)
const logoutPending = ref(false)

const navigation = [
  { label: 'Overview', to: '/admin', icon: 'i-lucide-shield' },
  { label: 'Users', to: '/admin/users', icon: 'i-lucide-users' },
  { label: 'SMTP', to: '/admin/smtp', icon: 'i-lucide-mail' },
  { label: 'Backups', to: '/admin/backups', icon: 'i-lucide-database-backup' }
]

function isActive(path: string) {
  return path === '/admin' ? route.path === path : route.path.startsWith(path)
}

async function logout() {
  if (logoutPending.value) return
  logoutPending.value = true
  try {
    await request('/logout', { method: 'POST' })
  } catch (error) {
    toast.add({
      title: 'Server logout failed',
      description: apiErrorMessage(error, 'Your local sign-in marker was still cleared.'),
      color: 'error'
    })
  } finally {
    useCookie(adminAuthCookieName).value = null
    currentAdmin.value = null
    logoutPending.value = false
    await navigateTo('/admin/login')
  }
}

watch(() => route.fullPath, () => {
  mobileOpen.value = false
})

onMounted(async () => {
  try {
    const admin = await getItem<AdminUser>('/me')
    currentAdmin.value = admin.data
  } catch {
    currentAdmin.value = null
  }
})
</script>

<template>
  <div class="min-h-screen bg-elevated/40 lg:pl-(--app-sidebar)">
    <aside class="fixed inset-y-0 left-0 z-30 hidden w-(--app-sidebar) border-r border-default bg-default lg:flex lg:flex-col">
      <div class="flex h-16 items-center gap-2 border-b border-default px-5">
        <NuxtLink
          to="/admin"
          class="flex items-center gap-2"
          aria-label="Tiny School admin"
        >
          <AppLogo />
        </NuxtLink>
        <UBadge
          label="Admin"
          color="warning"
          variant="subtle"
          size="sm"
        />
      </div>

      <nav
        class="flex-1 space-y-1 overflow-y-auto p-3"
        aria-label="Admin navigation"
      >
        <p class="px-3 pb-1 pt-2 text-xs font-semibold uppercase tracking-wider text-dimmed">
          Back office
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
        <div class="flex items-center gap-3 px-2 pb-3">
          <UAvatar
            icon="i-lucide-shield-check"
            size="sm"
          />
          <span class="min-w-0">
            <span class="block truncate text-sm font-semibold">{{ currentAdmin?.name || 'Administrator' }}</span>
            <span class="block truncate text-xs text-muted">{{ currentAdmin?.email || 'Loading account…' }}</span>
          </span>
        </div>
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
    </aside>

    <header class="sticky top-0 z-20 flex h-16 items-center gap-3 border-b border-default bg-default/90 px-4 backdrop-blur lg:hidden">
      <UButton
        icon="i-lucide-menu"
        color="neutral"
        variant="ghost"
        aria-label="Open navigation"
        @click="mobileOpen = true"
      />
      <span class="min-w-0 flex-1 truncate text-sm font-semibold">Tiny School admin</span>
      <UColorModeButton />
    </header>

    <main>
      <slot />
    </main>

    <USlideover
      v-model:open="mobileOpen"
      side="left"
      title="Admin navigation"
    >
      <template #body>
        <nav class="space-y-1">
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
      </template>
      <template #footer>
        <UButton
          icon="i-lucide-log-out"
          label="Log out"
          color="error"
          variant="ghost"
          class="w-full justify-start"
          :loading="logoutPending"
          @click="logout"
        />
      </template>
    </USlideover>
  </div>
</template>
