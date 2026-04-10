import { Suspense, useEffect, useRef, useState } from "react";
import { NavLink, Outlet, useLocation, useNavigate, useParams } from "react-router";
import { CreateWorkItemForm } from "../../components/forms/CreateWorkItemForm";
import { LoadingState } from "../../components/common/LoadingState";
import { useProjects } from "../../hooks/useProjects";
import { useAuth } from "../../hooks/useAuth";

export function AppShell() {
  const [createOpen, setCreateOpen] = useState(false);
  const mainRef = useRef<HTMLElement>(null);
  const location = useLocation();
  const navigate = useNavigate();
  const { projectId } = useParams<{ projectId: string }>();
  const { data: projectsData } = useProjects();
  const { user, logout } = useAuth();

  useEffect(() => {
    mainRef.current?.focus();
  }, [location.pathname]);

  const projects = projectsData?.projects ?? [];
  const currentProject = projects.find((p) => p.id === projectId);

  // Determine the current view path segment for navigation
  const viewMatch = location.pathname.match(
    /\/projects\/[^/]+\/?(board|list|graph|items\/.+)?/,
  );
  const currentView = viewMatch?.[1] ?? "";

  const handleProjectSwitch = (newProjectId: string) => {
    const view = currentView.startsWith("items/") ? "board" : currentView;
    const path = view
      ? `/projects/${newProjectId}/${view}`
      : `/projects/${newProjectId}`;
    navigate(`${path}${location.search}`);
  };

  const navItems = projectId
    ? [
        { to: `/projects/${projectId}`, label: "Home", end: true },
        { to: `/projects/${projectId}/board`, label: "Board", end: false },
        { to: `/projects/${projectId}/list`, label: "List", end: false },
        { to: `/projects/${projectId}/graph`, label: "Graph", end: false },
      ]
    : [];

  return (
    <div className="flex h-screen bg-gray-50 text-gray-900">
      <a
        href="#main-content"
        className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-50 focus:rounded-md focus:bg-blue-600 focus:px-4 focus:py-2 focus:text-white"
      >
        Skip to content
      </a>
      <aside className="w-56 flex-shrink-0 border-r border-gray-200 bg-white" role="navigation" aria-label="Main navigation">
        <NavLink
          to="/"
          className="block px-4 py-5 text-lg font-semibold tracking-tight hover:text-blue-600"
        >
          Lattice
        </NavLink>

        {projectId && projects.length > 0 && (
          <div className="px-2 pb-3">
            <select
              value={projectId}
              onChange={(e) => handleProjectSwitch(e.target.value)}
              className="w-full rounded-md border border-gray-300 px-2 py-1.5 text-sm focus:border-blue-500 focus:outline-none focus:ring-1 focus:ring-blue-500"
              aria-label="Select project"
            >
              {projects.map((p) => (
                <option key={p.id} value={p.id}>
                  {p.name}
                </option>
              ))}
            </select>
          </div>
        )}

        <nav className="flex flex-col gap-1 px-2">
          {navItems.map((item) => (
            <NavLink
              key={item.to}
              to={item.to}
              end={item.end}
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
          <NavLink
            to="/settings"
            className={({ isActive }) =>
              `rounded-md px-3 py-2 text-sm font-medium transition-colors ${
                isActive
                  ? "bg-gray-100 text-gray-900"
                  : "text-gray-600 hover:bg-gray-50 hover:text-gray-900"
              }`
            }
          >
            Settings
          </NavLink>
        </nav>
      </aside>
      <div className="flex flex-1 flex-col overflow-hidden">
        <header className="flex h-14 items-center justify-between border-b border-gray-200 bg-white px-6">
          <div className="text-sm font-medium text-gray-500">
            {currentProject?.name ?? "Projects"}
          </div>
          <div className="flex items-center gap-3">
            {projectId && (
              <button
                onClick={() => setCreateOpen(true)}
                className="rounded-md bg-blue-600 px-3 py-1.5 text-sm font-medium text-white hover:bg-blue-700"
              >
                Create
              </button>
            )}
            {user && (
              <span className="text-sm text-gray-500">{user.display_name}</span>
            )}
            <button
              onClick={() => {
                logout();
                navigate("/login");
              }}
              className="rounded-md px-3 py-1.5 text-sm font-medium text-gray-500 hover:text-gray-700"
            >
              Sign out
            </button>
          </div>
        </header>
        <main
          id="main-content"
          ref={mainRef}
          tabIndex={-1}
          className="flex-1 overflow-auto p-6 outline-none"
          aria-label="Page content"
        >
          <Suspense fallback={<LoadingState />}>
            <Outlet />
          </Suspense>
        </main>
      </div>

      {projectId && (
        <CreateWorkItemForm
          open={createOpen}
          onClose={() => setCreateOpen(false)}
        />
      )}
    </div>
  );
}
