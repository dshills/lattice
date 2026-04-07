import { lazy } from "react";
import { Routes, Route } from "react-router";
import { AppShell } from "./layouts/AppShell";

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

export function AppRouter() {
  return (
    <Routes>
      <Route element={<AppShell />}>
        <Route index element={<HomePage />} />
        <Route path="board" element={<BoardPage />} />
        <Route path="list" element={<ListPage />} />
        <Route path="graph" element={<GraphPage />} />
        <Route path="items/:id" element={<ItemDetailPage />} />
        <Route path="settings" element={<SettingsPage />} />
        <Route path="*" element={<NotFoundPage />} />
      </Route>
    </Routes>
  );
}
