<script setup lang="ts">
definePageMeta({ layout: 'dashboard' })
</script>

<template>
  <DomainDetailPage
    title="Assignment"
    endpoint="/api/v1/assignments"
    back-to="/dashboard/assignments"
    back-label="Assignments"
    icon="i-lucide-clipboard-check"
    :fields="[
      { key: 'type', label: 'Type' },
      { key: 'dueDate', label: 'Due date' },
      { key: 'totalScore', label: 'Total score' },
      { key: 'completion', label: 'Completion' },
      { key: 'class.name', label: 'Class' }
    ]"
  >
    <template #default="{ item }">
      <section>
        <div class="mb-5 flex items-start gap-3">
          <span class="grid size-9 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary">
            <UIcon
              name="i-lucide-users-round"
              class="size-4"
            />
          </span>
          <div>
            <h2 class="text-lg font-semibold text-highlighted">
              Assignees
            </h2>
            <p class="mt-1 text-sm leading-6 text-muted">
              Submit or reset scores. Each change creates a timestamped log.
            </p>
          </div>
        </div>
        <DomainScoreList
          kind="assignment"
          :items="item.assignees"
          :total-score="Number(item.totalScore || 100)"
        />
      </section>
    </template>
  </DomainDetailPage>
</template>
