<script setup lang="ts">
useHead({
  htmlAttrs: { lang: 'en' },
  link: [{ rel: 'icon', href: '/favicon.ico' }],
  meta: [{ name: 'viewport', content: 'width=device-width, initial-scale=1' }]
})

useSeoMeta({
  titleTemplate: title => title ? `${title} · Tiny School` : 'Tiny School',
  description: 'A calm, focused workspace for running a small school.'
})
</script>

<template>
  <UApp>
    <NuxtLayout>
      <NuxtPage />
    </NuxtLayout>
    <AppVersionFooter />
  </UApp>
</template>
