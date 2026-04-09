import { useQuery } from "@tanstack/react-query";
import { listWorkItems, getWorkItem } from "../lib/api/workitems";
import type { ListFilter } from "../lib/types";

export function useWorkItems(projectId: string, filter: ListFilter = {}) {
  return useQuery({
    queryKey: ["workitems", projectId, filter],
    queryFn: () => listWorkItems(projectId, filter),
    enabled: !!projectId,
  });
}

export function useWorkItem(projectId: string, id: string) {
  return useQuery({
    queryKey: ["workitem", projectId, id],
    queryFn: () => getWorkItem(projectId, id),
    enabled: !!projectId && !!id,
  });
}
