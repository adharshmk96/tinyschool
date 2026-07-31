<script setup lang="ts">
const toast = useToast()
const { request } = useApi()
const pending = ref(false)
const sent = ref(false)
const form = reactive({ email: '' })

async function submit() {
  if (!form.email) {
    toast.add({ title: 'Email is required', color: 'error' })
    return
  }
  pending.value = true
  try {
    await request('/auth/forgot-password', { method: 'POST', body: { email: form.email.trim() } })
    sent.value = true
  } catch (error) {
    toast.add({
      title: 'Could not send the reset link',
      description: apiErrorMessage(error, 'Please try again in a moment.'),
      color: 'error'
    })
  } finally {
    pending.value = false
  }
}

useSeoMeta({ title: 'Forgot password' })
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
            Reset your password
          </h1>
          <p class="mt-1 text-sm text-muted">
            Enter your email and we will send you a link to choose a new password.
          </p>
        </template>

        <div
          v-if="sent"
          class="space-y-4"
        >
          <UAlert
            color="success"
            variant="subtle"
            icon="i-lucide-mail-check"
            title="Check your inbox"
            description="If that email belongs to an account, a reset link is on its way. The link expires in one hour."
          />
          <UAlert
            color="warning"
            variant="subtle"
            icon="i-lucide-terminal"
            title="Email is not configured yet"
            description="Until SMTP is set up, the reset link is written to the API server log. Ask your administrator for it."
          />
          <UButton
            to="/login"
            block
            color="neutral"
            variant="outline"
            label="Back to log in"
          />
        </div>

        <form
          v-else
          class="space-y-5"
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
          <UButton
            type="submit"
            block
            size="lg"
            label="Send reset link"
            trailing-icon="i-lucide-arrow-right"
            :loading="pending"
          />
          <p class="text-center text-sm text-muted">
            Remembered it?
            <NuxtLink
              to="/login"
              class="font-semibold text-primary hover:underline"
            >Log in</NuxtLink>
          </p>
        </form>
      </UCard>
    </main>
  </div>
</template>
