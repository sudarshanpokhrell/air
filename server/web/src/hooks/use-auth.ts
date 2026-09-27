import { api, isApiError } from "@/lib/api"
import type { ChangePasswordInput, LoginInput, User } from "@/types/auth"
import {
  queryOptions,
  useMutation,
  useQueryClient,
  useSuspenseQuery,
} from "@tanstack/react-query"
import { useRouter } from "@tanstack/react-router"

type UserResponse = { user: User }

/** The signed-in user, or null when there's no valid session. */
export const meQuery = queryOptions({
  queryKey: ["me"],
  queryFn: async (): Promise<User | null> => {
    try {
      const { user } = await api.get("me").json<UserResponse>()
      return user
    } catch (e) {
      if (isApiError(e, 401)) return null
      throw e
    }
  },
  staleTime: Infinity,
  retry: false,
})

export function useUser() {
  const { data } = useSuspenseQuery(meQuery)
  return data
}

export function useLogin() {
  const client = useQueryClient()

  return useMutation({
    mutationFn: (input: LoginInput) =>
      api.post("auth/login", { json: input }).json<UserResponse>(),
    onSuccess: ({ user }) => {
      client.setQueryData(meQuery.queryKey, user)
    },
  })
}

export function useUpdateMe() {
  const client = useQueryClient()

  return useMutation({
    mutationFn: (input: { name: string }) =>
      api.patch("me", { json: input }).json<UserResponse>(),
    onSuccess: ({ user }) => {
      client.setQueryData(meQuery.queryKey, user)
    },
  })
}

/** Logs out every other session; this browser keeps a fresh one. */
export function useChangePassword() {
  return useMutation({
    mutationFn: async (input: ChangePasswordInput) => {
      await api.post("me/password", { json: input })
    },
  })
}

export function useLogout() {
  const client = useQueryClient()
  const router = useRouter()

  return useMutation({
    mutationFn: () => api.post("auth/logout"),
    onSettled: async () => {
      client.clear()
      client.setQueryData(meQuery.queryKey, null)
      await router.navigate({ to: "/login", replace: true })
    },
  })
}
