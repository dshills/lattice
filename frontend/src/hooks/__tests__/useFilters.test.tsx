import { describe, it, expect } from "vitest";
import { renderHook, act } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { useFilters } from "../useFilters";
import type { ReactNode } from "react";

function wrapper({ children }: { children: ReactNode }) {
  return <MemoryRouter>{children}</MemoryRouter>;
}

describe("useFilters", () => {
  it("starts with empty filters", () => {
    const { result } = renderHook(() => useFilters(), { wrapper });
    expect(result.current.filters).toEqual({});
    expect(result.current.activeFilterCount).toBe(0);
  });

  it("sets a filter", () => {
    const { result } = renderHook(() => useFilters(), { wrapper });
    act(() => result.current.setFilter("state", "InProgress"));
    expect(result.current.filters.state).toBe("InProgress");
    expect(result.current.activeFilterCount).toBe(1);
  });

  it("clears a filter by setting undefined", () => {
    const { result } = renderHook(() => useFilters(), { wrapper });
    act(() => result.current.setFilter("state", "InProgress"));
    act(() => result.current.setFilter("state", undefined));
    expect(result.current.filters.state).toBeUndefined();
  });

  it("clears all filters", () => {
    const { result } = renderHook(() => useFilters(), { wrapper });
    act(() => {
      result.current.setFilter("state", "InProgress");
      result.current.setFilter("type", "bug");
    });
    act(() => result.current.clearFilters());
    expect(result.current.filters).toEqual({});
    expect(result.current.activeFilterCount).toBe(0);
  });

  it("enforces mutual exclusion: is_blocked clears is_ready", () => {
    const { result } = renderHook(() => useFilters(), { wrapper });
    act(() => result.current.setFilter("is_ready", true));
    act(() => result.current.setFilter("is_blocked", true));
    expect(result.current.filters.is_blocked).toBe(true);
    expect(result.current.filters.is_ready).toBeUndefined();
  });

  it("enforces mutual exclusion: is_ready clears is_blocked", () => {
    const { result } = renderHook(() => useFilters(), { wrapper });
    act(() => result.current.setFilter("is_blocked", true));
    act(() => result.current.setFilter("is_ready", true));
    expect(result.current.filters.is_ready).toBe(true);
    expect(result.current.filters.is_blocked).toBeUndefined();
  });
});
