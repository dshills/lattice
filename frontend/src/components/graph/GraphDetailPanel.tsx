import { Link } from "react-router";
import type { WorkItem } from "../../lib/types";
import { StateBadge } from "../workitems/StateBadge";
import { TypeBadge } from "../workitems/TypeBadge";
import { TagBadge } from "../workitems/TagBadge";
import { RELATIONSHIP_LABELS } from "../../lib/constants";

interface GraphDetailPanelProps {
  item: WorkItem;
  onFocus: (id: string) => void;
}

export function GraphDetailPanel({ item, onFocus }: GraphDetailPanelProps) {
  return (
    <div className="w-72 border-l border-gray-200 bg-white p-4 overflow-y-auto">
      <h3 className="text-sm font-semibold text-gray-900">{item.title}</h3>

      <div className="mt-2 flex flex-wrap gap-1">
        <StateBadge state={item.state} />
        {item.type && <TypeBadge type={item.type} />}
      </div>

      <div className="mt-3 flex flex-wrap gap-1">
        {item.tags.map((tag) => (
          <TagBadge key={tag} tag={tag} />
        ))}
      </div>

      {item.description && (
        <p className="mt-3 text-xs text-gray-500 line-clamp-4">
          {item.description}
        </p>
      )}

      {item.relationships.length > 0 && (
        <div className="mt-3">
          <h4 className="text-xs font-semibold text-gray-500">Relationships</h4>
          <ul className="mt-1 space-y-1">
            {item.relationships.map((rel) => (
              <li key={rel.id} className="text-xs text-gray-500">
                {RELATIONSHIP_LABELS[rel.type] ?? rel.type}{" "}
                {rel.target_id.slice(0, 8)}...
              </li>
            ))}
          </ul>
        </div>
      )}

      <div className="mt-4 flex flex-col gap-2">
        <Link
          to={`/items/${item.id}`}
          className="text-xs text-blue-600 hover:underline"
        >
          Open full detail
        </Link>
        <button
          onClick={() => onFocus(item.id)}
          className="text-left text-xs text-blue-600 hover:underline"
        >
          Focus here
        </button>
      </div>
    </div>
  );
}
