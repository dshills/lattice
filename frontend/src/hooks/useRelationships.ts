import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  addRelationship,
  removeRelationship,
} from "../lib/api/relationships";
import { detectCycles, type CycleResult } from "../lib/api/cycles";
import { useToast } from "../components/common/Toast";
import { toastError } from "../lib/toastError";
import type { AddRelationshipInput } from "../lib/types";

interface UseRelationshipsOptions {
  onCycleDetected?: (result: CycleResult) => void;
}

export function useRelationships(
  projectId: string,
  sourceId: string,
  options?: UseRelationshipsOptions,
) {
  const queryClient = useQueryClient();
  const { addToast } = useToast();

  const invalidateSource = () => {
    queryClient.invalidateQueries({ queryKey: ["workitem", projectId, sourceId] });
    queryClient.invalidateQueries({ queryKey: ["workitems", projectId] });
  };

  const addRelationshipMutation = useMutation({
    mutationFn: (input: AddRelationshipInput) =>
      addRelationship(projectId, sourceId, input),
    onSuccess: async (_data, variables) => {
      invalidateSource();
      addToast("Relationship added", "success");

      if (
        variables.type === "depends_on" ||
        variables.type === "blocks"
      ) {
        try {
          const result = await detectCycles(projectId, sourceId);
          if (result.has_cycle) {
            options?.onCycleDetected?.(result);
          }
        } catch {
          // Cycle detection failure is non-blocking
        }
      }
    },
    onError: (err) => toastError(addToast, err),
  });

  const removeRelationshipMutation = useMutation({
    mutationFn: (relationshipId: string) =>
      removeRelationship(projectId, sourceId, relationshipId),
    onSuccess: () => {
      invalidateSource();
      addToast("Relationship removed", "success");
    },
    onError: (err) => toastError(addToast, err),
  });

  return { addRelationshipMutation, removeRelationshipMutation };
}
