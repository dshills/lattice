import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter, Routes, Route } from "react-router";
import { WorkItemRow } from "../WorkItemRow";
import type { WorkItem } from "../../../lib/types";

function makeItem(overrides: Partial<WorkItem> = {}): WorkItem {
  return {
    id: "test-id",
    project_id: "test-project",
    title: "Test Item",
    description: "A description",
    state: "InProgress",
    tags: ["bug", "urgent"],
    type: "task",
    parent_id: null,
    assignee_id: null,
    created_by: null,
    relationships: [
      { id: "r1", type: "depends_on", target_id: "dep-id" },
    ],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    is_blocked: false,
    ...overrides,
  };
}

function renderRow(item: WorkItem) {
  return render(
    <MemoryRouter initialEntries={["/projects/test-project/list"]}>
      <Routes>
        <Route path="/projects/:projectId/list" element={
          <table>
            <tbody>
              <WorkItemRow item={item} />
            </tbody>
          </table>
        } />
      </Routes>
    </MemoryRouter>,
  );
}

describe("WorkItemRow", () => {
  it("renders title as link", () => {
    renderRow(makeItem());
    const link = screen.getByText("Test Item");
    expect(link).toBeInTheDocument();
    expect(link.closest("a")).toHaveAttribute(
      "href",
      "/projects/test-project/items/test-id",
    );
  });

  it("renders state badge", () => {
    renderRow(makeItem());
    expect(screen.getByText("In Progress")).toBeInTheDocument();
  });

  it("renders type badge", () => {
    renderRow(makeItem());
    expect(screen.getByText("task")).toBeInTheDocument();
  });

  it("renders tags", () => {
    renderRow(makeItem());
    expect(screen.getByText("bug")).toBeInTheDocument();
    expect(screen.getByText("urgent")).toBeInTheDocument();
  });

  it("renders relationship summary", () => {
    renderRow(makeItem());
    expect(screen.getByText("1 depends on")).toBeInTheDocument();
  });
});
