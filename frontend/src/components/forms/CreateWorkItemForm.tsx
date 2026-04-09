import { useState } from "react";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import {
  createWorkItemSchema,
  type CreateWorkItemData,
} from "../../lib/validation";
import { useWorkItemMutations } from "../../hooks/useWorkItemMutations";
import { useProjectId } from "../../hooks/useProjectId";
import { Modal } from "../common/Modal";
import { TagEditor } from "./TagEditor";

interface CreateWorkItemFormProps {
  open: boolean;
  onClose: () => void;
  defaultParentId?: string;
  onCreated?: (id: string) => void;
}

export function CreateWorkItemForm({
  open,
  onClose,
  defaultParentId,
  onCreated,
}: CreateWorkItemFormProps) {
  const [showDetails, setShowDetails] = useState(false);
  const [tags, setTags] = useState<string[]>([]);
  const projectId = useProjectId();
  const { createMutation } = useWorkItemMutations(projectId);

  const {
    register,
    handleSubmit,
    reset,
    formState: { errors },
  } = useForm<CreateWorkItemData>({
    resolver: zodResolver(createWorkItemSchema),
    defaultValues: {
      parent_id: defaultParentId,
    },
  });

  const onSubmit = (data: CreateWorkItemData) => {
    createMutation.mutate(
      {
        ...data,
        tags: tags.length > 0 ? tags : undefined,
      },
      {
        onSuccess: (item) => {
          reset();
          setTags([]);
          setShowDetails(false);
          onClose();
          onCreated?.(item.id);
        },
      },
    );
  };

  return (
    <Modal open={open} onClose={onClose} title="Create Work Item">
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
        <div>
          <input
            {...register("title")}
            type="text"
            placeholder="Title"
            className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            autoFocus
          />
          {errors.title && (
            <p className="mt-1 text-xs text-red-500">{errors.title.message}</p>
          )}
        </div>

        {!showDetails && (
          <button
            type="button"
            onClick={() => setShowDetails(true)}
            className="text-sm text-blue-600 hover:underline"
          >
            Add details
          </button>
        )}

        {showDetails && (
          <>
            <div>
              <textarea
                {...register("description")}
                placeholder="Description"
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 min-h-[80px]"
              />
              {errors.description && (
                <p className="mt-1 text-xs text-red-500">
                  {errors.description.message}
                </p>
              )}
            </div>
            <div>
              <input
                {...register("type")}
                type="text"
                placeholder="Type (e.g., bug, feature, task)"
                className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="mb-1 block text-xs font-medium text-gray-500">
                Tags
              </label>
              <TagEditor tags={tags} onUpdate={setTags} />
            </div>
          </>
        )}

        <div className="flex justify-end gap-3 pt-2">
          <button
            type="button"
            onClick={onClose}
            className="rounded-md px-4 py-2 text-sm font-medium text-gray-700 hover:bg-gray-100"
          >
            Cancel
          </button>
          <button
            type="submit"
            disabled={createMutation.isPending}
            className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {createMutation.isPending ? "Creating..." : "Create"}
          </button>
        </div>

        {createMutation.isError && (
          <p className="text-xs text-red-500">
            Failed to create item. Try again.
          </p>
        )}
      </form>
    </Modal>
  );
}
