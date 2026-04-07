import { useDraggable } from "@dnd-kit/core";
import type { WorkItem } from "../../lib/types";
import { WorkItemCard } from "./WorkItemCard";

export function DraggableCard({ item }: { item: WorkItem }) {
  const { attributes, listeners, setNodeRef, transform, isDragging } =
    useDraggable({
      id: item.id,
      data: { item },
    });

  const style = transform
    ? {
        transform: `translate(${transform.x}px, ${transform.y}px)`,
      }
    : undefined;

  return (
    <div
      ref={setNodeRef}
      style={style}
      className={isDragging ? "opacity-50" : ""}
      {...listeners}
      {...attributes}
    >
      <WorkItemCard item={item} />
    </div>
  );
}
