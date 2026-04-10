import { lazy } from "react";
import { Routes, Route, Navigate } from "react-router";
import { AppShell } from "./layouts/AppShell";
import { useAuth } from "../hooks/useAuth";
import { LoadingState } from "../components/common/LoadingState";

const ProjectsPage = lazy(() =>
  import("../pages/ProjectsPage").then((m) => ({ default: m.ProjectsPage })),
);
const HomePage = lazy(() =>
  import("../pages/HomePage").then((m) => ({ default: m.HomePage })),
);
const BoardPage = lazy(() =>
  import("../pages/BoardPage").then((m) => ({ default: m.BoardPage })),
);
const ListPage = lazy(() =>
  import("../pages/ListPage").then((m) => ({ default: m.ListPage })),
);
const GraphPage = lazy(() =>
  import("../pages/GraphPage").then((m) => ({ default: m.GraphPage })),
);
const ItemDetailPage = lazy(() =>
  import("../pages/ItemDetailPage").then((m) => ({
    default: m.ItemDetailPage,
  })),
);
const SettingsPage = lazy(() =>
  import("../pages/SettingsPage").then((m) => ({ default: m.SettingsPage })),
);
const NotFoundPage = lazy(() =>
  import("../pages/NotFoundPage").then((m) => ({ default: m.NotFoundPage })),
);
const LoginPage = lazy(() =>
  import("../pages/LoginPage").then((m) => ({ default: m.LoginPage })),
);
const RegisterPage = lazy(() =>
  import("../pages/RegisterPage").then((m) => ({
    default: m.RegisterPage,
  })),
);

function ProtectedRoute({ children }: { children: React.ReactNode }) {
  const { isAuthenticated, isLoading } = useAuth();

  if (isLoading) {
    return <LoadingState />;
  }

  if (!isAuthenticated) {
    return <Navigate to="/login" replace />;
  }

  return <>{children}</>;
}

export function AppRouter() {
  return (
    <Routes>
      <Route path="/login" element={<LoginPage />} />
      <Route path="/register" element={<RegisterPage />} />
      <Route
        element={
          <ProtectedRoute>
            <AppShell />
          </ProtectedRoute>
        }
      >
        <Route index element={<ProjectsPage />} />
        <Route path="projects/:projectId">
          <Route index element={<HomePage />} />
          <Route path="board" element={<BoardPage />} />
          <Route path="list" element={<ListPage />} />
          <Route path="graph" element={<GraphPage />} />
          <Route path="items/:id" element={<ItemDetailPage />} />
        </Route>
        <Route path="settings" element={<SettingsPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  );
}
