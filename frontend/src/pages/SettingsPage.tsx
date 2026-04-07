import { getRole, isAdmin } from "../lib/config";

export function SettingsPage() {
  const role = getRole();

  return (
    <div className="mx-auto max-w-lg space-y-6">
      <h1 className="text-2xl font-semibold">Settings</h1>

      <div className="rounded-lg border border-gray-200 bg-white p-4 space-y-3">
        <div>
          <label className="text-xs font-semibold uppercase tracking-wide text-gray-500">
            Role
          </label>
          <p className="text-sm text-gray-900">
            {role}{" "}
            {isAdmin() && (
              <span className="text-xs text-blue-600">(admin access)</span>
            )}
          </p>
        </div>

        <div>
          <label className="text-xs font-semibold uppercase tracking-wide text-gray-500">
            API Backend
          </label>
          <p className="text-sm text-gray-900">
            {window.location.origin}/workitems
          </p>
        </div>
      </div>
    </div>
  );
}
