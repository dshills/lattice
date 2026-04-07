import {
  CONDITION_TAGS,
  CONDITION_TAG_COLORS,
} from "../../lib/constants";

export function TagBadge({ tag }: { tag: string }) {
  const isCondition = (CONDITION_TAGS as readonly string[]).includes(tag);
  const colors = isCondition
    ? CONDITION_TAG_COLORS[tag]
    : "bg-gray-100 text-gray-600";

  return (
    <span
      className={`inline-flex items-center rounded-full px-2 py-0.5 text-xs font-medium ${colors}`}
    >
      {tag}
    </span>
  );
}
