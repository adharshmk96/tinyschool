import { mkdir, readdir, unlink } from 'node:fs/promises'
import { join } from 'node:path'
import { expect, test } from '@playwright/test'
import { mockApi } from './fake-api'

const output = join(import.meta.dirname, 'output')
const nuxtCookie = (value: string) => encodeURIComponent(JSON.stringify(value))
const modes = ['light', 'dark'] as const
const routes = [
  ['01-home', '/'],
  ['02-register', '/register'],
  ['03-login', '/login'],
  ['04-overview', '/dashboard'],
  ['05-student', '/dashboard/students'],
  ['06-student-detail', '/dashboard/students/student-1/performance'],
  ['07-myclasses', '/dashboard/classes'],
  ['08-myclasses-detail', '/dashboard/classes/class-1/performance'],
  ['09-assignments', '/dashboard/assignments'],
  ['10-assignments-detail', '/dashboard/assignments/assignment-1'],
  ['11-exams', '/dashboard/exams'],
  ['12-exams-detail', '/dashboard/exams/exam-1/performance'],
  ['13-settings-account', '/dashboard/settings/account'],
  ['14-settings-school', '/dashboard/settings/schools'],
  ['15-settings-academic-year', '/dashboard/settings/academic-years'],
  ['16-admin-setup', '/admin/setup'],
  ['17-admin-login', '/admin/login'],
  ['18-admin-overview', '/admin'],
  ['19-admin-users', '/admin/users'],
  ['20-admin-smtp', '/admin/smtp'],
  ['21-admin-backups', '/admin/backups'],
  ['22-student-detail-behaviour', '/dashboard/students/student-1/behaviour'],
  ['23-student-detail-assignments', '/dashboard/students/student-1/assignments'],
  ['24-student-detail-exams', '/dashboard/students/student-1/exams'],
  ['25-student-detail-notes', '/dashboard/students/student-1/notes'],
  ['26-myclasses-detail-assignments', '/dashboard/classes/class-1/assignments'],
  ['27-myclasses-detail-exams', '/dashboard/classes/class-1/exams'],
  ['28-myclasses-detail-students', '/dashboard/classes/class-1/students'],
  ['29-exams-detail-students', '/dashboard/exams/exam-1/students'],
  ['30-settings-academic-year-create', '/dashboard/settings/academic-years/create'],
  ['31-settings-academic-year-detail', '/dashboard/settings/academic-years/year-1'],
  ['32-settings-academic-year-edit', '/dashboard/settings/academic-years/year-1/edit'],
  ['33-settings-data', '/dashboard/settings/data']
] as const

test.beforeAll(async () => {
  await mkdir(output, { recursive: true })
  for (const directory of [output, ...modes.map(mode => join(output, mode))]) {
    await mkdir(directory, { recursive: true })
    const files = await readdir(directory)
    await Promise.all(
      files
        .filter(file => file.endsWith('.png'))
        .map(file => unlink(join(directory, file)))
    )
  }
})

for (const mode of modes) {
  for (const [name, path] of routes) {
    test(`capture ${mode}/${name}`, async ({ page, context }) => {
      await page.emulateMedia({ colorScheme: mode })
      await mockApi(page, path === '/admin/setup')
      await context.addCookies([
        { name: 'tiny-school-authenticated', value: nuxtCookie('true'), url: 'http://127.0.0.1:3000' },
        { name: 'tiny-school-school', value: nuxtCookie('school-1'), url: 'http://127.0.0.1:3000' },
        { name: 'tiny-school-year', value: nuxtCookie('year-1'), url: 'http://127.0.0.1:3000' }
      ])
      if (path.startsWith('/admin') && !['/admin/login', '/admin/setup'].includes(path)) {
        await context.addCookies([{ name: 'tiny-school-admin', value: nuxtCookie('true'), url: 'http://127.0.0.1:3000' }])
      }

      await page.goto(path, { waitUntil: 'networkidle' })
      expect(new URL(page.url()).pathname).toBe(path)
      await expect(page.locator('body')).toBeVisible()
      await page.screenshot({ path: join(output, mode, `${name}.png`), fullPage: false })
    })
  }
}
