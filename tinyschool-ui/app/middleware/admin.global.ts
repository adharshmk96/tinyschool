/**
 * Guards the back office. Until the first administrator exists every /admin
 * route funnels into the setup page; afterwards setup is closed and the rest of
 * the console requires an admin session.
 */
export default defineNuxtRouteMiddleware(async (to) => {
  if (!to.path.startsWith('/admin')) return

  const setupPath = '/admin/setup'
  const loginPath = '/admin/login'
  const adminExists = useState<boolean | null>('admin-exists', () => null)

  try {
    const status = await fetchAdminStatus()
    adminExists.value = status.adminExists
  } catch {
    // The API is unreachable. Let the page render and surface its own error
    // rather than bouncing the visitor around.
    return
  }

  if (!adminExists.value)
    return to.path === setupPath ? undefined : navigateTo(setupPath)

  const authenticated = await fetchAdminSession()

  if (to.path === setupPath)
    return navigateTo(authenticated ? '/admin' : loginPath)

  if (to.path === loginPath)
    return authenticated ? navigateTo('/admin') : undefined

  if (authenticated === null)
    return navigateTo({ path: loginPath, query: { redirect: to.fullPath } })
})
