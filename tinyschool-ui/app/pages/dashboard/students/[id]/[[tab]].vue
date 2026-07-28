<script setup lang="ts">
definePageMeta({ layout: 'dashboard' })

const route = useRoute()
const router = useRouter()
const validTabs = ['performance', 'behaviour', 'assignments', 'exams', 'notes']
const activeTab = computed(() => validTabs.includes(String(route.params.tab)) ? String(route.params.tab) : 'performance')

onMounted(() => {
  if (!validTabs.includes(String(route.params.tab)))
    router.replace(`/dashboard/students/${route.params.id}/performance`)
})
</script>

<template>
  <DomainDetailPage
    title="Student"
    endpoint="/api/v1/students"
    back-to="/dashboard/students"
    back-label="Students"
    icon="i-lucide-graduation-cap"
    :active-tab="activeTab"
    :fields="[
      { key: 'email', label: 'Email' },
      { key: 'phone', label: 'Phone' },
      { key: 'grade', label: 'Grade' },
      { key: 'guardian.name', label: 'Guardian' },
      { key: 'guardian.phone', label: 'Guardian phone' },
      { key: 'residentAddress', label: 'Resident address' }
    ]"
    :tabs="[
      { label: 'Performance', value: 'performance', icon: 'i-lucide-chart-no-axes-combined' },
      { label: 'Behaviour', value: 'behaviour', icon: 'i-lucide-heart-handshake' },
      { label: 'Assignments', value: 'assignments', icon: 'i-lucide-clipboard-check' },
      { label: 'Exams', value: 'exams', icon: 'i-lucide-file-check-2' },
      { label: 'Notes', value: 'notes', icon: 'i-lucide-notebook-pen' }
    ]"
  >
    <template #default="{ item }">
      <DomainPerformancePanel
        v-if="activeTab === 'performance'"
        subject="Assignments"
        :data="item.performance"
      />
      <DomainActivityLogPanel v-else-if="activeTab === 'behaviour'" kind="behaviour" :items="item.behaviour" />
      <DomainScoreList v-else-if="activeTab === 'assignments'" kind="student" :items="item.assignments" />
      <DomainScoreList v-else-if="activeTab === 'exams'" kind="student" :items="item.exams" />
      <DomainActivityLogPanel v-else kind="notes" :items="item.notes" />
    </template>
  </DomainDetailPage>
</template>
