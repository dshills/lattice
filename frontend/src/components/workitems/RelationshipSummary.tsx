import { Link } from "react-router";
import type { Relationship } from "../../lib/types";
import { RELATIONSHIP_LABELS } from "../../lib/constants";

interface RelationshipSummaryProps {
  relationships: Relationship[];
}

export function RelationshipSummary({
  relationships,
}: RelationshipSummaryProps) {
  if (relationships.length === 0) {
    return <p className="text-sm text-gray-400">No relationships</p>;
  }

  return (
    <ul className="space-y-1">
      {relationships.map((rel) => (
        <li key={rel.id} className="flex items-center gap-2 text-sm">
          <span className="text-gray-500">
            {RELATIONSHIP_LABELS[rel.type]}
          </span>
          <Link
            to={`/items/${rel.target_id}`}
            className="text-blue-600 hover:underline"
          >
            {rel.target_id.slice(0, 8)}...
          </Link>
        </li>
      ))}
    </ul>
  );
}
