import { describe, it, expect, beforeEach } from "vitest";
import { getRole, isAdmin } from "../config";

describe("config", () => {
  beforeEach(() => {
    window.__LATTICE_CONFIG__ = undefined;
  });

  it("returns 'user' when config is not set", () => {
    expect(getRole()).toBe("user");
  });

  it("returns 'admin' when config sets admin role", () => {
    window.__LATTICE_CONFIG__ = { role: "admin" };
    expect(getRole()).toBe("admin");
  });

  it("returns 'user' when config sets user role", () => {
    window.__LATTICE_CONFIG__ = { role: "user" };
    expect(getRole()).toBe("user");
  });

  it("isAdmin returns true for admin", () => {
    window.__LATTICE_CONFIG__ = { role: "admin" };
    expect(isAdmin()).toBe(true);
  });

  it("isAdmin returns false for user", () => {
    window.__LATTICE_CONFIG__ = { role: "user" };
    expect(isAdmin()).toBe(false);
  });

  it("isAdmin returns false when config missing", () => {
    expect(isAdmin()).toBe(false);
  });
});
