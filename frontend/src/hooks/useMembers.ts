import { useQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import {
  listMembers,
  addMember,
  updateMemberRole,
  removeMember,
} from "../lib/api/members";
import { toastError } from "../lib/toastError";
import { useToast } from "../components/common/Toast";

export function useMembers(projectId: string | undefined) {
  return useQuery({
    queryKey: ["members", projectId],
    queryFn: () => listMembers(projectId!),
    enabled: !!projectId,
  });
}

export function useMemberMutations(projectId: string | undefined) {
  const queryClient = useQueryClient();
  const { addToast } = useToast();

  const invalidate = () =>
    queryClient.invalidateQueries({ queryKey: ["members", projectId] });

  const addMemberMutation = useMutation({
    mutationFn: ({ email, role }: { email: string; role: string }) =>
      addMember(projectId!, email, role),
    onSuccess: () => {
      invalidate();
      addToast("Member added", "info");
    },
    onError: (err) => toastError(addToast, err),
  });

  const updateRoleMutation = useMutation({
    mutationFn: ({ userId, role }: { userId: string; role: string }) =>
      updateMemberRole(projectId!, userId, role),
    onSuccess: () => {
      invalidate();
      addToast("Role updated", "info");
    },
    onError: (err) => toastError(addToast, err),
  });

  const removeMemberMutation = useMutation({
    mutationFn: (userId: string) => removeMember(projectId!, userId),
    onSuccess: () => {
      invalidate();
      addToast("Member removed", "info");
    },
    onError: (err) => toastError(addToast, err),
  });

  return { addMemberMutation, updateRoleMutation, removeMemberMutation };
}
