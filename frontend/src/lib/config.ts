interface LatticeConfig {
  role: "admin" | "user";
}

declare global {
  interface Window {
    __LATTICE_CONFIG__?: LatticeConfig;
  }
}

export function getRole(): "admin" | "user" {
  return window.__LATTICE_CONFIG__?.role ?? "user";
}

export function isAdmin(): boolean {
  return getRole() === "admin";
}
