import ky, { isHTTPError } from "ky"

export class ApiError extends Error {
  status: number
  fields?: Record<string, string>

  constructor(
    status: number,
    message: string,
    fields?: Record<string, string>
  ) {
    super(message)
    this.name = "ApiError"
    this.status = status
    this.fields = fields
  }
}

export function isApiError(e: unknown, status?: number): e is ApiError {
  return e instanceof ApiError && (status === undefined || e.status === status)
}

export function getErrorMessage(e: unknown): string {
  return e instanceof Error ? e.message : "Something went wrong."
}

let unauthorizedHandler = () => {}

export function onUnauthorized(handler: () => void) {
  unauthorizedHandler = handler
}

export const api = ky.create({
  prefix: "/api/v1",
  timeout: 10000,
  retry: { limit: 2 },
  hooks: {
    beforeError: [
      ({ request, error }) => {
        if (!isHTTPError(error)) return error
        const isAuthRoute = new URL(request.url).pathname.startsWith(
          "/api/v1/auth/"
        )
        if (error.response.status === 401 && !isAuthRoute) unauthorizedHandler()

        // The API answers {"error": "message"} or, for validation, {"error": {field: message}}.
        const detail = (
          error.data as { error?: string | Record<string, string> }
        )?.error
        const fields = typeof detail === "object" ? detail : undefined
        const message =
          typeof detail === "object" ? Object.values(detail).join(" ") : detail
        return new ApiError(
          error.response.status,
          message ?? error.response.statusText,
          fields
        )
      },
    ],
  },
})
