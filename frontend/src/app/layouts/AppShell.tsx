import { useState } from "react";
import { NavLink, Outlet } from "react-router";
import { CreateWorkItemForm } from "../../components/forms/CreateWorkItemForm";

const navItems = [
  { to: "/", label: "Home" },
  { to: "/board", label: "Board" },
  { to: "/list", label: "List" },
  { to: "/graph", label: "Graph" },
  { to: "/settings", label: "Settings" },
];

export function AppShell() {
  const [createOpen, setCreateOpen] = useState(false);

  return (
    <div className="flex h-screen bg-gray-50 text-gray-900">
      <aside className="w-56 flex-shrink-0 border-r border-gray-200 bg-white">
        <div className="px-4 py-5 text-lg font-semibold tracking-tight">
          Lattice
        </div>
        <nav className="flex flex-col gap-1 px-2">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.to === "/"}
              className={({ isActive }) =>
                `rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                  isActive
                    ? "bg-gray-100 text-gray-900"
                    : "text-gray-600 hover:bg-gray-50 hover:text-gray-900"
                }`
              }
            >
              {item.label}
            </NavLink>
          ))}
        </nav>
      </aside>
      <div className="flex flex-1 flex-col overflow-hidden">
        <header className="flex h-14 items-center justify-between border-b border-gray-200 bg-white px-6">
          <div className="text-sm font-medium text-gray-500">Workspace</div>
          <div className="flex items-center gap-3">
            <input
              type="text"
              placeholder="Search..."
              className="rounded-md border border-gray-300 px-3 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
            />
            <button
              onClick={() => setCreateOpen(true)}
              className="rounded-md bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700"
            >
              Create
            </button>
          </div>
        </header>
        <main className="flex-1 overflow-auto p-6">
          <Outlet />
        </main>
      </div>

      <CreateWorkItemForm
        open={createOpen}
        onClose={() => setCreateOpen(false)}
      />
    </div>
  );
}
