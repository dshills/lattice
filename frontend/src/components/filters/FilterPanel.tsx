import type { ListFilter, WorkItemState } from "../../lib/types";
import { STATES, STATE_LABELS } from "../../lib/constants";

type FilterKey = keyof ListFilter;

interface FilterPanelProps {
  filters: ListFilter;
  setFilter: (key: FilterKey, value: string | boolean | number | undefined) => void;
  clearFilters: () => void;
  activeFilterCount: number;
}

export function FilterPanel({
  filters,
  setFilter,
  clearFilters,
  activeFilterCount,
}: FilterPanelProps) {
  return (
    <div className="space-y-4 rounded-lg border border-gray-200 bg-white p-4">
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-semibold text-gray-700">Filters</h3>
        {activeFilterCount > 0 && (
          <button
            onClick={clearFilters}
            className="text-xs text-blue-600 hover:underline"
          >
            Clear all ({activeFilterCount})
          </button>
        )}
      </div>

      {/* State filter */}
      <div>
        <label className="mb-1 block text-xs font-medium text-gray-500">
          State
        </label>
        <div className="flex gap-1">
          {STATES.map((state) => (
            <button
              key={state}
              onClick={() =>
                setFilter(
                  "state",
                  filters.state === state ? undefined : state,
                )
              }
              className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
                filters.state === state
                  ? "bg-blue-100 text-blue-800"
                  : "bg-gray-100 text-gray-600 hover:bg-gray-200"
              }`}
            >
              {STATE_LABELS[state as WorkItemState]}
            </button>
          ))}
        </div>
      </div>

      {/* Type filter */}
      <div>
        <label className="mb-1 block text-xs font-medium text-gray-500">
          Type
        </label>
        <input
          type="text"
          value={filters.type ?? ""}
          onChange={(e) => setFilter("type", e.target.value || undefined)}
          placeholder="Filter by type..."
          className="w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
      </div>

      {/* Tags filter */}
      <div>
        <label className="mb-1 block text-xs font-medium text-gray-500">
          Tags
        </label>
        <input
          type="text"
          value={filters.tags ?? ""}
          onChange={(e) => setFilter("tags", e.target.value || undefined)}
          placeholder="tag1,tag2 (AND logic)"
          className="w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
      </div>

      {/* Blocked / Ready toggles */}
      <div>
        <label className="mb-1 block text-xs font-medium text-gray-500">
          Status
        </label>
        <div className="flex gap-2">
          <button
            onClick={() =>
              setFilter(
                "is_blocked",
                filters.is_blocked ? undefined : true,
              )
            }
            disabled={filters.is_ready === true}
            className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
              filters.is_blocked
                ? "bg-red-100 text-red-800"
                : "bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-40 disabled:cursor-not-allowed"
            }`}
          >
            Blocked
          </button>
          <button
            onClick={() =>
              setFilter("is_ready", filters.is_ready ? undefined : true)
            }
            disabled={filters.is_blocked === true}
            className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
              filters.is_ready
                ? "bg-green-100 text-green-800"
                : "bg-gray-100 text-gray-600 hover:bg-gray-200 disabled:opacity-40 disabled:cursor-not-allowed"
            }`}
          >
            Ready
          </button>
        </div>
      </div>
    </div>
  );
}
