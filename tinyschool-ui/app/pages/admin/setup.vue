<script setup lang="ts">
import type { AdminUser } from '~/types/api'

const toast = useToast()
const { postItem } = useAdminApi()
const pending = ref(false)
const form = reactive({ name: '', email: '', password: '', confirm: '' })

async function submit() {
  if (!form.email || !form.password) {
    toast.add({ title: 'Check your details', description: 'Email and password are required.', color: 'error' })
    return
  }
  if (form.password.length < 8) {
    toast.add({ title: 'Password too short', description: 'Use at least 8 characters.', color: 'error' })
    return
  }
  if (form.password !== form.confirm) {
    toast.add({ title: 'Passwords do not match', description: 'Retype the confirmation.', color: 'error' })
    return
  }
  pending.value = true
  try {
    await postItem<AdminUser>('/setup', {
      name: form.name.trim(),
      email: form.email.trim(),
      password: form.password
    })
    adminMarkerCookie().value = 'true'
    useState<boolean | null>('admin-exists').value = true
    toast.add({ title: 'Administrator created', description: 'You are signed in to the back office.', color: 'success' })
    await navigateTo('/admin')
  } catch (error) {
    toast.add({
      title: 'Could not create the administrator',
      description: apiErrorMessage(error, 'Please try again.'),
      color: 'error'
    })
  } finally {
    pending.value = false
  }
}

useSeoMeta({ title: 'Create administrator' })
</script>

<template>
  <div class="flex min-h-screen flex-col p-6 sm:p-10">
    <NuxtLink
      to="/"
      class="w-fit"
    ><AppLogo /></NuxtLink>
    <div class="mx-auto my-auto w-full max-w-md py-12">
      <UBadge
        label="First run"
        color="warning"
        variant="subtle"
        size="sm"
      />
      <h1 class="mt-3 text-3xl font-bold tracking-tight">
        Create the administrator
      </h1>
      <p class="mt-2 text-muted">
        No administrator exists yet. The account you create here owns the Tiny
        School back office.
      </p>
      <form
        class="mt-8 space-y-5"
        @submit.prevent="submit"
      >
        <UFormField label="Name">
          <UInput
            v-model="form.name"
            autocomplete="name"
            placeholder="Administrator"
            icon="i-lucide-user-round"
            class="w-full"
          />
        </UFormField>
        <UFormField
          label="Email"
          required
        >
          <UInput
            v-model="form.email"
            type="email"
            autocomplete="email"
            icon="i-lucide-mail"
            class="w-full"
          />
        </UFormField>
        <UFormField
          label="Password"
          required
        >
          <UInput
            v-model="form.password"
            type="password"
            autocomplete="new-password"
            icon="i-lucide-lock-keyhole"
            class="w-full"
          />
          <template #help>
            At least 8 characters.
          </template>
        </UFormField>
        <UFormField
          label="Confirm password"
          required
        >
          <UInput
            v-model="form.confirm"
            type="password"
            autocomplete="new-password"
            icon="i-lucide-lock-keyhole"
            class="w-full"
          />
        </UFormField>
        <UButton
          type="submit"
          block
          size="lg"
          label="Create administrator"
          trailing-icon="i-lucide-arrow-right"
          :loading="pending"
        />
      </form>
    </div>
  </div>
</template>
