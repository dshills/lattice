import { useMemo, useState, useCallback } from "react";
import { useWorkItems } from "../hooks/useWorkItems";
import { useFilters } from "../hooks/useFilters";
import { FilterPanel } from "../components/filters/FilterPanel";
import { SearchInput } from "../components/filters/SearchInput";
import { WorkItemRow } from "../components/workitems/WorkItemRow";
import { LoadingState } from "../components/common/LoadingState";
import { ErrorState } from "../components/common/ErrorState";
import { EmptyState } from "../components/common/EmptyState";
import type { WorkItem } from "../lib/types";

type SortField = "title" | "state" | "type" | "updated_at";
type SortDir = "asc" | "desc";

function SortHeader({
  field,
  label,
  sortField,
  sortDir,
  onSort,
}: {
  field: SortField;
  label: string;
  sortField: SortField;
  sortDir: SortDir;
  onSort: (field: SortField) => void;
}) {
  return (
    <th
      className="cursor-pointer px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-gray-500 hover:text-gray-700"
      onClick={() => onSort(field)}
    >
      {label} {sortField === field && (sortDir === "asc" ? "^" : "v")}
    </th>
  );
}

export function ListPage() {
  const { filters, setFilter, clearFilters, activeFilterCount } = useFilters();
  const pageSize = filters.page_size ?? 50;
  const [search, setSearch] = useState("");
  const [sortField, setSortField] = useState<SortField>("updated_at");
  const [sortDir, setSortDir] = useState<SortDir>("desc");

  const { data, isLoading, error, refetch } = useWorkItems({
    ...filters,
    page_size: pageSize,
  });

  const handleSort = useCallback(
    (field: SortField) => {
      if (field === sortField) {
        setSortDir(sortDir === "asc" ? "desc" : "asc");
      } else {
        setSortField(field);
        setSortDir("asc");
      }
    },
    [sortField, sortDir],
  );

  const handleSearch = useCallback((q: string) => setSearch(q), []);

  const items = useMemo(() => {
    let result = data?.items ?? [];

    // Client-side search (v1: search within loaded page)
    if (search) {
      const q = search.toLowerCase();
      result = result.filter(
        (item) =>
          item.title.toLowerCase().includes(q) ||
          (item.description ?? "").toLowerCase().includes(q),
      );
    }

    // Client-side sort
    result = [...result].sort((a, b) => {
      const aVal = a[sortField] ?? "";
      const bVal = b[sortField] ?? "";
      const cmp = String(aVal).localeCompare(String(bVal));
      return sortDir === "asc" ? cmp : -cmp;
    });

    return result;
  }, [data, search, sortField, sortDir]);

  return (
    <div className="flex gap-6">
      <div className="w-64 flex-shrink-0">
        <FilterPanel
          filters={filters}
          setFilter={setFilter}
          clearFilters={clearFilters}
          activeFilterCount={activeFilterCount}
        />
      </div>

      <div className="flex-1">
        <div className="mb-4">
          <SearchInput onSearch={handleSearch} placeholder="Search items..." />
        </div>

        {isLoading ? (
          <LoadingState />
        ) : error ? (
          <ErrorState message={error.message} onRetry={() => refetch()} />
        ) : items.length === 0 ? (
          <EmptyState
            title="No items found"
            description="No items match your current filters"
          />
        ) : (
          <>
            <div className="overflow-x-auto rounded-lg border border-gray-200 bg-white">
              <table className="w-full">
                <thead className="border-b border-gray-200 bg-gray-50">
                  <tr>
                    <SortHeader field="title" label="Title" sortField={sortField} sortDir={sortDir} onSort={handleSort} />
                    <SortHeader field="state" label="State" sortField={sortField} sortDir={sortDir} onSort={handleSort} />
                    <SortHeader field="type" label="Type" sortField={sortField} sortDir={sortDir} onSort={handleSort} />
                    <th className="px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-gray-500">
                      Tags
                    </th>
                    <th className="px-3 py-2 text-left text-xs font-semibold uppercase tracking-wide text-gray-500">
                      Relationships
                    </th>
                    <SortHeader field="updated_at" label="Updated" sortField={sortField} sortDir={sortDir} onSort={handleSort} />
                  </tr>
                </thead>
                <tbody>
                  {items.map((item: WorkItem) => (
                    <WorkItemRow key={item.id} item={item} />
                  ))}
                </tbody>
              </table>
            </div>

            <div className="mt-4 flex items-center justify-between text-sm text-gray-500">
              <div className="flex items-center gap-2">
                <span>Page size:</span>
                {[20, 50, 100, 200].map((size) => (
                  <button
                    key={size}
                    onClick={() => setFilter("page_size", size)}
                    className={`rounded px-2 py-0.5 ${
                      pageSize === size
                        ? "bg-blue-100 text-blue-800"
                        : "hover:bg-gray-100"
                    }`}
                  >
                    {size}
                  </button>
                ))}
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() =>
                    setFilter("page", (filters.page ?? 1) - 1)
                  }
                  disabled={!filters.page || filters.page <= 1}
                  className="rounded px-2 py-0.5 hover:bg-gray-100 disabled:opacity-40"
                >
                  Prev
                </button>
                <span>
                  Page {filters.page ?? 1} of{" "}
                  {Math.ceil((data?.total ?? 0) / pageSize) || 1}
                </span>
                <button
                  onClick={() =>
                    setFilter("page", (filters.page ?? 1) + 1)
                  }
                  disabled={
                    (filters.page ?? 1) >=
                    Math.ceil((data?.total ?? 0) / pageSize)
                  }
                  className="rounded px-2 py-0.5 hover:bg-gray-100 disabled:opacity-40"
                >
                  Next
                </button>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
