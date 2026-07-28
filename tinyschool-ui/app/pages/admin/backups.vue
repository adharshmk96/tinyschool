<script setup lang="ts">
definePageMeta({ layout: 'admin' })

// Placeholder only: nothing here talks to the API yet.
const rule = reactive({
  frequency: 'Daily',
  time: '02:00',
  retention: '14',
  destination: 'Local volume'
})

const frequencyOptions = ['Hourly', 'Daily', 'Weekly']
const destinationOptions = ['Local volume', 'S3 bucket', 'SFTP']

const sampleBackups = [
  { name: 'tinyschool-2026-07-28-0200.db', size: '4.2 MB', createdAt: '28 Jul 2026, 02:00' },
  { name: 'tinyschool-2026-07-27-0200.db', size: '4.1 MB', createdAt: '27 Jul 2026, 02:00' },
  { name: 'tinyschool-2026-07-26-0200.db', size: '4.0 MB', createdAt: '26 Jul 2026, 02:00' }
]

useSeoMeta({ title: 'Admin backups' })
</script>

<template>
  <div class="page-shell">
    <PageHeading
      eyebrow="Back office"
      title="Backups"
      description="Schedule snapshots of the SQLite database and decide how long they are kept."
    />

    <UAlert
      class="mt-8"
      color="warning"
      variant="subtle"
      icon="i-lucide-construction"
      title="Not wired up yet"
      description="This screen is a placeholder. The rules below are not saved and the listed snapshots are sample data."
    />

    <UCard class="mt-6">
      <template #header>
        <h2 class="font-semibold text-highlighted">
          Backup rule
        </h2>
      </template>
      <form
        class="grid gap-5 sm:grid-cols-2"
        @submit.prevent
      >
        <UFormField label="Frequency">
          <USelect
            v-model="rule.frequency"
            :items="frequencyOptions"
            class="w-full"
            disabled
          />
        </UFormField>
        <UFormField label="Run at">
          <UInput
            v-model="rule.time"
            type="time"
            class="w-full"
            disabled
          />
        </UFormField>
        <UFormField label="Keep for (days)">
          <UInput
            v-model="rule.retention"
            type="number"
            class="w-full"
            disabled
          />
        </UFormField>
        <UFormField label="Destination">
          <USelect
            v-model="rule.destination"
            :items="destinationOptions"
            class="w-full"
            disabled
          />
        </UFormField>
        <div class="flex items-end gap-2 sm:col-span-2">
          <UButton
            label="Back up now"
            color="neutral"
            variant="outline"
            icon="i-lucide-download"
            disabled
          />
          <UButton
            label="Save rule"
            icon="i-lucide-save"
            disabled
          />
        </div>
      </form>
    </UCard>

    <h2 class="mt-10 text-lg font-semibold text-highlighted">
      Recent snapshots
    </h2>
    <div class="mt-4 space-y-3">
      <UCard
        v-for="backup in sampleBackups"
        :key="backup.name"
      >
        <div class="flex flex-wrap items-center gap-4">
          <span class="grid size-10 place-items-center rounded-lg bg-primary/10 text-primary">
            <UIcon
              name="i-lucide-database"
              class="size-5"
            />
          </span>
          <div class="min-w-0 flex-1">
            <p class="truncate font-medium text-highlighted">
              {{ backup.name }}
            </p>
            <p class="text-xs text-muted">
              {{ backup.createdAt }} · {{ backup.size }}
            </p>
          </div>
          <UButton
            icon="i-lucide-download"
            label="Download"
            color="neutral"
            variant="ghost"
            disabled
          />
        </div>
      </UCard>
    </div>
  </div>
</template>
