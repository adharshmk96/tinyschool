<script setup lang="ts">
definePageMeta({ layout: 'dashboard' })

const route = useRoute()
const router = useRouter()
const validTabs = ['performance', 'students']
const activeTab = computed(() => validTabs.includes(String(route.params.tab)) ? String(route.params.tab) : 'performance')
onMounted(() => {
  if (!validTabs.includes(String(route.params.tab)))
    router.replace(`/dashboard/exams/${route.params.id}/performance`)
})
</script>

<template>
  <DomainDetailPage
    title="Exam"
    endpoint="/api/v1/exams"
    back-to="/dashboard/exams"
    back-label="Exams"
    icon="i-lucide-file-check-2"
    :active-tab="activeTab"
    :fields="[
      { key: 'class.name', label: 'Class' },
      { key: 'examDate', label: 'Exam date' },
      { key: 'totalScore', label: 'Total score' },
      { key: 'markedCount', label: 'Marked' }
    ]"
    :tabs="[
      { label: 'Performance', value: 'performance', icon: 'i-lucide-chart-no-axes-combined' },
      { label: 'Students', value: 'students', icon: 'i-lucide-users' }
    ]"
  >
    <template #default="{ item }">
      <DomainPerformancePanel
        v-if="activeTab === 'performance'"
        subject="Scores"
        :data="item.performance"
      />
      <DomainScoreList v-else kind="exam" :items="item.students" :total-score="Number(item.totalScore || 100)" />
    </template>
  </DomainDetailPage>
</template>
