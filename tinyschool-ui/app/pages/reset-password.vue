<script setup lang="ts">
const route = useRoute()
const toast = useToast()
const { request } = useApi()
const pending = ref(false)
const token = computed(() => (typeof route.query.token === 'string' ? route.query.token : ''))
const form = reactive({ password: '', confirmPassword: '' })

async function submit() {
  if (form.password.length < 8) {
    toast.add({ title: 'Password is too short', description: 'Use at least 8 characters.', color: 'error' })
    return
  }
  if (form.password !== form.confirmPassword) {
    toast.add({ title: 'Passwords do not match', color: 'error' })
    return
  }
  pending.value = true
  try {
    await request('/auth/reset-password', {
      method: 'POST',
      body: { token: token.value, newPassword: form.password }
    })
    // The reset signs every session out, so the stale marker has to go too.
    useCookie('tiny-school-authenticated').value = null
    toast.add({ title: 'Password updated', description: 'Log in with your new password.', color: 'success' })
    await navigateTo('/login')
  } catch (error) {
    toast.add({
      title: 'Could not reset your password',
      description: apiErrorMessage(error, 'The link may have expired. Request a new one.'),
      color: 'error'
    })
  } finally {
    pending.value = false
  }
}

useSeoMeta({ title: 'Choose a new password' })
</script>

<template>
  <div class="min-h-screen bg-elevated/40">
    <header class="p-6 sm:p-10">
      <NuxtLink to="/"><AppLogo /></NuxtLink>
    </header>
    <main class="mx-auto w-full max-w-md px-6 pb-16">
      <UCard>
        <template #header>
          <h1 class="text-2xl font-bold">
            Choose a new password
          </h1>
          <p class="mt-1 text-sm text-muted">
            Resetting your password signs you out everywhere else.
          </p>
        </template>

        <div
          v-if="!token"
          class="space-y-4"
        >
          <UAlert
            color="error"
            variant="subtle"
            icon="i-lucide-link-2-off"
            title="This link is incomplete"
            description="Open the reset link exactly as it was sent to you, or request a new one."
          />
          <UButton
            to="/forgot-password"
            block
            label="Request a new link"
          />
        </div>

        <form
          v-else
          class="space-y-5"
          @submit.prevent="submit"
        >
          <UFormField
            label="New password"
            required
          >
            <UInput
              v-model="form.password"
              type="password"
              autocomplete="new-password"
              icon="i-lucide-lock-keyhole"
              class="w-full"
            />
          </UFormField>
          <UFormField
            label="Confirm new password"
            required
          >
            <UInput
              v-model="form.confirmPassword"
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
            label="Update password"
            trailing-icon="i-lucide-arrow-right"
            :loading="pending"
          />
          <p class="text-center text-sm text-muted">
            <NuxtLink
              to="/login"
              class="font-semibold text-primary hover:underline"
            >Back to log in</NuxtLink>
          </p>
        </form>
      </UCard>
    </main>
  </div>
</template>
