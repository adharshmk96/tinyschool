<script setup lang="ts">
import type { AdminUser } from '~/types/api'

const route = useRoute()
const toast = useToast()
const { postItem } = useAdminApi()
const pending = ref(false)
const form = reactive({ email: '', password: '' })

async function submit() {
  if (!form.email || !form.password) {
    toast.add({ title: 'Check your details', description: 'Email and password are required.', color: 'error' })
    return
  }
  pending.value = true
  try {
    await postItem<AdminUser>('/login', { email: form.email.trim(), password: form.password })
    toast.add({ title: 'Welcome back', description: 'You are signed in to the back office.', color: 'success' })
    await navigateTo(typeof route.query.redirect === 'string' ? route.query.redirect : '/admin')
  } catch (error) {
    toast.add({
      title: 'Could not log in',
      description: apiErrorMessage(error, 'Check your email and password, then try again.'),
      color: 'error'
    })
  } finally {
    pending.value = false
  }
}

useSeoMeta({ title: 'Admin log in' })
</script>

<template>
  <div class="flex min-h-screen flex-col p-6 sm:p-10">
    <NuxtLink
      to="/"
      class="w-fit"
    ><AppLogo /></NuxtLink>
    <div class="mx-auto my-auto w-full max-w-md py-12">
      <UBadge
        label="Back office"
        color="warning"
        variant="subtle"
        size="sm"
      />
      <h1 class="mt-3 text-3xl font-bold tracking-tight">
        Admin sign in
      </h1>
      <p class="mt-2 text-muted">
        This console is for Tiny School administrators only.
      </p>
      <form
        class="mt-8 space-y-5"
        @submit.prevent="submit"
      >
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
            autocomplete="current-password"
            icon="i-lucide-lock-keyhole"
            class="w-full"
          />
        </UFormField>
        <UButton
          type="submit"
          block
          size="lg"
          label="Log in"
          trailing-icon="i-lucide-arrow-right"
          :loading="pending"
        />
      </form>
      <p class="mt-6 text-center text-sm text-muted">
        Looking for your school workspace?
        <NuxtLink
          to="/login"
          class="font-semibold text-primary hover:underline"
        >Go to Tiny School</NuxtLink>
      </p>
    </div>
  </div>
</template>
