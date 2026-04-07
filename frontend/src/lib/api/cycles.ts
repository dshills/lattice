import { apiFetch } from "./client";

export interface CycleResult {
  has_cycle: boolean;
  cycle: string[];
}

export function detectCycles(workItemId: string): Promise<CycleResult> {
  return apiFetch<CycleResult>(`/workitems/${workItemId}/cycles`);
}
