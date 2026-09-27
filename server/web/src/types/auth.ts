export interface User {
  id: string
  name: string
  email: string
  is_admin: boolean
  is_active: boolean
  created_at: string
  updated_at: string
}

export type LoginInput = {
  email: string
  password: string
}

export type CreateUserInput = {
  name: string
  email: string
  password: string
  is_admin: boolean
}

export type UpdateUserInput = {
  id: string
  name?: string
  is_admin?: boolean
  is_active?: boolean
}

export type ChangePasswordInput = {
  current_password: string
  new_password: string
}
