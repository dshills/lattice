import { useQuery } from "@tanstack/react-query";
import { detectCycles } from "../lib/api/cycles";

export function useCycles(
  projectId: string,
  workItemId: string,
  enabled = true,
) {
  return useQuery({
    queryKey: ["cycles", projectId, workItemId],
    queryFn: () => detectCycles(projectId, workItemId),
    enabled: !!projectId && !!workItemId && enabled,
  });
}
