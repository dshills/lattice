import { describe, it, expect, vi, afterEach } from "vitest";
import { apiFetch, ApiClientError } from "../client";
import { setAccessToken } from "../../authToken";

describe("apiFetch", () => {
  const originalFetch = globalThis.fetch;

  afterEach(() => {
    globalThis.fetch = originalFetch;
    setAccessToken(null);
  });

  it("returns parsed JSON on success", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({ id: "123", title: "Test" }),
    });

    const result = await apiFetch<{ id: string }>("/workitems/123");
    expect(result).toEqual({ id: "123", title: "Test" });
  });

  it("returns undefined on 204", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 204,
    });

    const result = await apiFetch<void>("/workitems/123", {
      method: "DELETE",
    });
    expect(result).toBeUndefined();
  });

  it("throws ApiClientError with parsed error body", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 404,
      json: () =>
        Promise.resolve({
          error: { code: "NOT_FOUND", message: "Item not found" },
        }),
    });

    try {
      await apiFetch("/workitems/bad");
      expect.unreachable("Should have thrown");
    } catch (err) {
      expect(err).toBeInstanceOf(ApiClientError);
      const apiErr = err as ApiClientError;
      expect(apiErr.status).toBe(404);
      expect(apiErr.code).toBe("NOT_FOUND");
      expect(apiErr.message).toBe("Item not found");
    }
  });

  it("throws ApiClientError with defaults on unparseable error", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: false,
      status: 500,
      json: () => Promise.reject(new Error("not json")),
    });

    try {
      await apiFetch("/workitems");
      expect.unreachable("Should have thrown");
    } catch (err) {
      const apiErr = err as ApiClientError;
      expect(apiErr.status).toBe(500);
      expect(apiErr.code).toBe("UNKNOWN");
      expect(apiErr.message).toBe("HTTP 500");
    }
  });

  it("sends Authorization header when token is set", async () => {
    setAccessToken("test-jwt-token");

    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({}),
    });

    await apiFetch("/workitems");

    const callArgs = (globalThis.fetch as ReturnType<typeof vi.fn>).mock
      .calls[0];
    const headers = callArgs[1].headers as Headers;
    expect(headers.get("Authorization")).toBe("Bearer test-jwt-token");
  });

  it("does not send Authorization header when no token", async () => {
    globalThis.fetch = vi.fn().mockResolvedValue({
      ok: true,
      status: 200,
      json: () => Promise.resolve({}),
    });

    await apiFetch("/workitems");

    const callArgs = (globalThis.fetch as ReturnType<typeof vi.fn>).mock
      .calls[0];
    const headers = callArgs[1].headers as Headers;
    expect(headers.get("Authorization")).toBeNull();
  });
});
