import { api } from "@/lib/api"
import type {
  AddMemberInput,
  ApiKey,
  App,
  AppMember,
  AppRole,
  CreateAppInput,
  Update,
} from "@/types/apps"
import {
  queryOptions,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query"

// ---------- apps ----------

/** Global admins get every app; everyone else the apps they're a member of. */
export const appsQuery = queryOptions({
  queryKey: ["apps"],
  queryFn: async () => {
    const { apps } = await api.get("apps").json<{ apps: App[] }>()
    return apps
  },
})

export const appQuery = (slug: string) =>
  queryOptions({
    queryKey: ["apps", slug],
    queryFn: () =>
      api.get(`apps/${slug}`).json<{ app: App; my_role: AppRole }>(),
  })

export function useCreateApp() {
  const client = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateAppInput) =>
      api.post("apps", { json: input }).json<{ app: App }>(),
    onSuccess: () => client.invalidateQueries({ queryKey: appsQuery.queryKey }),
  })
}

// ---------- members ----------

export const membersQuery = (slug: string) =>
  queryOptions({
    queryKey: ["apps", slug, "members"],
    queryFn: async () => {
      const { members } = await api
        .get(`apps/${slug}/members`)
        .json<{ members: AppMember[] }>()
      return members
    },
  })

export function useAddMember(slug: string) {
  const client = useQueryClient()

  return useMutation({
    mutationFn: async (input: AddMemberInput) => {
      await api.post(`apps/${slug}/members`, { json: input })
    },
    onSuccess: () =>
      client.invalidateQueries({ queryKey: membersQuery(slug).queryKey }),
  })
}

export function useUpdateMember(slug: string) {
  const client = useQueryClient()

  return useMutation({
    mutationFn: async ({ userId, role }: { userId: string; role: AppRole }) => {
      await api.patch(`apps/${slug}/members/${userId}`, { json: { role } })
    },
    onSuccess: () =>
      client.invalidateQueries({ queryKey: membersQuery(slug).queryKey }),
  })
}

export function useRemoveMember(slug: string) {
  const client = useQueryClient()

  return useMutation({
    mutationFn: async (userId: string) => {
      await api.delete(`apps/${slug}/members/${userId}`)
    },
    onSuccess: () =>
      client.invalidateQueries({ queryKey: membersQuery(slug).queryKey }),
  })
}

// ---------- API keys ----------

/** App admins see every key; developers only their own. */
export const apiKeysQuery = (slug: string) =>
  queryOptions({
    queryKey: ["apps", slug, "api-keys"],
    queryFn: async () => {
      const { api_keys } = await api
        .get(`apps/${slug}/api-keys`)
        .json<{ api_keys: ApiKey[] }>()
      return api_keys
    },
  })

/** The plain key is only in this response; the server keeps its hash. */
export function useCreateApiKey(slug: string) {
  const client = useQueryClient()

  return useMutation({
    mutationFn: (input: { name: string }) =>
      api
        .post(`apps/${slug}/api-keys`, { json: input })
        .json<{ api_key: ApiKey; key: string }>(),
    onSuccess: () =>
      client.invalidateQueries({ queryKey: apiKeysQuery(slug).queryKey }),
  })
}

export function useRevokeApiKey(slug: string) {
  const client = useQueryClient()

  return useMutation({
    mutationFn: async (keyId: string) => {
      await api.delete(`apps/${slug}/api-keys/${keyId}`)
    },
    onSuccess: () =>
      client.invalidateQueries({ queryKey: apiKeysQuery(slug).queryKey }),
  })
}

// ---------- updates ----------

export const updatesQuery = (slug: string) =>
  queryOptions({
    queryKey: ["apps", slug, "updates"],
    queryFn: async () => {
      const { updates } = await api
        .get(`apps/${slug}/updates`)
        .json<{ updates: Update[] }>()
      return updates
    },
  })
