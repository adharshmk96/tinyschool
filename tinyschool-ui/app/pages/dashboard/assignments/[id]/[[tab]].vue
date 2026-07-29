<script setup lang="ts">
definePageMeta({ layout: 'dashboard' })

const route = useRoute()
const router = useRouter()
const validTabs = ['performance', 'students']
const activeTab = computed(() => validTabs.includes(String(route.params.tab)) ? String(route.params.tab) : 'performance')
onMounted(() => {
  if (!validTabs.includes(String(route.params.tab)))
    router.replace(`/dashboard/assignments/${route.params.id}/performance`)
})
</script>

<template>
  <DomainDetailPage
    title="Assignment"
    endpoint="/api/v1/assignments"
    back-to="/dashboard/assignments"
    back-label="Assignments"
    icon="i-lucide-clipboard-check"
    grade-filter
    :fields="[
      { key: 'type', label: 'Type' },
      { key: 'dueDate', label: 'Due date' },
      { key: 'totalScore', label: 'Total score' },
      { key: 'completion', label: 'Completion' },
      { key: 'class.name', label: 'Class' }
    ]"
    :edit-fields="[
      { key: 'name', label: 'Name' },
      { key: 'dueDate', label: 'Due date', type: 'date' },
      { key: 'totalScore', label: 'Total score', type: 'number' }
    ]"
    :active-tab="activeTab"
    :tabs="[
      { label: 'Performance', value: 'performance', icon: 'i-lucide-chart-no-axes-combined' },
      { label: 'Students', value: 'students', icon: 'i-lucide-users' }
    ]"
  >
    <template #default="{ item, refresh }">
      <DomainPerformancePanel
        v-if="activeTab === 'performance'"
        mode="assignment"
        subject="Submissions"
        :data="item.performance"
      />
      <DomainScoreList
        v-else
        kind="assignment"
        :items="item.assignees"
        :total-score="Number(item.totalScore || 100)"
        @changed="refresh"
      />
    </template>
  </DomainDetailPage>
</template>
