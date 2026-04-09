import type {
  WorkItem,
  ListResponse,
  ListFilter,
  CreateWorkItemInput,
  UpdateWorkItemInput,
} from "../types";
import { apiFetch } from "./client";

function buildQueryString(filter: ListFilter): string {
  const params = new URLSearchParams();

  if (filter.state) params.set("state", filter.state);
  if (filter.tags) params.set("tags", filter.tags);
  if (filter.type) params.set("type", filter.type);
  if (filter.parent_id) params.set("parent_id", filter.parent_id);
  if (filter.is_blocked !== undefined)
    params.set("is_blocked", String(filter.is_blocked));
  if (filter.is_ready !== undefined)
    params.set("is_ready", String(filter.is_ready));
  if (filter.page !== undefined) params.set("page", String(filter.page));
  if (filter.page_size !== undefined)
    params.set("page_size", String(filter.page_size));

  const qs = params.toString();
  return qs ? `?${qs}` : "";
}

function base(projectId: string): string {
  return `/projects/${projectId}/workitems`;
}

export function listWorkItems(
  projectId: string,
  filter: ListFilter = {},
): Promise<ListResponse> {
  return apiFetch<ListResponse>(
    `${base(projectId)}${buildQueryString(filter)}`,
  );
}

export function getWorkItem(
  projectId: string,
  id: string,
): Promise<WorkItem> {
  return apiFetch<WorkItem>(`${base(projectId)}/${id}`);
}

export function createWorkItem(
  projectId: string,
  input: CreateWorkItemInput,
): Promise<WorkItem> {
  return apiFetch<WorkItem>(base(projectId), {
    method: "POST",
    body: JSON.stringify(input),
  });
}

export function updateWorkItem(
  projectId: string,
  id: string,
  input: UpdateWorkItemInput,
): Promise<WorkItem> {
  return apiFetch<WorkItem>(`${base(projectId)}/${id}`, {
    method: "PATCH",
    body: JSON.stringify(input),
  });
}

export function deleteWorkItem(
  projectId: string,
  id: string,
): Promise<void> {
  return apiFetch<void>(`${base(projectId)}/${id}`, {
    method: "DELETE",
  });
}
