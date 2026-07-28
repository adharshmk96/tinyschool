export default defineNuxtRouteMiddleware((to) => {
  if (!to.path.startsWith('/dashboard')) return
  const authenticated = useCookie('tiny-school-authenticated')
  if (authenticated.value !== 'true') {
    return navigateTo({ path: '/login', query: { redirect: to.fullPath } })
  }
})
