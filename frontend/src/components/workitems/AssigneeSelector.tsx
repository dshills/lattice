import { useMembers } from "../../hooks/useMembers";
import { useProjectId } from "../../hooks/useProjectId";

interface AssigneeSelectorProps {
  value: string | null;
  onChange: (assigneeId: string | null) => void;
  disabled?: boolean;
}

export function AssigneeSelector({ value, onChange, disabled }: AssigneeSelectorProps) {
  const projectId = useProjectId();
  const { data: members } = useMembers(projectId);

  return (
    <select
      value={value ?? ""}
      onChange={(e) => onChange(e.target.value || null)}
      disabled={disabled}
      className="w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500 disabled:opacity-50"
    >
      <option value="">Unassigned</option>
      {(members ?? []).map((m) => (
        <option key={m.user_id} value={m.user_id}>
          {m.display_name}
        </option>
      ))}
    </select>
  );
}
