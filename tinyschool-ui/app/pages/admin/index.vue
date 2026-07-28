<script setup lang="ts">
import type { AdminUser } from '~/types/api'

definePageMeta({ layout: 'admin' })

const { request } = useAdminApi()

const { data, status } = await useAsyncData('admin-overview-users', () =>
  request<{ data: AdminUser[], meta?: { total?: number } }>('/users?pageSize=100&sort=name&order=asc'))

const users = computed(() => data.value?.data ?? [])
const total = computed(() => Number(data.value?.meta?.total ?? users.value.length))
const blocked = computed(() => users.value.filter(user => user.blocked).length)
const admins = computed(() => users.value.filter(user => user.role === 'admin').length)

const sections = [
  {
    label: 'Users',
    description: 'Browse every account, search, sort and block access.',
    icon: 'i-lucide-users',
    to: '/admin/users'
  },
  {
    label: 'SMTP',
    description: 'Point Tiny School at a mail server for transactional email.',
    icon: 'i-lucide-mail',
    to: '/admin/smtp'
  },
  {
    label: 'Backups',
    description: 'Schedule and retain snapshots of the SQLite database.',
    icon: 'i-lucide-database-backup',
    to: '/admin/backups'
  }
]

useSeoMeta({ title: 'Admin overview' })
</script>

<template>
  <div class="page-shell">
    <PageHeading
      eyebrow="Back office"
      title="Admin overview"
      description="Operate the Tiny School instance: accounts, mail delivery and database backups."
    />

    <div
      v-if="status === 'pending'"
      class="mt-10 grid gap-5 sm:grid-cols-3"
    >
      <USkeleton
        v-for="index in 3"
        :key="index"
        class="h-36 rounded-xl"
      />
    </div>
    <div
      v-else
      class="mt-10 grid gap-5 sm:grid-cols-3"
    >
      <StatCard
        label="Accounts"
        :value="total"
        icon="i-lucide-users"
        to="/admin/users"
        detail="Registered users and administrators"
      />
      <StatCard
        label="Blocked"
        :value="blocked"
        icon="i-lucide-user-x"
        to="/admin/users"
        detail="Accounts denied sign-in"
      />
      <StatCard
        label="Administrators"
        :value="admins"
        icon="i-lucide-shield-check"
        to="/admin/users"
        detail="Accounts with back-office access"
      />
    </div>

    <h2 class="mt-12 text-lg font-semibold text-highlighted">
      Manage
    </h2>
    <div class="mt-4 grid gap-5 lg:grid-cols-3">
      <NuxtLink
        v-for="section in sections"
        :key="section.to"
        :to="section.to"
        class="group rounded-xl border border-default bg-default p-5 shadow-xs transition hover:-translate-y-0.5 hover:border-primary/50 hover:shadow-md"
      >
        <span class="grid size-10 place-items-center rounded-lg bg-primary/10 text-primary">
          <UIcon
            :name="section.icon"
            class="size-5"
          />
        </span>
        <p class="mt-5 font-semibold text-highlighted">{{ section.label }}</p>
        <p class="mt-1 text-sm text-muted">{{ section.description }}</p>
      </NuxtLink>
    </div>
  </div>
</template>
