import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createWorkItem,
  updateWorkItem,
  deleteWorkItem,
} from "../lib/api/workitems";
import { useToast } from "../components/common/Toast";
import { toastError } from "../lib/toastError";
import type { CreateWorkItemInput, UpdateWorkItemInput } from "../lib/types";

export function useWorkItemMutations() {
  const queryClient = useQueryClient();
  const { addToast } = useToast();

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["workitems"] });
  };

  const createMutation = useMutation({
    mutationFn: (input: CreateWorkItemInput) => createWorkItem(input),
    onSuccess: () => {
      invalidate();
      addToast("Work item created", "success");
    },
    onError: (err) => toastError(addToast, err),
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, input }: { id: string; input: UpdateWorkItemInput }) =>
      updateWorkItem(id, input),
    onSuccess: (_data, variables) => {
      queryClient.invalidateQueries({
        queryKey: ["workitem", variables.id],
      });
      invalidate();
    },
    onError: (err) => toastError(addToast, err),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteWorkItem(id),
    onSuccess: () => {
      invalidate();
      addToast("Work item deleted", "success");
    },
    onError: (err) => toastError(addToast, err),
  });

  return { createMutation, updateMutation, deleteMutation };
}
