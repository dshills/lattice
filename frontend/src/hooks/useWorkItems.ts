import { useQuery } from "@tanstack/react-query";
import { listWorkItems, getWorkItem } from "../lib/api/workitems";
import type { ListFilter } from "../lib/types";

export function useWorkItems(filter: ListFilter = {}) {
  return useQuery({
    queryKey: ["workitems", filter],
    queryFn: () => listWorkItems(filter),
  });
}

export function useWorkItem(id: string) {
  return useQuery({
    queryKey: ["workitem", id],
    queryFn: () => getWorkItem(id),
    enabled: !!id,
  });
}
