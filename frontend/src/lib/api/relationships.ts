import type { Relationship, AddRelationshipInput } from "../types";
import { apiFetch } from "./client";

export function addRelationship(
  projectId: string,
  workItemId: string,
  input: AddRelationshipInput,
): Promise<Relationship> {
  return apiFetch<Relationship>(
    `/projects/${projectId}/workitems/${workItemId}/relationships`,
    {
      method: "POST",
      body: JSON.stringify(input),
    },
  );
}

export function removeRelationship(
  projectId: string,
  workItemId: string,
  relationshipId: string,
): Promise<void> {
  return apiFetch<void>(
    `/projects/${projectId}/workitems/${workItemId}/relationships/${relationshipId}`,
    { method: "DELETE" },
  );
}
