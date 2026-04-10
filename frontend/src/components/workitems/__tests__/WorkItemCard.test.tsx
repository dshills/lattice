import { describe, it, expect } from "vitest";
import { render, screen } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { WorkItemCard } from "../WorkItemCard";
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
    assignee_id: null,
    created_by: null,
    relationships: [],
    created_at: new Date().toISOString(),
    updated_at: new Date().toISOString(),
    is_blocked: false,
    ...overrides,
  };
}

function renderCard(item: WorkItem) {
  return render(
    <MemoryRouter>
      <WorkItemCard item={item} />
    </MemoryRouter>,
  );
}

describe("WorkItemCard", () => {
  it("renders title", () => {
    renderCard(makeItem({ title: "Fix login bug" }));
    expect(screen.getByText("Fix login bug")).toBeInTheDocument();
  });

  it("renders type badge", () => {
    renderCard(makeItem({ type: "bug" }));
    expect(screen.getByText("bug")).toBeInTheDocument();
  });

  it("renders tags up to 3", () => {
    renderCard(
      makeItem({ tags: ["alpha", "beta", "gamma", "delta", "epsilon"] }),
    );
    expect(screen.getByText("alpha")).toBeInTheDocument();
    expect(screen.getByText("beta")).toBeInTheDocument();
    expect(screen.getByText("gamma")).toBeInTheDocument();
    expect(screen.getByText("+2")).toBeInTheDocument();
  });

  it("shows blocked icon when is_blocked", () => {
    renderCard(makeItem({ is_blocked: true }));
    expect(screen.getByTitle("Blocked by dependency")).toBeInTheDocument();
  });

  it("does not show blocked icon when not blocked", () => {
    renderCard(makeItem({ is_blocked: false }));
    expect(screen.queryByTitle("Blocked by dependency")).not.toBeInTheDocument();
  });

  it("renders condition tags with emphasis", () => {
    renderCard(makeItem({ tags: ["blocked", "regular"] }));
    expect(screen.getByText("blocked")).toBeInTheDocument();
    expect(screen.getByText("regular")).toBeInTheDocument();
  });
});
