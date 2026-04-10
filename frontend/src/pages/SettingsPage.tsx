import { useAuth } from "../hooks/useAuth";

export function SettingsPage() {
  const { user } = useAuth();

  return (
    <div className="mx-auto max-w-lg space-y-6">
      <h1 className="text-2xl font-semibold">Settings</h1>

      <div className="rounded-lg border border-gray-200 bg-white p-4 space-y-3">
        <div>
          <label className="text-xs font-semibold uppercase tracking-wide text-gray-500">
            Email
          </label>
          <p className="text-sm text-gray-900">{user?.email ?? "-"}</p>
        </div>

        <div>
          <label className="text-xs font-semibold uppercase tracking-wide text-gray-500">
            Display Name
          </label>
          <p className="text-sm text-gray-900">{user?.display_name ?? "-"}</p>
        </div>

        <div>
          <label className="text-xs font-semibold uppercase tracking-wide text-gray-500">
            API Backend
          </label>
          <p className="text-sm text-gray-900">
            {window.location.origin}/projects
          </p>
        </div>
      </div>
    </div>
  );
}
