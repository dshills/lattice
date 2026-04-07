import { useCallback, useMemo } from "react";
import {
  DndContext,
  DragOverlay,
  PointerSensor,
  useSensor,
  useSensors,
  type DragStartEvent,
  type DragEndEvent,
} from "@dnd-kit/core";
import { useState } from "react";
import { useWorkItems } from "../hooks/useWorkItems";
import { useWorkItemMutations } from "../hooks/useWorkItemMutations";
import { BoardColumn } from "../components/workitems/BoardColumn";
import { WorkItemCard } from "../components/workitems/WorkItemCard";
import { STATES } from "../lib/constants";
import { isAdmin } from "../lib/config";
import type { WorkItem, WorkItemState } from "../lib/types";

const VALID_TRANSITIONS: Record<WorkItemState, WorkItemState[]> = {
  NotDone: ["InProgress"],
  InProgress: ["Completed"],
  Completed: [],
};

const ADMIN_BACKWARD: Record<WorkItemState, WorkItemState[]> = {
  NotDone: [],
  InProgress: ["NotDone"],
  Completed: ["InProgress"],
};

function getAllowedTargets(
  fromState: WorkItemState,
  admin: boolean,
): WorkItemState[] {
  const forward = VALID_TRANSITIONS[fromState];
  const backward = admin ? ADMIN_BACKWARD[fromState] : [];
  return [...forward, ...backward];
}

export function BoardPage() {
  const { data, isLoading, error } = useWorkItems({ page_size: 200 });
  const { updateMutation } = useWorkItemMutations();
  const [activeItem, setActiveItem] = useState<WorkItem | null>(null);

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

  const admin = isAdmin();

  const allowedTargets = useMemo(() => {
    if (!activeItem) return new Set<WorkItemState>();
    return new Set(getAllowedTargets(activeItem.state, admin));
  }, [activeItem, admin]);

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

      const allowed = getAllowedTargets(item.state, admin);
      if (!allowed.includes(targetState)) return;

      const isBackward =
        ADMIN_BACKWARD[item.state]?.includes(targetState) ?? false;

      updateMutation.mutate({
        id: item.id,
        input: {
          state: targetState,
          ...(isBackward ? { override: true } : {}),
        },
      });
    },
    [admin, updateMutation],
  );

  if (isLoading) {
    return <p className="text-gray-500">Loading board...</p>;
  }

  if (error) {
    return (
      <p className="text-red-500">Failed to load board: {error.message}</p>
    );
  }

  return (
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
  );
}
