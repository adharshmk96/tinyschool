<script setup lang="ts">
const props = withDefaults(defineProps<{
  title: string
  description: string
  confirmLabel?: string
}>(), { confirmLabel: 'Delete' })

const emit = defineEmits<{ confirm: [] }>()
const open = defineModel<boolean>({ default: false })

function confirm() {
  emit('confirm')
  open.value = false
}
</script>

<template>
  <UModal
    v-model:open="open"
    :title="props.title"
    :description="props.description"
  >
    <template #body>
      <UAlert
        color="error"
        variant="subtle"
        icon="i-lucide-triangle-alert"
        title="This action cannot be undone."
      />
    </template>
    <template #footer>
      <div class="flex w-full justify-end gap-2">
        <UButton
          color="neutral"
          variant="ghost"
          label="Cancel"
          @click="open = false"
        />
        <UButton
          color="error"
          :label="confirmLabel"
          @click="confirm"
        />
      </div>
    </template>
  </UModal>
</template>
