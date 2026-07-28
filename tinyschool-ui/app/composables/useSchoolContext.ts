import type { AcademicYear, School } from '~/types/api'

export function useSchoolContext() {
  const schools = useState<School[]>('schools', () => [])
  const academicYears = useState<AcademicYear[]>('academic-years', () => [])
  const selectedSchoolId = useCookie<string | undefined>('tiny-school-school')
  const selectedYearId = useCookie<string | undefined>('tiny-school-year')
  const loading = useState('school-context-loading', () => false)
  const loaded = useState('school-context-loaded', () => false)
  const { getCollection } = useApi()

  const selectedSchool = computed(() =>
    schools.value.find(item => item.id === selectedSchoolId.value) || schools.value[0]
  )
  const selectedYear = computed(() =>
    academicYears.value.find(item => item.id === selectedYearId.value)
    || academicYears.value.find(item => item.isCurrent)
    || academicYears.value[0]
  )

  async function load() {
    if (loading.value || loaded.value) return
    loading.value = true
    try {
      const [schoolResponse, yearResponse] = await Promise.all([
        getCollection<School>('/schools'),
        getCollection<AcademicYear>('/academic-years')
      ])
      schools.value = schoolResponse.data || []
      academicYears.value = yearResponse.data || []
      selectedSchoolId.value ||= schools.value[0]?.id
      selectedYearId.value ||= academicYears.value.find(item => item.isCurrent)?.id || academicYears.value[0]?.id
      loaded.value = true
    } finally {
      loading.value = false
    }
  }

  return {
    schools,
    academicYears,
    selectedSchoolId,
    selectedYearId,
    selectedSchool,
    selectedYear,
    loading,
    load
  }
}
