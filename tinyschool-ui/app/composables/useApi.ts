import type { ApiCollection, ApiItem } from '~/types/api'

interface ApiRequestOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'PATCH' | 'DELETE'
  // FormData is used for file uploads, where $fetch sets the multipart headers.
  body?: Record<string, unknown> | FormData
  responseType?: 'blob'
}

export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = String(config.public.apiBase).replace(/\/$/, '')
  const requestHeaders = import.meta.server ? useRequestHeaders(['cookie']) : undefined

  async function rawRequest<T>(path: string, options: ApiRequestOptions = {}) {
    return await $fetch<T>(`${baseURL}${path}`, {
      ...options,
      credentials: 'include',
      headers: requestHeaders
    })
  }

  async function request<T>(path: string, options: ApiRequestOptions = {}) {
    try {
      return await rawRequest<T>(path, options)
    } catch (error) {
      if (path.startsWith('/auth/') || apiErrorStatus(error) !== 401) throw error
      try {
        await rawRequest('/auth/refresh', { method: 'POST' })
        return await rawRequest<T>(path, options)
      } catch {
        useCookie('tiny-school-authenticated').value = null
        if (import.meta.client) await navigateTo('/login')
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

  async function patchItem<T>(path: string, body: Record<string, unknown>) {
    return await request<ApiItem<T>>(path, { method: 'PATCH', body })
  }

  async function put(path: string, body: Record<string, unknown>) {
    await request<unknown>(path, { method: 'PUT', body })
  }

  async function post(path: string) {
    await request<unknown>(path, { method: 'POST' })
  }

  async function download(path: string) {
    return await request<Blob>(path, { responseType: 'blob' })
  }

  async function upload<T>(path: string, body: FormData) {
    return await request<ApiItem<T>>(path, { method: 'POST', body })
  }

  return { baseURL, request, getItem, getCollection, postItem, patchItem, put, post, download, upload }
}

export function apiErrorMessage(error: unknown, fallback: string) {
  if (!error || typeof error !== 'object') return fallback
  const data = 'data' in error ? error.data : undefined
  if (!data || typeof data !== 'object' || !('error' in data)) return fallback
  const apiError = data.error
  if (!apiError || typeof apiError !== 'object' || !('message' in apiError)) return fallback
  return typeof apiError.message === 'string' ? apiError.message : fallback
}

function apiErrorStatus(error: unknown) {
  if (!error || typeof error !== 'object') return undefined
  if ('statusCode' in error && typeof error.statusCode === 'number') return error.statusCode
  if (!('response' in error) || !error.response || typeof error.response !== 'object') return undefined
  if (!('status' in error.response) || typeof error.response.status !== 'number') return undefined
  return error.response.status
}
