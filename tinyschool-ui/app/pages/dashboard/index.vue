<script setup lang="ts">
import type { Overview } from '~/types/api'

definePageMeta({ layout: 'dashboard' })
useSeoMeta({ title: 'Overview' })

const { getItem } = useApi()
const { selectedYear } = useSchoolContext()
const fallback: Overview = { students: 0, classes: 0, assignments: 0, exams: 0, upcoming: [] }
const { data, status, error, refresh } = await useAsyncData('overview', async () => {
  const response = await getItem<Overview>('/overview')
  return response.data
}, { default: () => fallback })

const cards = computed(() => [
  { label: 'Students', value: data.value?.students || 0, icon: 'i-lucide-users', to: '/dashboard/students' },
  { label: 'Classes', value: data.value?.classes || 0, icon: 'i-lucide-panels-top-left', to: '/dashboard/classes' },
  { label: 'Assignments', value: data.value?.assignments || 0, icon: 'i-lucide-clipboard-check', to: '/dashboard/assignments' },
  { label: 'Exams', value: data.value?.exams || 0, icon: 'i-lucide-file-chart-column', to: '/dashboard/exams' }
])

const upcoming = computed(() => (data.value?.upcoming || []).map((item) => {
  const date = new Date(`${item.date}T00:00:00`)
  const audience = item.className || 'Individual students'
  const students = `${item.studentCount} ${item.studentCount === 1 ? 'student' : 'students'}`
  return {
    ...item,
    day: new Intl.DateTimeFormat('en', { day: '2-digit' }).format(date),
    month: new Intl.DateTimeFormat('en', { month: 'short' }).format(date).toUpperCase(),
    meta: `${audience} · ${students}`,
    to: `/dashboard/${item.kind === 'exam' ? 'exams' : 'assignments'}/${item.id}`
  }
}))
</script>

<template>
  <div class="page-shell">
    <AppBreadcrumb
      :items="[
        { label: 'Overview', icon: 'i-lucide-layout-dashboard' }
      ]"
    />
    <PageHeading
      title="Overview"
      :description="`A quick look at ${selectedYear?.name || 'your current academic year'}.`"
      eyebrow="Dashboard"
    >
      <template #actions>
        <UButton
          icon="i-lucide-refresh-cw"
          color="neutral"
          variant="outline"
          label="Refresh"
          :loading="status === 'pending'"
          @click="refresh()"
        />
      </template>
    </PageHeading>

    <UAlert
      v-if="error"
      class="mt-6"
      color="error"
      variant="subtle"
      icon="i-lucide-wifi-off"
      title="The overview is unavailable"
      description="Start the local API and try again."
    />

    <div class="mt-9 grid gap-5 sm:grid-cols-2 xl:grid-cols-4">
      <template v-if="status === 'pending'">
        <USkeleton
          v-for="index in 4"
          :key="index"
          class="h-48 rounded-xl"
        />
      </template>
      <template v-else>
        <StatCard
          v-for="card in cards"
          :key="card.label"
          v-bind="card"
          detail="In this academic year"
        />
      </template>
    </div>

    <div class="mt-9 grid gap-6 xl:grid-cols-[1.4fr_1fr]">
      <UCard>
        <template #header>
          <div class="flex items-center justify-between">
            <div>
              <h2 class="font-semibold">
                Coming up
              </h2><p class="text-sm text-muted">
                Important dates and deadlines
              </p>
            </div>
            <UButton
              label="View assignments"
              to="/dashboard/assignments"
              color="neutral"
              variant="ghost"
              size="sm"
            />
          </div>
        </template>
        <div class="divide-y divide-default">
          <NuxtLink
            v-for="item in upcoming"
            :key="`${item.kind}-${item.id}`"
            :to="item.to"
            class="flex items-center gap-4 py-4 first:pt-0 last:pb-0"
          >
            <div class="grid size-12 shrink-0 place-items-center rounded-lg bg-elevated text-center">
              <span class="text-sm font-bold leading-none">{{ item.day }}</span>
              <span class="text-[10px] font-semibold text-muted">{{ item.month }}</span>
            </div>
            <div class="min-w-0">
              <p class="truncate text-sm font-semibold">
                {{ item.title }}
              </p><p class="truncate text-xs text-muted">
                {{ item.meta }}
              </p>
            </div>
          </NuxtLink>
          <EmptyState
            v-if="!upcoming.length"
            icon="i-lucide-calendar-check"
            title="Nothing coming up"
            description="Future assignment deadlines and exam dates will appear here."
          />
        </div>
      </UCard>
      <UCard>
        <template #header>
          <h2 class="font-semibold">
            Quick actions
          </h2><p class="text-sm text-muted">
            Common tasks
          </p>
        </template>
        <div class="grid grid-cols-2 gap-3">
          <UButton
            to="/dashboard/students"
            icon="i-lucide-user-plus"
            label="Add student"
            color="neutral"
            variant="soft"
            class="h-20 flex-col justify-center"
          />
          <UButton
            to="/dashboard/classes"
            icon="i-lucide-panel-top-open"
            label="New class"
            color="neutral"
            variant="soft"
            class="h-20 flex-col justify-center"
          />
          <UButton
            to="/dashboard/assignments"
            icon="i-lucide-clipboard-plus"
            label="Assignment"
            color="neutral"
            variant="soft"
            class="h-20 flex-col justify-center"
          />
          <UButton
            to="/dashboard/exams"
            icon="i-lucide-file-plus-2"
            label="Schedule exam"
            color="neutral"
            variant="soft"
            class="h-20 flex-col justify-center"
          />
        </div>
      </UCard>
    </div>
  </div>
</template>
