import { useState } from "react";
import { useMembers, useMemberMutations } from "../hooks/useMembers";
import { useProjectRole } from "../hooks/useProjectRole";
import { useProjectId } from "../hooks/useProjectId";
import { useAuth } from "../hooks/useAuth";
import { LoadingState } from "../components/common/LoadingState";
import { ErrorState } from "../components/common/ErrorState";
import { ConfirmDialog } from "../components/common/ConfirmDialog";
import type { ProjectMember } from "../lib/api/members";

const ROLES = ["owner", "member", "viewer"] as const;

const ROLE_COLORS: Record<string, string> = {
  owner: "bg-purple-100 text-purple-800",
  member: "bg-blue-100 text-blue-800",
  viewer: "bg-gray-100 text-gray-600",
};

export function MembersPage() {
  const projectId = useProjectId();
  const { user } = useAuth();
  const { isOwner } = useProjectRole(projectId);
  const { data: members, isLoading, error, refetch } = useMembers(projectId);
  const { addMemberMutation, updateRoleMutation, removeMemberMutation } =
    useMemberMutations(projectId);

  const [email, setEmail] = useState("");
  const [role, setRole] = useState<string>("member");
  const [removingMember, setRemovingMember] = useState<ProjectMember | null>(
    null,
  );

  const handleInvite = (e: React.FormEvent) => {
    e.preventDefault();
    if (!email.trim()) return;
    addMemberMutation.mutate(
      { email: email.trim(), role },
      { onSuccess: () => setEmail("") },
    );
  };

  if (isLoading) return <LoadingState />;
  if (error)
    return <ErrorState message={error.message} onRetry={() => refetch()} />;

  return (
    <div className="mx-auto max-w-2xl space-y-6">
      <h1 className="text-xl font-semibold">Members</h1>

      {isOwner && (
        <form
          onSubmit={handleInvite}
          className="flex items-end gap-3 rounded-lg border border-gray-200 bg-white p-4"
        >
          <div className="flex-1">
            <label className="mb-1 block text-xs font-medium text-gray-500">
              Email
            </label>
            <input
              type="email"
              required
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              placeholder="user@example.com"
              className="w-full rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
          </div>
          <div>
            <label className="mb-1 block text-xs font-medium text-gray-500">
              Role
            </label>
            <select
              value={role}
              onChange={(e) => setRole(e.target.value)}
              className="rounded-md border border-gray-300 px-3 py-2 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            >
              {ROLES.map((r) => (
                <option key={r} value={r}>
                  {r}
                </option>
              ))}
            </select>
          </div>
          <button
            type="submit"
            disabled={addMemberMutation.isPending}
            className="rounded-md bg-blue-600 px-4 py-2 text-sm font-medium text-white hover:bg-blue-700 disabled:opacity-50"
          >
            {addMemberMutation.isPending ? "Adding..." : "Add"}
          </button>
        </form>
      )}

      <div className="rounded-lg border border-gray-200 bg-white">
        {(members ?? []).map((member) => (
          <div
            key={member.user_id}
            className="flex items-center justify-between border-b border-gray-100 px-4 py-3 last:border-b-0"
          >
            <div>
              <p className="text-sm font-medium text-gray-900">
                {member.display_name}
                {member.user_id === user?.id && (
                  <span className="ml-1 text-xs text-gray-400">(you)</span>
                )}
              </p>
              <p className="text-xs text-gray-500">{member.email}</p>
            </div>
            <div className="flex items-center gap-2">
              {isOwner && member.user_id !== user?.id ? (
                <select
                  value={member.role}
                  onChange={(e) =>
                    updateRoleMutation.mutate({
                      userId: member.user_id,
                      role: e.target.value,
                    })
                  }
                  className="rounded-md border border-gray-300 px-2 py-1 text-xs focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
                >
                  {ROLES.map((r) => (
                    <option key={r} value={r}>
                      {r}
                    </option>
                  ))}
                </select>
              ) : (
                <span
                  className={`rounded-full px-2.5 py-0.5 text-xs font-medium ${ROLE_COLORS[member.role] ?? ROLE_COLORS.viewer}`}
                >
                  {member.role}
                </span>
              )}
              {isOwner && member.user_id !== user?.id && (
                <button
                  onClick={() => setRemovingMember(member)}
                  className="rounded px-2 py-1 text-xs text-red-500 hover:bg-red-50"
                >
                  Remove
                </button>
              )}
            </div>
          </div>
        ))}
      </div>

      <ConfirmDialog
        open={!!removingMember}
        onClose={() => setRemovingMember(null)}
        onConfirm={() => {
          if (removingMember) {
            removeMemberMutation.mutate(removingMember.user_id, {
              onSuccess: () => setRemovingMember(null),
            });
          }
        }}
        title="Remove member"
        message={`Remove ${removingMember?.display_name} from this project?`}
        confirmLabel="Remove"
        danger
      />
    </div>
  );
}
