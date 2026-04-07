import type { WorkItemState } from "../../lib/types";
import { STATE_LABELS, STATE_COLORS } from "../../lib/constants";

export function StateBadge({ state }: { state: WorkItemState }) {
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full px-2 py-0.5 text-xs font-medium ${STATE_COLORS[state]}`}
    >
      {STATE_LABELS[state]}
    </span>
  );
}
