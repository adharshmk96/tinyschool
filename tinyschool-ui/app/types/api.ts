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

export interface BackupSettings {
  enabled: boolean
  frequency: 'daily' | 'every_2_days' | 'weekly'
  runAt: string
  maxBackups: number
  nextRunAt?: string
}

export interface DatabaseBackup {
  name: string
  size: number
  createdAt: string
}

export interface Overview {
  students: number
  classes: number
  assignments: number
  exams: number
  upcoming: UpcomingItem[]
}

export interface UpcomingItem {
  id: string
  kind: 'assignment' | 'exam'
  title: string
  date: string
  className?: string
  studentCount: number
}

export interface School {
  id: string
  name: string
  classrooms: string[]
  classroomsInUse?: string[]
}

export interface StudentClassroom {
  academicYearId: string
  academicYearName?: string
  classroom: string
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
  endDate?: string
  durationDays: number
  isCurrent?: boolean
  segments: AcademicSegment[]
}

export interface ImportSummary {
  schools: number
  academicYears: number
  students: number
  studentLogs: number
  classes: number
  assignments: number
  exams: number
  scores: number
}
