<script setup lang="ts">
defineProps<{ compact?: boolean }>()
</script>

<template>
  <span class="inline-flex items-center gap-2.5 font-bold tracking-tight">
    <span class="grid size-9 place-items-center rounded-xl bg-primary text-white shadow-sm shadow-primary/25">
      <UIcon
        name="i-lucide-school"
        class="size-5"
      />
    </span>
    <span
      v-if="!compact"
      class="text-lg"
    >Tiny School</span>
  </span>
</template>
