import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Routes, Route } from "react-router";
import { DndContext } from "@dnd-kit/core";
import { QueryClient, QueryClientProvider } from "@tanstack/react-query";
import { BoardColumn } from "../BoardColumn";
import type { WorkItem } from "../../../lib/types";

function makeItem(overrides: Partial<WorkItem> = {}): WorkItem {
  return {
    id: "test-id",
    project_id: "test-project",
    title: "Test Item",
    description: "",
    state: "NotDone",
    tags: [],
    type: "",
    parent_id: null,
    relationships: [],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    is_blocked: false,
    ...overrides,
  };
}

function renderColumn(items: WorkItem[]) {
  const queryClient = new QueryClient();
  return render(
    <QueryClientProvider client={queryClient}>
      <MemoryRouter initialEntries={["/projects/test-project/board"]}>
        <Routes>
          <Route path="/projects/:projectId/board" element={
            <DndContext>
              <BoardColumn state="NotDone" items={items} disabled={false} />
            </DndContext>
          } />
        </Routes>
      </MemoryRouter>
    </QueryClientProvider>,
  );
}

describe("BoardColumn", () => {
  it("renders column header with count", () => {
    renderColumn([makeItem(), makeItem({ id: "2", title: "Second" })]);
    expect(screen.getByText("Not Done")).toBeInTheDocument();
    expect(screen.getByText("2")).toBeInTheDocument();
  });

  it("renders empty state when no items", () => {
    renderColumn([]);
    expect(screen.getByText("No items")).toBeInTheDocument();
  });

  it("renders item titles", () => {
    renderColumn([
      makeItem({ id: "1", title: "First" }),
      makeItem({ id: "2", title: "Second" }),
    ]);
    expect(screen.getByText("First")).toBeInTheDocument();
    expect(screen.getByText("Second")).toBeInTheDocument();
  });
});
