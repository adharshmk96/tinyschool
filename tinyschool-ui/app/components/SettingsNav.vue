<script setup lang="ts">
const route = useRoute()
const links = [
  { label: 'Account', description: 'Profile and password', icon: 'i-lucide-user-round', to: '/dashboard/settings/account' },
  { label: 'My schools', description: 'Schools and grades', icon: 'i-lucide-school', to: '/dashboard/settings/schools' },
  { label: 'Academic years', description: 'Terms and vacations', icon: 'i-lucide-calendar-range', to: '/dashboard/settings/academic-years' },
  { label: 'Data', description: 'Clear workspace data', icon: 'i-lucide-database', to: '/dashboard/settings/data' }
]
</script>

<template>
  <nav
    class="flex gap-2 overflow-x-auto rounded-xl border border-default bg-default p-2 shadow-xs lg:sticky lg:top-24 lg:flex-col lg:self-start"
    aria-label="Settings navigation"
  >
    <NuxtLink
      v-for="link in links"
      :key="link.to"
      :to="link.to"
      class="flex min-w-max items-center gap-3 rounded-lg px-3.5 py-3 transition lg:min-w-0"
      :class="route.path.startsWith(link.to) ? 'bg-primary/10 text-primary' : 'text-muted hover:bg-elevated hover:text-highlighted'"
    >
      <UIcon
        :name="link.icon"
        class="size-5 shrink-0"
      />
      <span>
        <span class="block text-sm font-semibold">{{ link.label }}</span>
        <span class="hidden text-xs opacity-75 lg:block">{{ link.description }}</span>
      </span>
    </NuxtLink>
  </nav>
</template>
