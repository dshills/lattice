import { apiFetch } from "./client";

export interface ProjectMember {
  id: string;
  project_id: string;
  user_id: string;
  role: "owner" | "member" | "viewer";
  created_at: string;
  email: string;
  display_name: string;
}

export function listMembers(projectId: string): Promise<ProjectMember[]> {
  return apiFetch<ProjectMember[]>(`/projects/${projectId}/members`);
}

export function addMember(
  projectId: string,
  email: string,
  role: string,
): Promise<ProjectMember> {
  return apiFetch<ProjectMember>(`/projects/${projectId}/members`, {
    method: "POST",
    body: JSON.stringify({ email, role }),
  });
}

export function updateMemberRole(
  projectId: string,
  userId: string,
  role: string,
): Promise<void> {
  return apiFetch<void>(`/projects/${projectId}/members/${userId}`, {
    method: "PATCH",
    body: JSON.stringify({ role }),
  });
}

export function removeMember(
  projectId: string,
  userId: string,
): Promise<void> {
  return apiFetch<void>(`/projects/${projectId}/members/${userId}`, {
    method: "DELETE",
  });
}
