import type { AdminUser, ApiCollection, ApiItem, AdminStatus } from '~/types/api'

interface AdminRequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  body?: Record<string, unknown>
  responseType?: 'blob'
}

/**
 * The back office talks to the same API with its own session cookie, so it
 * needs its own refresh and its own redirect target on failure.
 */
export function useAdminApi() {
  const config = useRuntimeConfig()
  const baseURL = String(config.public.apiBase).replace(/\/$/, '')
  const requestHeaders = import.meta.server ? useRequestHeaders(['cookie']) : undefined

  async function rawRequest<T>(path: string, options: AdminRequestOptions = {}) {
    return await $fetch<T>(`${baseURL}/admin${path}`, {
      ...options,
      credentials: 'include',
      headers: requestHeaders
    })
  }

  async function request<T>(path: string, options: AdminRequestOptions = {}) {
    try {
      return await rawRequest<T>(path, options)
    } catch (error) {
      const unauthenticated = ['/login', '/setup', '/status', '/refresh'].includes(path)
      if (unauthenticated || apiErrorStatus(error) !== 401) throw error
      try {
        await rawRequest('/refresh', { method: 'POST' })
        return await rawRequest<T>(path, options)
      } catch {
        if (import.meta.client) await navigateTo('/admin/login')
        throw error
      }
    }
  }

  async function getItem<T>(path: string) {
    return await request<ApiItem<T>>(path)
  }

  async function getCollection<T>(path: string) {
    return await request<ApiCollection<T>>(path)
  }

  async function postItem<T>(path: string, body?: Record<string, unknown>) {
    return await request<ApiItem<T>>(path, { method: 'POST', body })
  }

  async function download(path: string) {
    return await request<Blob>(path, { responseType: 'blob' })
  }

  return { baseURL, request, getItem, getCollection, postItem, download }
}

/** Reads whether the first administrator still has to be created. */
export async function fetchAdminStatus() {
  const config = useRuntimeConfig()
  const baseURL = String(config.public.apiBase).replace(/\/$/, '')
  const response = await $fetch<ApiItem<AdminStatus>>(`${baseURL}/admin/status`)
  return response.data
}

/** Validates the real server-side admin session used by protected API routes. */
export async function fetchAdminSession() {
  const config = useRuntimeConfig()
  const baseURL = String(config.public.apiBase).replace(/\/$/, '')
  const headers = import.meta.server ? useRequestHeaders(['cookie']) : undefined

  try {
    const response = await $fetch<ApiItem<AdminUser>>(`${baseURL}/admin/me`, {
      credentials: 'include',
      headers
    })
    return response.data
  } catch {
    return null
  }
}

function apiErrorStatus(error: unknown) {
  if (!error || typeof error !== 'object') return undefined
  if ('statusCode' in error && typeof error.statusCode === 'number') return error.statusCode
  if (!('response' in error) || !error.response || typeof error.response !== 'object') return undefined
  if (!('status' in error.response) || typeof error.response.status !== 'number') return undefined
  return error.response.status
}
