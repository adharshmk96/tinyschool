<script setup lang="ts">
const toast = useToast()
const { postItem } = useApi()
const pending = ref(false)
const form = reactive({
  name: '',
  email: '',
  school: '',
  password: '',
  confirmPassword: ''
})

async function submit() {
  if (!form.name || !form.email || !form.password) {
    toast.add({ title: 'Missing information', description: 'Name, email and password are required.', color: 'error' })
    return
  }
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
    await postItem('/auth/register', {
      name: form.name.trim(),
      email: form.email.trim(),
      password: form.password,
      schoolName: form.school.trim()
    })
    useCookie('tiny-school-authenticated', {
      sameSite: 'lax',
      maxAge: 60 * 60 * 24
    }).value = 'true'
    toast.add({ title: 'Your workspace is ready', color: 'success' })
    await navigateTo('/dashboard')
  } catch (error) {
    toast.add({
      title: 'Could not create account',
      description: apiErrorMessage(error, 'Check your details and try again.'),
      color: 'error'
    })
  } finally {
    pending.value = false
  }
}

useSeoMeta({ title: 'Create account' })
</script>

<template>
  <div class="min-h-screen bg-elevated/40">
    <header class="p-6 sm:p-10">
      <NuxtLink to="/"><AppLogo /></NuxtLink>
    </header>
    <main class="mx-auto w-full max-w-xl px-6 pb-16">
      <UCard>
        <template #header>
          <h1 class="text-2xl font-bold">
            Create your workspace
          </h1>
          <p class="mt-1 text-sm text-muted">
            You can invite others and add more schools later.
          </p>
        </template>
        <form
          class="space-y-5"
          @submit.prevent="submit"
        >
          <div class="grid gap-5 sm:grid-cols-2">
            <UFormField
              label="Your name"
              required
            >
              <UInput
                v-model="form.name"
                autocomplete="name"
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
                class="w-full"
              />
            </UFormField>
          </div>
          <UFormField
            label="School name"
            hint="Optional"
          >
            <UInput
              v-model="form.school"
              icon="i-lucide-school"
              placeholder="Oakridge Learning Centre"
              class="w-full"
            />
          </UFormField>
          <div class="grid gap-5 sm:grid-cols-2">
            <UFormField
              label="Password"
              required
            >
              <UInput
                v-model="form.password"
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
                v-model="form.confirmPassword"
                type="password"
                autocomplete="new-password"
                class="w-full"
              />
            </UFormField>
          </div>
          <UButton
            type="submit"
            block
            size="lg"
            label="Create account"
            trailing-icon="i-lucide-arrow-right"
            :loading="pending"
          />
        </form>
      </UCard>
      <p class="mt-6 text-center text-sm text-muted">
        Already have an account?
        <NuxtLink
          to="/login"
          class="font-semibold text-primary hover:underline"
        >Log in</NuxtLink>
      </p>
    </main>
  </div>
</template>
