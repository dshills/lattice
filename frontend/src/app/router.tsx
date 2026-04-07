import { Routes, Route } from "react-router";
import { AppShell } from "./layouts/AppShell";
import { HomePage } from "../pages/HomePage";
import { BoardPage } from "../pages/BoardPage";
import { ListPage } from "../pages/ListPage";
import { GraphPage } from "../pages/GraphPage";
import { ItemDetailPage } from "../pages/ItemDetailPage";
import { SettingsPage } from "../pages/SettingsPage";
import { NotFoundPage } from "../pages/NotFoundPage";

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
