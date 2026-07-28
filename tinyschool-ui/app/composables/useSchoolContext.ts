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
  const availableAcademicYears = computed(() =>
    academicYears.value.filter(item => item.schoolId === selectedSchool.value?.id)
  )
  const selectedYear = computed(() =>
    availableAcademicYears.value.find(item => item.id === selectedYearId.value)
    || availableAcademicYears.value.find(item => item.isCurrent)
    || availableAcademicYears.value[0]
  )

  watch(selectedSchoolId, () => {
    if (!availableAcademicYears.value.some(item => item.id === selectedYearId.value)) {
      selectedYearId.value = availableAcademicYears.value.find(item => item.isCurrent)?.id
        || availableAcademicYears.value[0]?.id
    }
  })

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
      if (!schools.value.some(item => item.id === selectedSchoolId.value))
        selectedSchoolId.value = schools.value[0]?.id
      const yearsForSchool = academicYears.value.filter(item => item.schoolId === selectedSchoolId.value)
      if (!yearsForSchool.some(item => item.id === selectedYearId.value))
        selectedYearId.value = yearsForSchool.find(item => item.isCurrent)?.id || yearsForSchool[0]?.id
      loaded.value = true
    } finally {
      loading.value = false
    }
  }

  return {
    schools,
    academicYears,
    availableAcademicYears,
    selectedSchoolId,
    selectedYearId,
    selectedSchool,
    selectedYear,
    loading,
    load
  }
}
