export interface ApiItem<T> {
  data: T
}

export interface ApiCollection<T> {
  data: T[]
  meta?: { total: number }
}

export interface User {
  id: string
  name: string
  email: string
}

export interface Overview {
  students: number
  classes: number
  assignments: number
  exams: number
}

export interface School {
  id: string
  name: string
  grades: string[]
}

export interface AcademicSegment {
  id?: string
  name: string
  type: 'term' | 'vacation'
  durationDays: number
}

export interface AcademicYear {
  id: string
  name: string
  startDate: string
  durationDays: number
  isCurrent?: boolean
  segments: AcademicSegment[]
}
