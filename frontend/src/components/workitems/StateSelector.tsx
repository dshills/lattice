import type { WorkItemState } from "../../lib/types";
import { STATES, STATE_LABELS, STATE_COLORS } from "../../lib/constants";
import { isAdmin } from "../../lib/config";

const FORWARD: Record<WorkItemState, WorkItemState[]> = {
  NotDone: ["InProgress"],
  InProgress: ["Completed"],
  Completed: [],
};

const BACKWARD: Record<WorkItemState, WorkItemState[]> = {
  NotDone: [],
  InProgress: ["NotDone"],
  Completed: ["InProgress"],
};

interface StateSelectorProps {
  current: WorkItemState;
  onChange: (state: WorkItemState, override: boolean) => void;
  disabled?: boolean;
}

export function StateSelector({
  current,
  onChange,
  disabled,
}: StateSelectorProps) {
  const admin = isAdmin();
  const forward = FORWARD[current];
  const backward = admin ? BACKWARD[current] : [];

  return (
    <div className="flex flex-wrap gap-1">
      {STATES.map((state) => {
        const isCurrent = state === current;
        const isForward = forward.includes(state);
        const isBackward = backward.includes(state);
        const canSelect = isForward || isBackward;

        return (
          <button
            key={state}
            disabled={disabled || isCurrent || !canSelect}
            onClick={() => onChange(state, isBackward)}
            className={`rounded-full px-3 py-1 text-xs font-medium transition-colors ${
              isCurrent
                ? `${STATE_COLORS[state]} ring-2 ring-offset-1 ring-current`
                : canSelect
                  ? `${STATE_COLORS[state]} cursor-pointer opacity-70 hover:opacity-100`
                  : "bg-gray-50 text-gray-300 cursor-not-allowed"
            }`}
          >
            {STATE_LABELS[state]}
            {isBackward && " (Override)"}
          </button>
        );
      })}
    </div>
  );
}
