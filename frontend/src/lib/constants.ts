import type { WorkItemState, RelationshipType } from "./types";

export const STATE_LABELS: Record<WorkItemState, string> = {
  NotDone: "Not Done",
  InProgress: "In Progress",
  Completed: "Completed",
};

export const STATE_COLORS: Record<WorkItemState, string> = {
  NotDone: "bg-gray-100 text-gray-700",
  InProgress: "bg-blue-100 text-blue-800",
  Completed: "bg-green-100 text-green-700",
};

export const CONDITION_TAGS = ["blocked", "delayed", "needs-review"] as const;

export const CONDITION_TAG_COLORS: Record<string, string> = {
  blocked: "bg-red-100 text-red-800",
  delayed: "bg-amber-100 text-amber-800",
  "needs-review": "bg-purple-100 text-purple-800",
};

export const RELATIONSHIP_LABELS: Record<RelationshipType, string> = {
  depends_on: "depends on",
  blocks: "blocks",
  relates_to: "relates to",
  duplicate_of: "duplicate of",
};

export const STATES: WorkItemState[] = ["NotDone", "InProgress", "Completed"];
