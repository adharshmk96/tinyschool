<script setup lang="ts">
import type { ApiItem, BackupSettings, DatabaseBackup } from '~/types/api'

definePageMeta({ layout: 'admin' })

const toast = useToast()
const { request, download } = useAdminApi()
const frequencyOptions = [
  { label: 'Every day', value: 'daily' },
  { label: 'Every 2 days', value: 'every_2_days' },
  { label: 'Every week', value: 'weekly' }
]

const form = reactive<BackupSettings>({
  enabled: false,
  frequency: 'daily',
  runAt: '02:00',
  maxBackups: 14
})
const saving = ref(false)
const backingUp = ref(false)
const restoring = ref<string | null>(null)
const downloading = ref<string | null>(null)
const restoreOpen = ref(false)
const restoreTarget = ref<DatabaseBackup | null>(null)

const { data: settingsData, status: settingsStatus, error: settingsError, refresh: refreshSettings } = await useAsyncData(
  'admin-backup-settings',
  () => request<ApiItem<BackupSettings>>('/backups/settings')
)
const { data: backupData, status: backupsStatus, error: backupsError, refresh: refreshBackups } = await useAsyncData(
  'admin-backups',
  () => request<{ data: DatabaseBackup[] }>('/backups')
)

watch(settingsData, (response) => {
  if (!response?.data) return
  Object.assign(form, response.data)
}, { immediate: true })

const backups = computed(() => backupData.value?.data ?? [])

function formatSize(bytes: number) {
  if (bytes < 1024) return `${bytes} B`
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`
}

function formatDate(value: string) {
  return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium', timeStyle: 'short' }).format(new Date(value))
}

async function save() {
  saving.value = true
  try {
    const response = await request<ApiItem<BackupSettings>>('/backups/settings', {
      method: 'PUT',
      body: {
        enabled: form.enabled,
        frequency: form.frequency,
        runAt: form.runAt,
        maxBackups: Number(form.maxBackups)
      }
    })
    Object.assign(form, response.data)
    await refreshSettings()
    toast.add({ title: 'Backup settings saved', color: 'success' })
  } catch (error) {
    toast.add({ title: 'Could not save backup settings', description: apiErrorMessage(error, 'Please check the fields and try again.'), color: 'error' })
  } finally {
    saving.value = false
  }
}

async function backupNow() {
  backingUp.value = true
  try {
    await request('/backups', { method: 'POST' })
    await refreshBackups()
    toast.add({ title: 'Database backup created', color: 'success' })
  } catch (error) {
    toast.add({ title: 'Could not create backup', description: apiErrorMessage(error, 'Please try again.'), color: 'error' })
  } finally {
    backingUp.value = false
  }
}

async function downloadBackup(backup: DatabaseBackup) {
  downloading.value = backup.name
  try {
    const blob = await download(`/backups/${encodeURIComponent(backup.name)}/download`)
    const url = URL.createObjectURL(blob)
    const anchor = document.createElement('a')
    anchor.href = url
    anchor.download = backup.name
    anchor.click()
    URL.revokeObjectURL(url)
  } catch (error) {
    toast.add({ title: 'Could not download backup', description: apiErrorMessage(error, 'Please try again.'), color: 'error' })
  } finally {
    downloading.value = null
  }
}

function confirmRestore(backup: DatabaseBackup) {
  restoreTarget.value = backup
  restoreOpen.value = true
}

async function restoreBackup() {
  const backup = restoreTarget.value
  if (!backup) return
  restoring.value = backup.name
  try {
    await request(`/backups/${encodeURIComponent(backup.name)}/restore`, { method: 'POST' })
    await Promise.all([refreshBackups(), refreshSettings()])
    toast.add({ title: 'Database restored', description: 'A pre-restore safety backup was also created.', color: 'success' })
  } catch (error) {
    toast.add({ title: 'Could not restore backup', description: apiErrorMessage(error, 'The current database was not changed.'), color: 'error' })
  } finally {
    restoring.value = null
    restoreTarget.value = null
  }
}

useSeoMeta({ title: 'Admin backups' })
</script>

<template>
  <div class="page-shell">
    <PageHeading
      eyebrow="Back office"
      title="Backups"
      description="Schedule snapshots of the SQLite database, download them, or restore an earlier state."
    />

    <UAlert
      v-if="settingsError || backupsError"
      class="mt-8"
      color="error"
      variant="subtle"
      icon="i-lucide-circle-alert"
      title="Could not load backups"
      description="Check that the API is running, then reload this page."
    />

    <UCard class="mt-8">
      <template #header>
        <div class="flex items-center justify-between gap-4">
          <div>
            <h2 class="font-semibold text-highlighted">
              Scheduled backups
            </h2>
            <p class="mt-1 text-sm text-muted">
              Times use the API server's local timezone.
            </p>
          </div>
          <USwitch
            v-model="form.enabled"
            aria-label="Enable scheduled backups"
          />
        </div>
      </template>

      <form
        class="grid gap-5 sm:grid-cols-3"
        @submit.prevent="save"
      >
        <UFormField label="Frequency">
          <USelect
            v-model="form.frequency"
            :items="frequencyOptions"
            value-key="value"
            class="w-full"
            :disabled="!form.enabled"
          />
        </UFormField>
        <UFormField label="Run at">
          <UInput
            v-model="form.runAt"
            type="time"
            class="w-full"
            :disabled="!form.enabled"
          />
        </UFormField>
        <UFormField
          label="Max backups"
          help="Oldest backups are removed after a successful backup."
        >
          <UInput
            v-model.number="form.maxBackups"
            type="number"
            min="1"
            max="100"
            class="w-full"
          />
        </UFormField>

        <div class="flex flex-wrap items-center gap-3 sm:col-span-3">
          <UButton
            type="button"
            label="Back up now"
            color="neutral"
            variant="outline"
            icon="i-lucide-database-backup"
            :loading="backingUp"
            :disabled="saving || settingsStatus === 'pending'"
            @click="backupNow"
          />
          <UButton
            type="submit"
            label="Save"
            icon="i-lucide-save"
            :loading="saving"
            :disabled="backingUp || settingsStatus === 'pending'"
          />
          <p
            v-if="form.enabled && form.nextRunAt"
            class="text-sm text-muted"
          >
            Next run {{ formatDate(form.nextRunAt) }}
          </p>
        </div>
      </form>
    </UCard>

    <div class="mt-10 flex items-center justify-between gap-4">
      <h2 class="text-lg font-semibold text-highlighted">
        Database backups
      </h2>
      <span class="text-sm text-muted">{{ backups.length }} saved</span>
    </div>

    <div
      v-if="backupsStatus === 'pending'"
      class="mt-4 space-y-3"
    >
      <USkeleton
        v-for="index in 3"
        :key="index"
        class="h-20 w-full"
      />
    </div>

    <div
      v-else-if="backups.length"
      class="mt-4 space-y-3"
    >
      <UCard
        v-for="backup in backups"
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
              {{ formatDate(backup.createdAt) }} · {{ formatSize(backup.size) }}
            </p>
          </div>
          <div class="flex gap-2">
            <UButton
              icon="i-lucide-download"
              label="Download"
              color="neutral"
              variant="ghost"
              :loading="downloading === backup.name"
              :disabled="Boolean(restoring)"
              @click="downloadBackup(backup)"
            />
            <UButton
              icon="i-lucide-history"
              label="Restore"
              color="error"
              variant="soft"
              :loading="restoring === backup.name"
              :disabled="Boolean(restoring)"
              @click="confirmRestore(backup)"
            />
          </div>
        </div>
      </UCard>
    </div>

    <EmptyState
      v-else
      class="mt-4"
      icon="i-lucide-database-backup"
      title="No backups yet"
      description="Create one now or enable scheduled backups."
    />

    <ConfirmDialog
      v-model="restoreOpen"
      title="Restore this database backup?"
      :description="`Current data will be replaced with ${restoreTarget?.name}. A safety backup of the current database will be created first.`"
      confirm-label="Restore database"
      @confirm="restoreBackup"
    />
  </div>
</template>
