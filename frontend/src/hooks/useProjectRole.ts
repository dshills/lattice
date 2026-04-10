import { useMemo } from "react";
import { useMembers } from "./useMembers";
import { useAuth } from "./useAuth";

export function useProjectRole(projectId: string | undefined) {
  const { user } = useAuth();
  const { data: members, isLoading } = useMembers(projectId);

  const role = useMemo(() => {
    if (!members || !user) return null;
    const membership = members.find((m) => m.user_id === user.id);
    return membership?.role ?? null;
  }, [members, user]);

  return {
    role,
    isOwner: role === "owner",
    canWrite: role === "owner" || role === "member",
    isLoading,
  };
}
