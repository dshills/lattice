import type { ApiError } from "../types";
import { getAccessToken, setAccessToken } from "../authToken";
import { refresh } from "./auth";

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

async function doFetch<T>(
  path: string,
  options: RequestInit,
): Promise<T> {
  const response = await fetch(path, options);

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

function buildOptions(options: RequestInit): RequestInit {
  const headers = new Headers(options.headers as HeadersInit);
  headers.set("Content-Type", "application/json");

  const token = getAccessToken();
  if (token) {
    headers.set("Authorization", `Bearer ${token}`);
  }

  return { ...options, headers, credentials: "include" };
}

// Shared refresh promise to prevent concurrent refresh requests.
let refreshPromise: Promise<string> | null = null;

async function refreshToken(): Promise<string> {
  if (refreshPromise) {
    return refreshPromise;
  }
  refreshPromise = refresh()
    .then((res) => {
      setAccessToken(res.access_token);
      return res.access_token;
    })
    .finally(() => {
      refreshPromise = null;
    });
  return refreshPromise;
}

export async function apiFetch<T>(
  path: string,
  options: RequestInit = {},
): Promise<T> {
  try {
    return await doFetch<T>(path, buildOptions(options));
  } catch (err) {
    if (err instanceof ApiClientError && err.status === 401) {
      // Attempt token refresh and retry once.
      try {
        await refreshToken();
        return await doFetch<T>(path, buildOptions(options));
      } catch {
        // Refresh failed — redirect to login.
        setAccessToken(null);
        window.location.href = "/login";
        throw err;
      }
    }
    throw err;
  }
}
