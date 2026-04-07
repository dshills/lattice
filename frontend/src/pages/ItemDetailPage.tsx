import { useState, useCallback } from "react";
import { useParams, useNavigate } from "react-router";
import { useWorkItem } from "../hooks/useWorkItems";
import { useWorkItemMutations } from "../hooks/useWorkItemMutations";
import { useRelationships } from "../hooks/useRelationships";
import { InlineEditableText } from "../components/common/InlineEditableText";
import { LoadingState } from "../components/common/LoadingState";
import { ErrorState } from "../components/common/ErrorState";
import { StateSelector } from "../components/workitems/StateSelector";
import { TagEditor } from "../components/forms/TagEditor";
import { ParentChildPanel } from "../components/forms/ParentChildPanel";
import { RelationshipSummary } from "../components/workitems/RelationshipSummary";
import { RelationshipEditor } from "../components/forms/RelationshipEditor";
import { ConfirmDialog } from "../components/common/ConfirmDialog";
import type { RelationshipType, WorkItemState } from "../lib/types";

export function ItemDetailPage() {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: item, isLoading, error, refetch } = useWorkItem(id!);
  const { updateMutation, deleteMutation } = useWorkItemMutations();
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [cycleWarning, setCycleWarning] = useState<string | null>(null);
  const { addRelationshipMutation, removeRelationshipMutation } =
    useRelationships(id!, {
      onCycleDetected: (result) => {
        const msg = `Dependency cycle detected: ${result.cycle.map((c) => c.slice(0, 8)).join(" -> ")}`;
        setCycleWarning(msg);
        setTimeout(() => setCycleWarning(null), 8000);
      },
    });

  const saveField = useCallback(
    (field: string) => async (value: string) => {
      if (!id) return;
      await updateMutation.mutateAsync({
        id,
        input: { [field]: value },
      });
    },
    [id, updateMutation],
  );

  const handleStateChange = useCallback(
    (state: WorkItemState, override: boolean) => {
      if (!id) return;
      updateMutation.mutate({
        id,
        input: { state, ...(override ? { override: true } : {}) },
      });
    },
    [id, updateMutation],
  );

  const handleTagsUpdate = useCallback(
    (tags: string[]) => {
      if (!id) return;
      updateMutation.mutate({ id, input: { tags } });
    },
    [id, updateMutation],
  );

  const handleParentChange = useCallback(
    (parentId: string | null) => {
      if (!id) return;
      updateMutation.mutate({
        id,
        input: { parent_id: parentId ?? "" },
      });
    },
    [id, updateMutation],
  );

  const handleAddRelationship = useCallback(
    (type: RelationshipType, targetId: string) => {
      addRelationshipMutation.mutate({ type, target_id: targetId });
    },
    [addRelationshipMutation],
  );

  const handleRemoveRelationship = useCallback(
    (relationshipId: string) => {
      removeRelationshipMutation.mutate(relationshipId);
    },
    [removeRelationshipMutation],
  );

  const handleDelete = useCallback(() => {
    if (!id) return;
    deleteMutation.mutate(id, {
      onSuccess: () => navigate("/board"),
    });
  }, [id, deleteMutation, navigate]);

  if (isLoading) {
    return <LoadingState />;
  }

  if (error || !item) {
    return <ErrorState message={error?.message ?? "Item not found"} onRetry={() => refetch()} />;
  }

  return (
    <div className="mx-auto grid max-w-5xl gap-8 lg:grid-cols-[1fr_300px]">
      {/* Cycle warning toast */}
      {cycleWarning && (
        <div className="col-span-full rounded-md bg-amber-50 border border-amber-200 p-3 text-sm text-amber-800">
          {cycleWarning}
        </div>
      )}

      {/* Main content */}
      <div className="space-y-6">
        <InlineEditableText
          value={item.title}
          onSave={saveField("title")}
          className="text-xl font-semibold"
          placeholder="Title"
        />

        <InlineEditableText
          value={item.description}
          onSave={saveField("description")}
          as="textarea"
          placeholder="Add a description..."
        />

        <div>
          <h3 className="mb-2 text-sm font-semibold text-gray-500">
            Relationships
          </h3>
          <RelationshipSummary
            relationships={item.relationships}
            onRemove={handleRemoveRelationship}
          />
          <div className="mt-3">
            <RelationshipEditor
              sourceId={item.id}
              onAdd={handleAddRelationship}
              isPending={addRelationshipMutation.isPending}
            />
          </div>
        </div>
      </div>

      {/* Sidebar */}
      <div className="space-y-6">
        <div>
          <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
            State
          </h4>
          <StateSelector
            current={item.state}
            onChange={handleStateChange}
            disabled={updateMutation.isPending}
          />
        </div>

        <div>
          <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
            Type
          </h4>
          <InlineEditableText
            value={item.type}
            onSave={saveField("type")}
            placeholder="Set type..."
          />
        </div>

        <div>
          <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
            Tags
          </h4>
          <TagEditor tags={item.tags} onUpdate={handleTagsUpdate} />
        </div>

        <ParentChildPanel item={item} onChangeParent={handleParentChange} />

        <div className="text-xs text-gray-400 space-y-1">
          <p>Created: {new Date(item.created_at).toLocaleString()}</p>
          <p>Updated: {new Date(item.updated_at).toLocaleString()}</p>
        </div>

        <button
          onClick={() => setShowDeleteConfirm(true)}
          className="rounded-md px-3 py-1.5 text-sm font-medium text-red-600 hover:bg-red-50"
        >
          Delete item
        </button>
      </div>

      <ConfirmDialog
        open={showDeleteConfirm}
        onClose={() => setShowDeleteConfirm(false)}
        onConfirm={handleDelete}
        title="Delete work item"
        message="Are you sure you want to delete this item? This action cannot be undone."
        confirmLabel="Delete"
        danger
      />
    </div>
  );
}
