import type { ApiError } from "../types";
import { getRole } from "../config";

export class ApiClientError extends Error {
  status: number;
  code: string;

  constructor(status: number, code: string, message: string) {
    super(message);
    this.name = "ApiClientError";
    this.status = status;
    this.code = code;
  }
}

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  const headers = new Headers(options.headers as HeadersInit);
  headers.set("Content-Type", "application/json");

  const role = getRole();
  if (role === "admin") {
    headers.set("X-Role", "admin");
  }

  const response = await fetch(path, {
    ...options,
    headers,
  });

  if (!response.ok) {
    let code = "UNKNOWN";
    let message = `HTTP ${response.status}`;

    try {
      const body = (await response.json()) as ApiError;
      code = body.error.code;
      message = body.error.message;
    } catch {
      // Use defaults
    }

    throw new ApiClientError(response.status, code, message);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}
