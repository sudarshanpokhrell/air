import { meQuery } from "@/hooks/use-auth"
import { api } from "@/lib/api"
import type { CreateUserInput, UpdateUserInput, User } from "@/types/auth"
import {
  queryOptions,
  useMutation,
  useQueryClient,
} from "@tanstack/react-query"

/** Every user on the server. Global admins only. */
export const usersQuery = queryOptions({
  queryKey: ["users"],
  queryFn: async () => {
    const { users } = await api.get("users").json<{ users: User[] }>()
    return users
  },
})

type UserResponse = { user: User }

export function useCreateUser() {
  const client = useQueryClient()

  return useMutation({
    mutationFn: (input: CreateUserInput) =>
      api.post("users", { json: input }).json<UserResponse>(),
    onSuccess: () =>
      client.invalidateQueries({ queryKey: usersQuery.queryKey }),
  })
}

/** Rename, make/remove global admin, deactivate/reactivate. */
export function useUpdateUser() {
  const client = useQueryClient()

  return useMutation({
    mutationFn: ({ id, ...input }: UpdateUserInput) =>
      api.patch(`users/${id}`, { json: input }).json<UserResponse>(),
    onSuccess: () => {
      client.invalidateQueries({ queryKey: usersQuery.queryKey })
      client.invalidateQueries({ queryKey: meQuery.queryKey })
    },
  })
}
