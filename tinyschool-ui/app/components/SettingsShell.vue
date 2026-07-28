<script setup lang="ts">
const route = useRoute()

const breadcrumbs = computed(() => {
  const items: Array<{ label: string, to?: string, icon?: string }> = [
    { label: 'Overview', to: '/dashboard', icon: 'i-lucide-layout-dashboard' },
    { label: 'Settings', to: '/dashboard/settings/account', icon: 'i-lucide-settings' }
  ]

  if (route.path.includes('/academic-years')) {
    items.push({ label: 'Academic years', to: '/dashboard/settings/academic-years', icon: 'i-lucide-calendar-range' })
    if (route.path.endsWith('/create'))
      items.push({ label: 'Create', icon: 'i-lucide-calendar-plus' })
    else if (route.path.endsWith('/edit'))
      items.push({ label: 'Edit', icon: 'i-lucide-pencil' })
    else if (route.path !== '/dashboard/settings/academic-years')
      items.push({ label: 'Details', icon: 'i-lucide-calendar-days' })
  } else if (route.path.endsWith('/schools')) {
    items.push({ label: 'My schools', icon: 'i-lucide-school' })
  } else if (route.path.endsWith('/account')) {
    items.push({ label: 'Account', icon: 'i-lucide-user-round' })
  } else if (route.path.endsWith('/data')) {
    items.push({ label: 'Data', icon: 'i-lucide-database' })
  }

  return items
})
</script>

<template>
  <div class="page-shell">
    <AppBreadcrumb :items="breadcrumbs" />
    <PageHeading
      title="Settings"
      description="Manage your account, schools, grades and academic calendar."
      eyebrow="Workspace"
    />
    <div class="mt-10 grid gap-8 lg:grid-cols-[16rem_minmax(0,1fr)] lg:gap-10">
      <SettingsNav />
      <div class="min-w-0">
        <slot />
      </div>
    </div>
  </div>
</template>
