import type { User } from "../types";

export interface AuthResponse {
  user: User;
  access_token: string;
}

export interface RegisterInput {
  email: string;
  display_name: string;
  password: string;
}

export interface LoginInput {
  email: string;
  password: string;
}

export interface UpdateMeInput {
  display_name?: string;
  password?: string;
  current_password?: string;
}

async function authFetch<T>(path: string, options: RequestInit = {}): Promise<T> {
  const headers = new Headers(options.headers as HeadersInit);
  headers.set("Content-Type", "application/json");

  const response = await fetch(path, { ...options, headers });

  if (!response.ok) {
    let message = `HTTP ${response.status}`;
    try {
      const body = await response.json();
      message = body.error?.message ?? message;
    } catch {
      // Use default
    }
    throw new Error(message);
  }

  if (response.status === 204) {
    return undefined as T;
  }

  return response.json() as Promise<T>;
}

export function register(input: RegisterInput): Promise<AuthResponse> {
  return authFetch<AuthResponse>("/auth/register", {
    method: "POST",
    body: JSON.stringify(input),
    credentials: "include",
  });
}

export function login(input: LoginInput): Promise<AuthResponse> {
  return authFetch<AuthResponse>("/auth/login", {
    method: "POST",
    body: JSON.stringify(input),
    credentials: "include",
  });
}

export function refresh(): Promise<{ access_token: string }> {
  return authFetch<{ access_token: string }>("/auth/refresh", {
    method: "POST",
    credentials: "include",
  });
}

export function getMe(token: string): Promise<User> {
  return authFetch<User>("/users/me", {
    headers: { Authorization: `Bearer ${token}` },
  });
}

export function updateMe(token: string, input: UpdateMeInput): Promise<User> {
  return authFetch<User>("/users/me", {
    method: "PATCH",
    body: JSON.stringify(input),
    headers: { Authorization: `Bearer ${token}` },
  });
}
