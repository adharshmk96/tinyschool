<script setup lang="ts">
useSeoMeta({ title: 'Simple school management' })
</script>

<template>
  <div class="relative min-h-screen overflow-hidden">
    <div class="app-grid pointer-events-none absolute inset-0 opacity-60 [mask-image:linear-gradient(to_bottom,black,transparent_80%)]" />
    <header class="relative z-10">
      <UContainer class="flex h-20 items-center justify-between">
        <AppLogo />
        <div class="flex items-center gap-2">
          <UColorModeButton />
          <UButton
            to="/login"
            color="neutral"
            variant="ghost"
            label="Log in"
          />
          <UButton
            to="/register"
            label="Get started"
            trailing-icon="i-lucide-arrow-right"
          />
        </div>
      </UContainer>
    </header>

    <main class="relative">
      <UContainer class="grid min-h-[calc(100vh-5rem)] items-center gap-14 py-16 lg:grid-cols-[1.1fr_.9fr]">
        <div class="max-w-3xl">
          <UBadge
            color="primary"
            variant="subtle"
            label="Made for small schools"
            icon="i-lucide-sparkles"
          />
          <h1 class="mt-6 text-5xl font-bold tracking-tight text-balance sm:text-6xl lg:text-7xl">
            Your whole school,
            <span class="text-primary">beautifully simple.</span>
          </h1>
          <p class="mt-6 max-w-2xl text-lg leading-8 text-muted">
            Keep students, classes, assignments and exams in one calm workspace. Less administration, more time for teaching.
          </p>
          <div class="mt-8 flex flex-wrap gap-3">
            <UButton
              to="/register"
              size="xl"
              label="Create your school"
              trailing-icon="i-lucide-arrow-right"
            />
            <UButton
              to="/login"
              size="xl"
              color="neutral"
              variant="outline"
              label="Log in"
            />
          </div>
          <div class="mt-10 flex flex-wrap gap-x-6 gap-y-3 text-sm text-muted">
            <span class="flex items-center gap-2"><UIcon
              name="i-lucide-check-circle-2"
              class="text-primary"
            />Quick to set up</span>
            <span class="flex items-center gap-2"><UIcon
              name="i-lucide-check-circle-2"
              class="text-primary"
            />Works on every device</span>
            <span class="flex items-center gap-2"><UIcon
              name="i-lucide-check-circle-2"
              class="text-primary"
            />Built for educators</span>
          </div>
        </div>

        <div class="relative mx-auto w-full max-w-lg">
          <div class="absolute -inset-10 rounded-full bg-primary/10 blur-3xl" />
          <UCard class="relative shadow-2xl shadow-primary/10">
            <div class="flex items-center justify-between border-b border-default pb-5">
              <div>
                <p class="font-semibold">
                  Good morning, Avery
                </p>
                <p class="text-sm text-muted">
                  Monday, 28 July
                </p>
              </div>
              <UAvatar text="AM" />
            </div>
            <div class="grid grid-cols-2 gap-3 py-5">
              <div
                v-for="item in [
                  ['Students', '128', 'i-lucide-users'],
                  ['Classes', '8', 'i-lucide-panels-top-left'],
                  ['Assignments', '14', 'i-lucide-clipboard-check'],
                  ['Exams', '4', 'i-lucide-file-chart-column']
                ]"
                :key="item[0]"
                class="rounded-xl bg-elevated p-4"
              >
                <UIcon
                  :name="item[2]"
                  class="size-5 text-primary"
                />
                <p class="mt-4 text-2xl font-bold">
                  {{ item[1] }}
                </p>
                <p class="text-xs text-muted">
                  {{ item[0] }}
                </p>
              </div>
            </div>
            <UAlert
              color="primary"
              variant="subtle"
              icon="i-lucide-calendar-days"
              title="Midterm exams begin in 12 days"
            />
          </UCard>
        </div>
      </UContainer>
    </main>
  </div>
</template>
