import { useCallback, useMemo } from "react";
import { useSearchParams } from "react-router";
import type { ListFilter, WorkItemState } from "../lib/types";

type FilterKey = keyof ListFilter;

export function useFilters() {
  const [searchParams, setSearchParams] = useSearchParams();

  const filters: ListFilter = useMemo(() => {
    const f: ListFilter = {};
    const state = searchParams.get("state");
    if (state) f.state = state as WorkItemState;
    const type = searchParams.get("type");
    if (type) f.type = type;
    const tags = searchParams.get("tags");
    if (tags) f.tags = tags;
    const parentId = searchParams.get("parent_id");
    if (parentId) f.parent_id = parentId;
    const assigneeId = searchParams.get("assignee_id");
    if (assigneeId) f.assignee_id = assigneeId;
    const isBlocked = searchParams.get("is_blocked");
    if (isBlocked) f.is_blocked = isBlocked === "true";
    const isReady = searchParams.get("is_ready");
    if (isReady) f.is_ready = isReady === "true";
    const page = searchParams.get("page");
    if (page) f.page = Number(page);
    const pageSize = searchParams.get("page_size");
    if (pageSize) f.page_size = Number(pageSize);
    return f;
  }, [searchParams]);

  const setFilter = useCallback(
    (key: FilterKey, value: string | boolean | number | undefined) => {
      setSearchParams((prev) => {
        const next = new URLSearchParams(prev);
        if (value === undefined || value === "") {
          next.delete(String(key));
        } else {
          next.set(String(key), String(value));
        }

        // Mutual exclusion: is_blocked and is_ready
        if (key === "is_blocked" && value) {
          next.delete("is_ready");
        }
        if (key === "is_ready" && value) {
          next.delete("is_blocked");
        }

        // Reset to page 1 when filters change
        if (key !== "page") {
          next.delete("page");
        }

        return next;
      });
    },
    [setSearchParams],
  );

  const clearFilters = useCallback(() => {
    setSearchParams({});
  }, [setSearchParams]);

  const activeFilterCount = useMemo(() => {
    let count = 0;
    if (filters.state) count++;
    if (filters.type) count++;
    if (filters.tags) count++;
    if (filters.parent_id) count++;
    if (filters.assignee_id) count++;
    if (filters.is_blocked !== undefined) count++;
    if (filters.is_ready !== undefined) count++;
    return count;
  }, [filters]);

  return { filters, setFilter, clearFilters, activeFilterCount };
}
