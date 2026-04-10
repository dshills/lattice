import { useNavigate } from "react-router";
import type { WorkItem } from "../../lib/types";
import { TagBadge } from "./TagBadge";
import { TypeBadge } from "./TypeBadge";
import { UserInitials } from "./UserInitials";
import { CONDITION_TAGS } from "../../lib/constants";

function timeAgo(dateStr: string): string {
  const seconds = Math.floor(
    (Date.now() - new Date(dateStr).getTime()) / 1000,
  );
  if (seconds < 60) return "just now";
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}h ago`;
  const days = Math.floor(hours / 24);
  return `${days}d ago`;
}

const MAX_TAGS = 3;

export function WorkItemCard({ item }: { item: WorkItem }) {
  const navigate = useNavigate();

  const conditionTags = item.tags.filter((t) =>
    (CONDITION_TAGS as readonly string[]).includes(t),
  );
  const regularTags = item.tags.filter(
    (t) => !(CONDITION_TAGS as readonly string[]).includes(t),
  );
  const visibleTags = [...conditionTags, ...regularTags].slice(0, MAX_TAGS);
  const overflowCount = item.tags.length - visibleTags.length;

  return (
    <div
      className="cursor-pointer rounded-lg border border-gray-200 bg-white p-3 shadow-sm transition-shadow hover:shadow-md"
      onClick={() => navigate(`/items/${item.id}`)}
      role="button"
      tabIndex={0}
      onKeyDown={(e) => {
        if (e.key === "Enter") navigate(`/items/${item.id}`);
      }}
    >
      <div className="mb-1 flex items-start justify-between gap-2">
        <h3 className="text-sm font-medium text-gray-900 leading-snug">
          {item.title}
        </h3>
        {item.is_blocked && (
          <span
            className="flex-shrink-0 text-red-500"
            title="Blocked by dependency"
          >
            &#9888;
          </span>
        )}
      </div>

      <div className="mt-2 flex flex-wrap items-center gap-1">
        {item.type && <TypeBadge type={item.type} />}
        {visibleTags.map((tag) => (
          <TagBadge key={tag} tag={tag} />
        ))}
        {overflowCount > 0 && (
          <span className="text-xs text-gray-400">+{overflowCount}</span>
        )}
      </div>

      <div className="mt-2 flex items-center justify-between text-xs text-gray-400">
        <span>{timeAgo(item.updated_at)}</span>
        <div className="flex items-center gap-1">
          {item.parent_id && <span title="Has parent">&#x1F517;</span>}
          {item.assignee_name && (
            <UserInitials name={item.assignee_name} />
          )}
        </div>
      </div>
    </div>
  );
}
