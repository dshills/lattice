import { useState } from "react";
import { useWorkItemMutations } from "../../hooks/useWorkItemMutations";

export function QuickAdd() {
  const [title, setTitle] = useState("");
  const { createMutation } = useWorkItemMutations();

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    const trimmed = title.trim();
    if (!trimmed) return;

    createMutation.mutate({ title: trimmed });
    setTitle("");
  };

  return (
    <div className="border-t border-gray-200 px-2 py-2">
      <form onSubmit={handleSubmit}>
        <input
          type="text"
          value={title}
          onChange={(e) => setTitle(e.target.value)}
          placeholder="Add item..."
          className="w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm placeholder-gray-400 focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
          disabled={createMutation.isPending}
        />
      </form>
      {createMutation.isError && (
        <p className="mt-1 text-xs text-red-500">
          Failed to create item. Try again.
        </p>
      )}
    </div>
  );
}
