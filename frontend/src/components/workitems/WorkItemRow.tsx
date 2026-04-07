import { useState } from "react";
import { Link } from "react-router";
import type { WorkItem } from "../../lib/types";
import { StateBadge } from "./StateBadge";
import { TypeBadge } from "./TypeBadge";
import { TagBadge } from "./TagBadge";
import { RELATIONSHIP_LABELS } from "../../lib/constants";

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

function relationshipSummary(item: WorkItem): string {
  if (item.relationships.length === 0) return "";
  const counts: Record<string, number> = {};
  for (const rel of item.relationships) {
    const label = RELATIONSHIP_LABELS[rel.type] ?? rel.type;
    counts[label] = (counts[label] ?? 0) + 1;
  }
  return Object.entries(counts)
    .map(([label, count]) => `${count} ${label}`)
    .join(", ");
}

const MAX_TAGS = 3;

export function WorkItemRow({ item }: { item: WorkItem }) {
  const [expanded, setExpanded] = useState(false);

  const visibleTags = item.tags.slice(0, MAX_TAGS);
  const overflow = item.tags.length - MAX_TAGS;

  return (
    <>
      <tr
        className="cursor-pointer border-b border-gray-100 hover:bg-gray-50"
        onClick={() => setExpanded(!expanded)}
      >
        <td className="px-3 py-2 text-sm">
          <Link
            to={`/items/${item.id}`}
            onClick={(e) => e.stopPropagation()}
            className="font-medium text-blue-600 hover:underline"
          >
            {item.title}
          </Link>
          {item.is_blocked && (
            <span className="ml-1 text-red-500" title="Blocked">
              &#9888;
            </span>
          )}
        </td>
        <td className="px-3 py-2">
          <StateBadge state={item.state} />
        </td>
        <td className="px-3 py-2">
          {item.type && <TypeBadge type={item.type} />}
        </td>
        <td className="px-3 py-2">
          <div className="flex flex-wrap gap-1">
            {visibleTags.map((tag) => (
              <TagBadge key={tag} tag={tag} />
            ))}
            {overflow > 0 && (
              <span className="text-xs text-gray-400">+{overflow}</span>
            )}
          </div>
        </td>
        <td className="px-3 py-2 text-xs text-gray-500">
          {relationshipSummary(item)}
        </td>
        <td className="px-3 py-2 text-xs text-gray-400">
          {timeAgo(item.updated_at)}
        </td>
      </tr>
      {expanded && (
        <tr className="border-b border-gray-100 bg-gray-50">
          <td colSpan={6} className="px-6 py-3">
            <p className="mb-2 text-sm text-gray-600">
              {item.description || "No description"}
            </p>
            {item.relationships.length > 0 && (
              <div className="space-y-1">
                {item.relationships.map((rel) => (
                  <div key={rel.id} className="text-xs text-gray-500">
                    {RELATIONSHIP_LABELS[rel.type]}{" "}
                    <Link
                      to={`/items/${rel.target_id}`}
                      className="text-blue-600 hover:underline"
                    >
                      {rel.target_id.slice(0, 8)}...
                    </Link>
                  </div>
                ))}
              </div>
            )}
          </td>
        </tr>
      )}
    </>
  );
}
