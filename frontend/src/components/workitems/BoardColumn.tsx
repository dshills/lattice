import { useDroppable } from "@dnd-kit/core";
import type { WorkItemState, WorkItem } from "../../lib/types";
import { STATE_LABELS } from "../../lib/constants";
import { DraggableCard } from "./DraggableCard";
import { QuickAdd } from "./QuickAdd";

interface BoardColumnProps {
  state: WorkItemState;
  items: WorkItem[];
  disabled: boolean;
  readOnly?: boolean;
}

export function BoardColumn({ state, items, disabled, readOnly }: BoardColumnProps) {
  const { setNodeRef, isOver } = useDroppable({
    id: state,
    disabled,
  });

  return (
    <div
      ref={setNodeRef}
      className={`flex w-80 flex-shrink-0 flex-col rounded-lg bg-gray-100 ${
        isOver && !disabled ? "ring-2 ring-blue-400" : ""
      } ${isOver && disabled ? "ring-2 ring-red-300" : ""}`}
    >
      <div className="flex items-center justify-between px-3 py-2">
        <h2 className="text-sm font-semibold text-gray-700">
          {STATE_LABELS[state]}
        </h2>
        <span className="rounded-full bg-gray-200 px-2 py-0.5 text-xs font-medium text-gray-600">
          {items.length}
        </span>
      </div>

      {items.length === 0 ? (
        <div className="flex-1 px-2 pb-2">
          <p className="py-8 text-center text-sm text-gray-400">No items</p>
        </div>
      ) : (
        <div className="flex-1 space-y-2 overflow-y-auto px-2 pb-2" role="list" aria-label={`${STATE_LABELS[state]} items`}>
          {items.map((item) => <DraggableCard key={item.id} item={item} />)}
        </div>
      )}

      {state === "NotDone" && !readOnly && <QuickAdd />}
    </div>
  );
}
