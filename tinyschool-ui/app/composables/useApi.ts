import type { ApiCollection, ApiItem } from '~/types/api'

export function useApi() {
  const config = useRuntimeConfig()
  const baseURL = String(config.public.apiBase).replace(/\/$/, '')

  async function getItem<T>(path: string) {
    return await $fetch<ApiItem<T>>(`${baseURL}${path}`)
  }

  async function getCollection<T>(path: string) {
    return await $fetch<ApiCollection<T>>(`${baseURL}${path}`)
  }

  return { baseURL, getItem, getCollection }
}
