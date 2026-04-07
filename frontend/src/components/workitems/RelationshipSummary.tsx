import { Link } from "react-router";
import type { Relationship } from "../../lib/types";
import { RELATIONSHIP_LABELS } from "../../lib/constants";

interface RelationshipListProps {
  relationships: Relationship[];
  onRemove?: (relationshipId: string) => void;
  compact?: boolean;
}

export function RelationshipSummary({
  relationships,
  onRemove,
  compact,
}: RelationshipListProps) {
  if (relationships.length === 0) {
    return <p className="text-sm text-gray-400">No relationships</p>;
  }

  if (compact) {
    // Compact inline summary for list/card views
    const counts: Record<string, number> = {};
    for (const rel of relationships) {
      const label = RELATIONSHIP_LABELS[rel.type] ?? rel.type;
      counts[label] = (counts[label] ?? 0) + 1;
    }
    return (
      <span className="text-xs text-gray-500">
        {Object.entries(counts)
          .map(([label, count]) => `${count} ${label}`)
          .join(", ")}
      </span>
    );
  }

  // Full grouped list for detail view
  const grouped = new Map<string, Relationship[]>();
  for (const rel of relationships) {
    const label = RELATIONSHIP_LABELS[rel.type] ?? rel.type;
    const group = grouped.get(label) ?? [];
    group.push(rel);
    grouped.set(label, group);
  }

  return (
    <div className="space-y-3">
      {Array.from(grouped.entries()).map(([label, rels]) => (
        <div key={label}>
          <h4 className="mb-1 text-xs font-semibold uppercase tracking-wide text-gray-500">
            {label}
          </h4>
          <ul className="space-y-1">
            {rels.map((rel) => (
              <li
                key={rel.id}
                className="flex items-center justify-between text-sm"
              >
                <Link
                  to={`/items/${rel.target_id}`}
                  className="text-blue-600 hover:underline"
                >
                  {rel.target_id.slice(0, 8)}...
                </Link>
                {onRemove && (
                  <button
                    onClick={() => onRemove(rel.id)}
                    className="text-xs text-gray-400 hover:text-red-500"
                    aria-label={`Remove relationship ${rel.id}`}
                  >
                    Remove
                  </button>
                )}
              </li>
            ))}
          </ul>
        </div>
      ))}
    </div>
  );
}
