export type AppRole = "admin" | "developer"

export interface App {
  id: string
  slug: string
  name: string
  platforms?: AppPlatform[] // only from GET /apps/{slug}
  created_at: string
}

export interface AppPlatform {
  platform: "ios" | "android"
  bundle_id: string
  enabled: boolean
  created_at: string
}

export interface CodeSigningStatus {
  enabled: boolean
  key_id: string
}

export interface AppDetail {
  app: App
  my_role: AppRole
  code_signing: CodeSigningStatus
}

export interface AppMember {
  user_id: string
  email: string
  name: string
  role: AppRole
  created_at: string
}

export interface ApiKey {
  id: string
  name: string
  created_by: string
  created_at: string
  revoked_at: string | null
}

export interface Update {
  id: string
  group_id: string
  channel: string
  platform: "ios" | "android"
  runtime_version: string
  kind: "update" | "rollback_to_embedded"
  message: string
  git_commit: string
  rolled_back_at: string | null
  created_at: string
}

export type CreateAppInput = {
  slug: string
  name: string
}

export type AddMemberInput = {
  email: string
  role: AppRole
}
