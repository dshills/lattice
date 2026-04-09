import { useCallback, useMemo, useState } from "react";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
  type DragStartEvent,
  type DragEndEvent,
} from "@dnd-kit/core";
import { useWorkItems } from "../hooks/useWorkItems";
import { useWorkItemMutations } from "../hooks/useWorkItemMutations";
import { useFilters } from "../hooks/useFilters";
import { BoardColumn } from "../components/workitems/BoardColumn";
import { WorkItemCard } from "../components/workitems/WorkItemCard";
import { FilterPanel } from "../components/filters/FilterPanel";
import { LoadingState } from "../components/common/LoadingState";
import { ErrorState } from "../components/common/ErrorState";
import { STATES } from "../lib/constants";
import type { WorkItem, WorkItemState } from "../lib/types";

const ALL_STATES: WorkItemState[] = ["NotDone", "InProgress", "Completed"];

function getAllowedTargets(
  fromState: WorkItemState,
): WorkItemState[] {
  return ALL_STATES.filter((s) => s !== fromState);
}

export function BoardPage() {
  const { filters, setFilter, clearFilters, activeFilterCount } = useFilters();
  const { data, isLoading, error, refetch } = useWorkItems({
    ...filters,
    page_size: 200,
  });
  const { updateMutation } = useWorkItemMutations();
  const [activeItem, setActiveItem] = useState<WorkItem | null>(null);
  const [showFilters, setShowFilters] = useState(false);

  const sensors = useSensors(
    useSensor(PointerSensor, { activationConstraint: { distance: 8 } }),
  );

  const columns = useMemo(() => {
    const items = data?.items ?? [];
    return {
      NotDone: items.filter((i) => i.state === "NotDone"),
      InProgress: items.filter((i) => i.state === "InProgress"),
      Completed: items.filter((i) => i.state === "Completed"),
    };
  }, [data]);

  const allowedTargets = useMemo(() => {
    if (!activeItem) return new Set<WorkItemState>();
    return new Set(getAllowedTargets(activeItem.state));
  }, [activeItem]);

  const handleDragStart = useCallback((event: DragStartEvent) => {
    const item = event.active.data.current?.item as WorkItem | undefined;
    setActiveItem(item ?? null);
  }, []);

  const handleDragEnd = useCallback(
    (event: DragEndEvent) => {
      setActiveItem(null);
      const { active, over } = event;
      if (!over) return;

      const item = active.data.current?.item as WorkItem | undefined;
      if (!item) return;

      const targetState = over.id as WorkItemState;
      if (targetState === item.state) return;

      const isForwardTransition =
        (item.state === "NotDone" && targetState === "InProgress") ||
        (item.state === "InProgress" && targetState === "Completed");

      updateMutation.mutate({
        id: item.id,
        input: {
          state: targetState,
          override: !isForwardTransition,
        },
      });
    },
    [updateMutation],
  );

  if (isLoading) {
    return <LoadingState />;
  }

  if (error) {
    return <ErrorState message={error.message} onRetry={() => refetch()} />;
  }

  return (
    <div>
      <div className="mb-4 flex items-center gap-3">
        <button
          onClick={() => setShowFilters(!showFilters)}
          className={`rounded-md px-3 py-1.5 text-sm font-medium ${
            activeFilterCount > 0
              ? "bg-blue-100 text-blue-800"
              : "bg-gray-100 text-gray-600 hover:bg-gray-200"
          }`}
        >
          Filters{activeFilterCount > 0 ? ` (${activeFilterCount})` : ""}
        </button>
      </div>

      <div className="flex gap-6">
        {showFilters && (
          <div className="w-64 flex-shrink-0">
            <FilterPanel
              filters={filters}
              setFilter={setFilter}
              clearFilters={clearFilters}
              activeFilterCount={activeFilterCount}
            />
          </div>
        )}

        <DndContext
          sensors={sensors}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
        >
          <div className="flex gap-4 overflow-x-auto pb-4">
            {STATES.map((state) => (
              <BoardColumn
                key={state}
                state={state}
                items={columns[state]}
                disabled={
                  activeItem !== null &&
                  !allowedTargets.has(state) &&
                  state !== activeItem.state
                }
              />
            ))}
          </div>

          <DragOverlay>
            {activeItem ? <WorkItemCard item={activeItem} /> : null}
          </DragOverlay>
        </DndContext>
      </div>
    </div>
  );
}
