<script setup lang="ts">
definePageMeta({ layout: 'dashboard' })

const route = useRoute()
const router = useRouter()
const validTabs = ['performance', 'assignments', 'exams', 'students']
const activeTab = computed(() => validTabs.includes(String(route.params.tab)) ? String(route.params.tab) : 'performance')
onMounted(() => {
  if (!validTabs.includes(String(route.params.tab)))
    router.replace(`/dashboard/classes/${route.params.id}/performance`)
})
</script>

<template>
  <DomainDetailPage
    title="Class"
    endpoint="/api/v1/classes"
    back-to="/dashboard/classes"
    back-label="My Classes"
    icon="i-lucide-presentation"
    :active-tab="activeTab"
    :fields="[
      { key: 'subject', label: 'Subject' },
      { key: 'grade', label: 'Grade' },
      { key: 'studentCount', label: 'Students' },
      { key: 'description', label: 'Description' }
    ]"
    :tabs="[
      { label: 'Performance', value: 'performance', icon: 'i-lucide-chart-no-axes-combined' },
      { label: 'Assignments', value: 'assignments', icon: 'i-lucide-clipboard-check' },
      { label: 'Exams', value: 'exams', icon: 'i-lucide-file-check-2' },
      { label: 'Students', value: 'students', icon: 'i-lucide-users' }
    ]"
  >
    <template #default="{ item }">
      <DomainPerformancePanel
        v-if="activeTab === 'performance'"
        mode="class"
        subject="Assignments"
        :data="item.performance"
      />
      <DomainRelatedCards
        v-else
        :kind="activeTab"
        :items="item[activeTab]"
      />
    </template>
  </DomainDetailPage>
</template>
