import { Link } from "react-router";
import { useProjectId } from "../../hooks/useProjectId";
import type { WorkItem } from "../../lib/types";
import { StateBadge } from "./StateBadge";
import { TypeBadge } from "./TypeBadge";

export function CompactCard({ item }: { item: WorkItem }) {
  const projectId = useProjectId();
  return (
    <Link
      to={`/projects/${projectId}/items/${item.id}`}
      className="flex items-center gap-3 rounded-md border border-gray-100 px-3 py-2 transition-colors hover:bg-gray-50"
    >
      <div className="min-w-0 flex-1">
        <p className="truncate text-sm font-medium text-gray-900">
          {item.title}
        </p>
      </div>
      <div className="flex items-center gap-1.5">
        <StateBadge state={item.state} />
        {item.type && <TypeBadge type={item.type} />}
        {item.is_blocked && (
          <span className="text-xs text-red-500" title="Blocked">
            &#9888;
          </span>
        )}
      </div>
    </Link>
  );
}
