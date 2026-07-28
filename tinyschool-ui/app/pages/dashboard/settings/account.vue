<script setup lang="ts">
import type { User } from '~/types/api'

definePageMeta({ layout: 'dashboard' })
useSeoMeta({ title: 'Account settings' })

const toast = useToast()
const { getItem } = useApi()
const { data, status, error } = await useAsyncData('me', async () => (await getItem<User>('/me')).data, {
  default: () => ({ id: '', name: '', email: '' })
})
const profile = reactive({ name: '', email: '' })
const password = reactive({ current: '', next: '', confirm: '' })

watch(data, (user) => {
  profile.name = user?.name || ''
  profile.email = user?.email || ''
}, { immediate: true })

function saveProfile() {
  if (!profile.name.trim()) {
    toast.add({ title: 'Name is required', color: 'error' })
    return
  }
  toast.add({ title: 'Profile updated', description: 'Placeholder changes are saved for this session.', color: 'success' })
}

function changePassword() {
  if (!password.current || password.next.length < 8) {
    toast.add({ title: 'Check your passwords', description: 'Enter your current password and use at least 8 characters.', color: 'error' })
    return
  }
  if (password.next !== password.confirm) {
    toast.add({ title: 'New passwords do not match', color: 'error' })
    return
  }
  password.current = ''
  password.next = ''
  password.confirm = ''
  toast.add({ title: 'Password updated', color: 'success' })
}
</script>

<template>
  <SettingsShell>
    <div class="space-y-6">
      <UAlert
        v-if="error"
        color="error"
        variant="subtle"
        icon="i-lucide-wifi-off"
        title="Account details could not be loaded"
      />
      <UCard>
        <template #header>
          <h2 class="font-semibold">
            Profile
          </h2>
          <p class="text-sm text-muted">
            Update the name shown across your workspace.
          </p>
        </template>
        <form
          class="max-w-xl space-y-5"
          @submit.prevent="saveProfile"
        >
          <UFormField
            label="Name"
            required
          >
            <UInput
              v-model="profile.name"
              class="w-full"
              :loading="status === 'pending'"
              autocomplete="name"
            />
          </UFormField>
          <UFormField
            label="Email"
            hint="Email cannot be changed"
          >
            <UInput
              v-model="profile.email"
              class="w-full"
              disabled
              icon="i-lucide-lock-keyhole"
            />
          </UFormField>
          <UButton
            type="submit"
            icon="i-lucide-save"
            label="Save profile"
          />
        </form>
      </UCard>

      <UCard>
        <template #header>
          <h2 class="font-semibold">
            Change password
          </h2>
          <p class="text-sm text-muted">
            Choose a unique password with at least 8 characters.
          </p>
        </template>
        <form
          class="max-w-xl space-y-5"
          @submit.prevent="changePassword"
        >
          <UFormField
            label="Current password"
            required
          >
            <UInput
              v-model="password.current"
              type="password"
              autocomplete="current-password"
              class="w-full"
            />
          </UFormField>
          <div class="grid gap-5 sm:grid-cols-2">
            <UFormField
              label="New password"
              required
            >
              <UInput
                v-model="password.next"
                type="password"
                autocomplete="new-password"
                class="w-full"
              />
            </UFormField>
            <UFormField
              label="Confirm password"
              required
            >
              <UInput
                v-model="password.confirm"
                type="password"
                autocomplete="new-password"
                class="w-full"
              />
            </UFormField>
          </div>
          <UButton
            type="submit"
            color="neutral"
            icon="i-lucide-key-round"
            label="Update password"
          />
        </form>
      </UCard>
    </div>
  </SettingsShell>
</template>
