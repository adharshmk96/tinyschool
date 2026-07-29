<script setup lang="ts">
const config = useRuntimeConfig()

const version = computed(() => {
  const raw = String(config.public.appVersion || 'dev').trim()
  if (!raw) return 'dev'
  return raw.startsWith('v') || raw === 'dev' ? raw : `v${raw}`
})
</script>

<template>
  <p
    class="pointer-events-none fixed inset-x-0 bottom-1.5 z-50 select-none text-center text-[10px] leading-none text-dimmed/60"
    aria-label="Application version"
  >
    {{ version }}
  </p>
</template>
