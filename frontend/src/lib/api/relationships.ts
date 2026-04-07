import type { Relationship, AddRelationshipInput } from "../types";
import { apiFetch } from "./client";

export function addRelationship(
  workItemId: string,
  input: AddRelationshipInput,
): Promise<Relationship> {
  return apiFetch<Relationship>(`/workitems/${workItemId}/relationships`, {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function removeRelationship(
  workItemId: string,
  relationshipId: string,
): Promise<void> {
  return apiFetch<void>(
    `/workitems/${workItemId}/relationships/${relationshipId}`,
    { method: "DELETE" },
  );
}
