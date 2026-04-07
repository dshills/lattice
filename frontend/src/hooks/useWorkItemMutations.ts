import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  createWorkItem,
  updateWorkItem,
  deleteWorkItem,
} from "../lib/api/workitems";
import type { CreateWorkItemInput, UpdateWorkItemInput } from "../lib/types";

export function useWorkItemMutations() {
  const queryClient = useQueryClient();

  const invalidate = () => {
    queryClient.invalidateQueries({ queryKey: ["workitems"] });
  };

  const createMutation = useMutation({
    mutationFn: (input: CreateWorkItemInput) => createWorkItem(input),
    onSuccess: () => invalidate(),
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
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => deleteWorkItem(id),
    onSuccess: () => invalidate(),
  });

  return { createMutation, updateMutation, deleteMutation };
}
