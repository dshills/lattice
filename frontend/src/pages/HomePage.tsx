import { useState } from "react";
import { Link } from "react-router";
import { useWorkItems } from "../hooks/useWorkItems";
import { CompactCard } from "../components/workitems/CompactCard";
import { CreateWorkItemForm } from "../components/forms/CreateWorkItemForm";
import { DEFAULT_PROJECT_ID, STATE_LABELS } from "../lib/constants";
import type { WorkItemState } from "../lib/types";

function SummaryCard({
  state,
  total,
}: {
  state: WorkItemState;
  total: number;
}) {
  const colors: Record<WorkItemState, string> = {
    NotDone: "border-gray-200 bg-gray-50",
    InProgress: "border-blue-200 bg-blue-50",
    Completed: "border-green-200 bg-green-50",
  };

  return (
    <div
      className={`rounded-lg border p-4 ${colors[state]}`}
    >
      <p className="text-sm font-medium text-gray-500">{STATE_LABELS[state]}</p>
      <p className="mt-1 text-2xl font-bold text-gray-900">{total}</p>
    </div>
  );
}

export function HomePage() {
  const [createOpen, setCreateOpen] = useState(false);

  const { data: notDoneData } = useWorkItems(DEFAULT_PROJECT_ID, {
    state: "NotDone",
    page_size: 1,
  });
  const { data: inProgressData } = useWorkItems(DEFAULT_PROJECT_ID, {
    state: "InProgress",
    page_size: 1,
  });
  const { data: completedData } = useWorkItems(DEFAULT_PROJECT_ID, {
    state: "Completed",
    page_size: 1,
  });
  const { data: blockedData } = useWorkItems(DEFAULT_PROJECT_ID, {
    is_blocked: true,
    page_size: 5,
  });
  const { data: inProgressItems } = useWorkItems(DEFAULT_PROJECT_ID, {
    state: "InProgress",
    page_size: 5,
  });
  const { data: recentData } = useWorkItems(DEFAULT_PROJECT_ID, { page_size: 5 });

  return (
    <div className="space-y-6">
      {/* Summary cards */}
      <div className="grid grid-cols-3 gap-4">
        <SummaryCard state="NotDone" total={notDoneData?.total ?? 0} />
        <SummaryCard state="InProgress" total={inProgressData?.total ?? 0} />
        <SummaryCard state="Completed" total={completedData?.total ?? 0} />
      </div>

      {/* Quick actions */}
      <div className="flex items-center gap-3">
        <button
          onClick={() => setCreateOpen(true)}
          className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700"
        >
          Create Work Item
        </button>
        <Link
          to="/board"
          className="rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200"
        >
          View Board
        </Link>
        <Link
          to="/list"
          className="rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200"
        >
          View List
        </Link>
        <Link
          to="/graph"
          className="rounded-md bg-gray-100 px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-200"
        >
          View Graph
        </Link>
      </div>

      {/* Sections */}
      <div className="grid gap-6 lg:grid-cols-3">
        <Section
          title="Blocked Items"
          items={blockedData?.items ?? []}
          filterLink="/list?is_blocked=true"
        />
        <Section
          title="In Progress"
          items={inProgressItems?.items ?? []}
          filterLink="/list?state=InProgress"
        />
        <Section
          title="Recently Updated"
          items={recentData?.items ?? []}
          filterLink="/list"
        />
      </div>

      <CreateWorkItemForm
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />
    </div>
  );
}

function Section({
  title,
  items,
  filterLink,
}: {
  title: string;
  items: import("../lib/types").WorkItem[];
  filterLink: string;
}) {
  return (
    <div>
      <h2 className="mb-2 text-sm font-semibold text-gray-700">{title}</h2>
      {items.length === 0 ? (
        <p className="text-sm text-gray-400">None</p>
      ) : (
        <div className="space-y-1">
          {items.map((item) => (
            <CompactCard key={item.id} item={item} />
          ))}
        </div>
      )}
      <Link
        to={filterLink}
        className="mt-2 inline-block text-xs text-blue-600 hover:underline"
      >
        View all &rarr;
      </Link>
    </div>
  );
}
