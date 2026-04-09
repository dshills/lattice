import { apiFetch } from "./client";

export interface CycleResult {
  has_cycle: boolean;
  cycle: string[];
}

export function detectCycles(
  projectId: string,
  workItemId: string,
): Promise<CycleResult> {
  return apiFetch<CycleResult>(
    `/projects/${projectId}/workitems/${workItemId}/cycles`,
  );
}
