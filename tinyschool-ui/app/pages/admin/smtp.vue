<script setup lang="ts">
definePageMeta({ layout: 'admin' })

// Placeholder only: the fields are disabled until the API grows SMTP settings.
const form = reactive({
  host: '',
  port: '587',
  username: '',
  password: '',
  fromName: 'Tiny School',
  fromEmail: '',
  encryption: 'STARTTLS'
})

const encryptionOptions = ['None', 'STARTTLS', 'TLS']

useSeoMeta({ title: 'Admin SMTP' })
</script>

<template>
  <div class="page-shell">
    <PageHeading
      eyebrow="Back office"
      title="SMTP"
      description="Configure the mail server Tiny School uses for invitations, password resets and reports."
    />

    <UAlert
      class="mt-8"
      color="warning"
      variant="subtle"
      icon="i-lucide-construction"
      title="Not wired up yet"
      description="This screen is a placeholder. Saving is disabled until the API stores SMTP settings. Until then, password reset links are written to the API server log instead of being emailed."
    />

    <UCard class="mt-6">
      <form
        class="grid gap-5 sm:grid-cols-2"
        @submit.prevent
      >
        <UFormField label="Host">
          <UInput
            v-model="form.host"
            placeholder="smtp.example.com"
            icon="i-lucide-server"
            class="w-full"
            disabled
          />
        </UFormField>
        <UFormField label="Port">
          <UInput
            v-model="form.port"
            class="w-full"
            disabled
          />
        </UFormField>
        <UFormField label="Username">
          <UInput
            v-model="form.username"
            class="w-full"
            disabled
          />
        </UFormField>
        <UFormField label="Password">
          <UInput
            v-model="form.password"
            type="password"
            class="w-full"
            disabled
          />
        </UFormField>
        <UFormField label="From name">
          <UInput
            v-model="form.fromName"
            class="w-full"
            disabled
          />
        </UFormField>
        <UFormField label="From email">
          <UInput
            v-model="form.fromEmail"
            type="email"
            placeholder="no-reply@example.com"
            class="w-full"
            disabled
          />
        </UFormField>
        <UFormField label="Encryption">
          <USelect
            v-model="form.encryption"
            :items="encryptionOptions"
            class="w-full"
            disabled
          />
        </UFormField>
        <div class="flex items-end gap-2 sm:col-span-2">
          <UButton
            label="Send test email"
            color="neutral"
            variant="outline"
            icon="i-lucide-send"
            disabled
          />
          <UButton
            label="Save settings"
            icon="i-lucide-save"
            disabled
          />
        </div>
      </form>
    </UCard>
  </div>
</template>
