import { useState } from "react";
import type { RelationshipType } from "../../lib/types";
import { RELATIONSHIP_LABELS } from "../../lib/constants";

const RELATIONSHIP_TYPES: RelationshipType[] = [
  "depends_on",
  "blocks",
  "relates_to",
  "duplicate_of",
];

interface RelationshipEditorProps {
  sourceId: string;
  onAdd: (type: RelationshipType, targetId: string) => void;
  isPending?: boolean;
}

export function RelationshipEditor({
  sourceId,
  onAdd,
  isPending,
}: RelationshipEditorProps) {
  const [open, setOpen] = useState(false);
  const [type, setType] = useState<RelationshipType>("depends_on");
  const [targetId, setTargetId] = useState("");
  const [error, setError] = useState("");

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = targetId.trim();
    if (!trimmed) {
      setError("Target ID is required");
      return;
    }
    if (trimmed === sourceId) {
      setError("Cannot create a relationship to itself");
      return;
    }
    setError("");
    onAdd(type, trimmed);
  };

  if (!open) {
    return (
      <button
        onClick={() => setOpen(true)}
        className="text-sm text-blue-600 hover:underline"
      >
        Add relationship
      </button>
    );
  }

  return (
    <form onSubmit={handleSubmit} className="space-y-3 rounded-lg border border-gray-200 p-3">
      <div>
        <label className="mb-1 block text-xs font-medium text-gray-500">
          Type
        </label>
        <select
          value={type}
          onChange={(e) => setType(e.target.value as RelationshipType)}
          className="w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        >
          {RELATIONSHIP_TYPES.map((t) => (
            <option key={t} value={t}>
              {RELATIONSHIP_LABELS[t]}
            </option>
          ))}
        </select>
      </div>

      <div>
        <label className="mb-1 block text-xs font-medium text-gray-500">
          Target Item ID
        </label>
        <input
          type="text"
          value={targetId}
          onChange={(e) => setTargetId(e.target.value)}
          placeholder="Enter work item ID..."
          className="w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
        />
        {error && <p className="mt-1 text-xs text-red-500">{error}</p>}
      </div>

      {targetId.trim() && targetId.trim() !== sourceId && (
        <p className="text-xs text-gray-500">
          This item {RELATIONSHIP_LABELS[type]}: {targetId.trim().slice(0, 8)}
          ...
        </p>
      )}

      <div className="flex gap-2">
        <button
          type="submit"
          disabled={isPending}
          className="rounded-md bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
        >
          {isPending ? "Adding..." : "Add"}
        </button>
        <button
          type="button"
          onClick={() => {
            setOpen(false);
            setError("");
          }}
          className="rounded-md px-3 py-1.5 text-sm font-medium text-gray-600 hover:bg-gray-100"
        >
          Cancel
        </button>
      </div>
    </form>
  );
}
