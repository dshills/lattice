import { useMutation, useQueryClient } from "@tanstack/react-query";
import {
  addRelationship,
  removeRelationship,
} from "../lib/api/relationships";
import { detectCycles, type CycleResult } from "../lib/api/cycles";
import type { AddRelationshipInput } from "../lib/types";

interface UseRelationshipsOptions {
  onCycleDetected?: (result: CycleResult) => void;
}

export function useRelationships(
  sourceId: string,
  options?: UseRelationshipsOptions,
) {
  const queryClient = useQueryClient();

  const invalidateSource = () => {
    queryClient.invalidateQueries({ queryKey: ["workitem", sourceId] });
    queryClient.invalidateQueries({ queryKey: ["workitems"] });
  };

  const addRelationshipMutation = useMutation({
    mutationFn: (input: AddRelationshipInput) =>
      addRelationship(sourceId, input),
    onSuccess: async (_data, variables) => {
      invalidateSource();

      if (
        variables.type === "depends_on" ||
        variables.type === "blocks"
      ) {
        try {
          const result = await detectCycles(sourceId);
          if (result.has_cycle) {
            options?.onCycleDetected?.(result);
          }
        } catch {
          // Cycle detection failure is non-blocking
        }
      }
    },
  });

  const removeRelationshipMutation = useMutation({
    mutationFn: (relationshipId: string) =>
      removeRelationship(sourceId, relationshipId),
    onSuccess: () => invalidateSource(),
  });

  return { addRelationshipMutation, removeRelationshipMutation };
}
