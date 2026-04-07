import { useQuery } from "@tanstack/react-query";
import { detectCycles } from "../lib/api/cycles";

export function useCycles(workItemId: string, enabled = true) {
  return useQuery({
    queryKey: ["cycles", workItemId],
    queryFn: () => detectCycles(workItemId),
    enabled: !!workItemId && enabled,
  });
}
