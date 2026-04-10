import { useState, useCallback } from "react";
import { useParams, useNavigate } from "react-router";
import { useWorkItem } from "../hooks/useWorkItems";
import { useWorkItemMutations } from "../hooks/useWorkItemMutations";
import { useRelationships } from "../hooks/useRelationships";
import { InlineEditableText } from "../components/common/InlineEditableText";
import { LoadingState } from "../components/common/LoadingState";
import { ErrorState } from "../components/common/ErrorState";
import { StateSelector } from "../components/workitems/StateSelector";
import { AssigneeSelector } from "../components/workitems/AssigneeSelector";
import { TagEditor } from "../components/forms/TagEditor";
import { ParentChildPanel } from "../components/forms/ParentChildPanel";
import { RelationshipSummary } from "../components/workitems/RelationshipSummary";
import { RelationshipEditor } from "../components/forms/RelationshipEditor";
import { ConfirmDialog } from "../components/common/ConfirmDialog";
import { useProjectId } from "../hooks/useProjectId";
import { useProjectRole } from "../hooks/useProjectRole";
import type { RelationshipType, WorkItemState } from "../lib/types";

export function ItemDetailPage() {
  const projectId = useProjectId();
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const { data: item, isLoading, error, refetch } = useWorkItem(projectId, id!);
  const { updateMutation, deleteMutation } = useWorkItemMutations(projectId);
  const { canWrite, isOwner } = useProjectRole(projectId);
  const [showDeleteConfirm, setShowDeleteConfirm] = useState(false);
  const [cycleWarning, setCycleWarning] = useState<string | null>(null);
  const { addRelationshipMutation, removeRelationshipMutation } =
    useRelationships(projectId, id!, {
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

  const handleAssigneeChange = useCallback(
    (assigneeId: string | null) => {
      if (!id) return;
      updateMutation.mutate({
        id,
        input: { assignee_id: assigneeId ?? "" },
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
      onSuccess: () => navigate(`/projects/${projectId}/board`),
    });
  }, [id, deleteMutation, navigate, projectId]);

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
        {canWrite ? (
          <InlineEditableText
            value={item.title}
            onSave={saveField("title")}
            className="text-xl font-semibold"
            placeholder="Title"
          />
        ) : (
          <h1 className="text-xl font-semibold">{item.title}</h1>
        )}

        {canWrite ? (
          <InlineEditableText
            value={item.description}
            onSave={saveField("description")}
            as="textarea"
            placeholder="Add a description..."
          />
        ) : (
          <p className="text-sm text-gray-700 whitespace-pre-wrap">
            {item.description || "No description"}
          </p>
        )}

        <div>
          <h3 className="mb-2 text-sm font-semibold text-gray-500">
            Relationships
          </h3>
          <RelationshipSummary
            relationships={item.relationships}
            onRemove={canWrite ? handleRemoveRelationship : undefined}
          />
          {canWrite && (
            <div className="mt-3">
              <RelationshipEditor
                sourceId={item.id}
                onAdd={handleAddRelationship}
                isPending={addRelationshipMutation.isPending}
              />
            </div>
          )}
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
            disabled={!canWrite || updateMutation.isPending}
            canOverride={isOwner}
          />
        </div>

        <div>
          <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
            Assignee
          </h4>
          {canWrite ? (
            <AssigneeSelector
              value={item.assignee_id}
              onChange={handleAssigneeChange}
              disabled={updateMutation.isPending}
            />
          ) : (
            <p className="text-sm text-gray-700">
              {item.assignee_name || "Unassigned"}
            </p>
          )}
        </div>

        <div>
          <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
            Type
          </h4>
          {canWrite ? (
            <InlineEditableText
              value={item.type}
              onSave={saveField("type")}
              placeholder="Set type..."
            />
          ) : (
            <p className="text-sm text-gray-700">{item.type || "-"}</p>
          )}
        </div>

        <div>
          <h4 className="mb-2 text-xs font-semibold uppercase tracking-wide text-gray-500">
            Tags
          </h4>
          {canWrite ? (
            <TagEditor tags={item.tags} onUpdate={handleTagsUpdate} />
          ) : (
            <div className="flex flex-wrap gap-1">
              {item.tags.length > 0
                ? item.tags.map((t) => (
                    <span
                      key={t}
                      className="rounded-full bg-gray-100 px-2 py-0.5 text-xs text-gray-600"
                    >
                      {t}
                    </span>
                  ))
                : <span className="text-sm text-gray-400">None</span>}
            </div>
          )}
        </div>

        <ParentChildPanel
          item={item}
          onChangeParent={canWrite ? handleParentChange : undefined}
        />

        <div className="text-xs text-gray-400 space-y-1">
          <p>Created: {new Date(item.created_at).toLocaleString()}</p>
          <p>Updated: {new Date(item.updated_at).toLocaleString()}</p>
        </div>

        {canWrite && (
          <button
            onClick={() => setShowDeleteConfirm(true)}
            className="rounded-md px-3 py-1.5 text-sm font-medium text-red-600 hover:bg-red-50"
          >
            Delete item
          </button>
        )}
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
