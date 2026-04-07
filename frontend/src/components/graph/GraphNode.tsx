import { Handle, Position, type NodeProps } from "@xyflow/react";
import type { WorkItem } from "../../lib/types";
import { STATE_COLORS, STATE_LABELS } from "../../lib/constants";

type GraphNodeData = {
  item: WorkItem;
  selected: boolean;
};

export function GraphNode({ data }: NodeProps) {
  const { item, selected } = data as GraphNodeData;
  const title =
    item.title.length > 30 ? item.title.slice(0, 30) + "..." : item.title;

  return (
    <div
      className={`rounded-lg border bg-white px-3 py-2 shadow-sm transition-shadow ${
        selected
          ? "border-blue-500 ring-2 ring-blue-200 shadow-md"
          : "border-gray-200"
      }`}
      style={{ minWidth: 150 }}
    >
      <Handle type="target" position={Position.Top} className="!bg-gray-400" />
      <div className="flex items-center gap-2">
        <span
          className={`inline-flex rounded-full px-1.5 py-0.5 text-[10px] font-medium ${STATE_COLORS[item.state]}`}
        >
          {STATE_LABELS[item.state]}
        </span>
        {item.is_blocked && (
          <span className="text-xs text-red-500" title="Blocked">
            &#9888;
          </span>
        )}
      </div>
      <p className="mt-1 text-xs font-medium text-gray-800">{title}</p>
      {item.type && (
        <p className="mt-0.5 text-[10px] text-gray-400">{item.type}</p>
      )}
      <Handle
        type="source"
        position={Position.Bottom}
        className="!bg-gray-400"
      />
    </div>
  );
}
