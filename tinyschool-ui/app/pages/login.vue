<script setup lang="ts">
const route = useRoute()
const toast = useToast()
const { postItem } = useApi()
const pending = ref(false)
const form = reactive({ email: '', password: '' })

async function submit() {
  if (!form.email || !form.password) {
    toast.add({ title: 'Check your details', description: 'Email and password are required.', color: 'error' })
    return
  }
  pending.value = true
  try {
    await postItem('/auth/login', {
      email: form.email.trim(),
      password: form.password
    })
    useCookie('tiny-school-authenticated', {
      sameSite: 'lax',
      maxAge: 60 * 60 * 24
    }).value = 'true'
    toast.add({ title: 'Welcome back', description: 'You are signed in.', color: 'success' })
    await navigateTo(typeof route.query.redirect === 'string' ? route.query.redirect : '/dashboard')
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

useSeoMeta({ title: 'Log in' })
</script>

<template>
  <div class="grid min-h-screen lg:grid-cols-2">
    <div class="flex flex-col p-6 sm:p-10">
      <NuxtLink
        to="/"
        class="w-fit"
      ><AppLogo /></NuxtLink>
      <div class="mx-auto my-auto w-full max-w-md py-12">
        <h1 class="text-3xl font-bold tracking-tight">
          Welcome back
        </h1>
        <p class="mt-2 text-muted">
          Log in to continue to your school workspace.
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
          New to Tiny School?
          <NuxtLink
            to="/register"
            class="font-semibold text-primary hover:underline"
          >Create an account</NuxtLink>
        </p>
      </div>
    </div>
    <div class="app-grid relative hidden overflow-hidden bg-primary/5 p-12 lg:grid lg:place-items-center">
      <blockquote class="relative max-w-lg text-3xl font-semibold leading-relaxed tracking-tight">
        “Everything we need is finally in one clear, quiet place.”
        <footer class="mt-6 text-base font-normal text-muted">
          Avery Morgan · Oakridge Learning Centre
        </footer>
      </blockquote>
    </div>
  </div>
</template>
