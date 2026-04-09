import { useParams } from "react-router";

export function useProjectId(): string {
  const { projectId } = useParams<{ projectId: string }>();
  if (!projectId) {
    throw new Error("useProjectId must be used within a project-scoped route");
  }
  return projectId;
}
