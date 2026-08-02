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
    || currentYearForToday(availableAcademicYears.value)
    || availableAcademicYears.value[0]
  )

  function today() {
    const now = new Date()
    const month = String(now.getMonth() + 1).padStart(2, '0')
    const day = String(now.getDate()).padStart(2, '0')
    return `${now.getFullYear()}-${month}-${day}`
  }

  function currentYearForToday(years: AcademicYear[]) {
    const currentDate = today()
    return years.find(item => item.endDate && item.startDate <= currentDate && currentDate <= item.endDate)
  }

  function selectCurrentYear() {
    const years = academicYears.value.filter(item => item.schoolId === selectedSchoolId.value)
    selectedYearId.value = currentYearForToday(years)?.id || years[0]?.id
  }

  watch(selectedSchoolId, () => {
    if (!availableAcademicYears.value.some(item => item.id === selectedYearId.value)) {
      selectCurrentYear()
    }
  })

  async function load(force = false) {
    if (loading.value || (loaded.value && !force)) return
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
      selectCurrentYear()
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
    load,
    selectCurrentYear
  }
}
