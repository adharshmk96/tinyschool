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

export interface AdminUser {
  id: string
  name: string
  email: string
  role: 'user' | 'admin'
  blocked: boolean
  blockedAt?: string
  createdAt?: string
}

export interface AdminStatus {
  adminExists: boolean
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

export interface StudentGrade {
  academicYearId: string
  academicYearName?: string
  grade: string
  isCurrent?: boolean
}

export interface AcademicSegment {
  id?: string
  name: string
  type: 'term' | 'vacation'
  durationDays: number
}

export interface AcademicYear {
  id: string
  schoolId?: string
  name: string
  startDate: string
  durationDays: number
  isCurrent?: boolean
  segments: AcademicSegment[]
}
