import { Link } from "react-router";
import { useWorkItems } from "../../hooks/useWorkItems";
import type { WorkItem } from "../../lib/types";

interface ParentChildPanelProps {
  item: WorkItem;
  onChangeParent: (parentId: string | null) => void;
}

export function ParentChildPanel({
  item,
  onChangeParent,
}: ParentChildPanelProps) {
  const { data: childrenData } = useWorkItems(item.project_id, {
    parent_id: item.id,
    page_size: 10,
  });

  return (
    <div className="space-y-4">
      <div>
        <h4 className="mb-1 text-xs font-semibold uppercase tracking-wide text-gray-500">
          Parent
        </h4>
        {item.parent_id ? (
          <div className="flex items-center gap-2">
            <Link
              to={`/items/${item.parent_id}`}
              className="text-sm text-blue-600 hover:underline"
            >
              {item.parent_id}
            </Link>
            <button
              onClick={() => onChangeParent(null)}
              className="text-xs text-gray-400 hover:text-red-500"
            >
              Remove
            </button>
          </div>
        ) : (
          <p className="text-sm text-gray-400">None</p>
        )}
      </div>

      <div>
        <h4 className="mb-1 text-xs font-semibold uppercase tracking-wide text-gray-500">
          Children ({childrenData?.total ?? 0})
        </h4>
        {childrenData?.items.length ? (
          <ul className="space-y-1">
            {childrenData.items.map((child) => (
              <li key={child.id}>
                <Link
                  to={`/items/${child.id}`}
                  className="text-sm text-blue-600 hover:underline"
                >
                  {child.title}
                </Link>
              </li>
            ))}
          </ul>
        ) : (
          <p className="text-sm text-gray-400">No children</p>
        )}
      </div>
    </div>
  );
}
