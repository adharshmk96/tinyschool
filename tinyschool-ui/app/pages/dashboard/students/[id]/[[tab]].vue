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
      { key: 'classroom', label: 'Classroom (current year)' },
      { key: 'guardian.name', label: 'Guardian' },
      { key: 'guardian.email', label: 'Guardian email' },
      { key: 'guardian.phone', label: 'Guardian phone' },
      { key: 'residentAddress', label: 'Resident address' }
    ]"
    :edit-fields="[
      { key: 'firstName', label: 'First name' },
      { key: 'lastName', label: 'Last name' },
      { key: 'email', label: 'Email', type: 'email' },
      { key: 'phone', label: 'Phone' },
      { key: 'classrooms', label: 'Classroom by academic year', type: 'classrooms' },
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
    <template #default="{ item, refresh }">
      <DomainPerformancePanel
        v-if="activeTab === 'performance'"
        subject="Assignments"
        :data="item.performance"
      />
      <DomainActivityLogPanel
        v-else-if="activeTab === 'behaviour'"
        kind="behaviour"
        :items="item.behaviour"
        @changed="refresh"
      />
      <DomainScoreList
        v-else-if="activeTab === 'assignments'"
        kind="student"
        :items="item.assignments"
        @changed="refresh"
      />
      <DomainScoreList
        v-else-if="activeTab === 'exams'"
        kind="student"
        :items="item.exams"
        @changed="refresh"
      />
      <DomainActivityLogPanel
        v-else
        kind="notes"
        :items="item.notes"
        @changed="refresh"
      />
    </template>
  </DomainDetailPage>
</template>
