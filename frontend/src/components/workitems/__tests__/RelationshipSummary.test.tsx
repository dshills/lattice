import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { MemoryRouter } from "react-router";
import { RelationshipSummary } from "../RelationshipSummary";
import type { Relationship } from "../../../lib/types";

const relationships: Relationship[] = [
  { id: "r1", type: "depends_on", target_id: "target-1-uuid-long" },
  { id: "r2", type: "depends_on", target_id: "target-2-uuid-long" },
  { id: "r3", type: "blocks", target_id: "target-3-uuid-long" },
];

describe("RelationshipSummary", () => {
  it("renders empty message when no relationships", () => {
    render(
      <MemoryRouter>
        <RelationshipSummary relationships={[]} />
      </MemoryRouter>,
    );
    expect(screen.getByText("No relationships")).toBeInTheDocument();
  });

  it("renders grouped relationships in full mode", () => {
    render(
      <MemoryRouter>
        <RelationshipSummary relationships={relationships} />
      </MemoryRouter>,
    );
    expect(screen.getByText("depends on")).toBeInTheDocument();
    expect(screen.getByText("blocks")).toBeInTheDocument();
  });

  it("renders compact summary", () => {
    render(
      <MemoryRouter>
        <RelationshipSummary relationships={relationships} compact />
      </MemoryRouter>,
    );
    expect(screen.getByText("2 depends on, 1 blocks")).toBeInTheDocument();
  });

  it("shows remove buttons when onRemove provided", () => {
    const onRemove = vi.fn();
    render(
      <MemoryRouter>
        <RelationshipSummary
          relationships={relationships}
          onRemove={onRemove}
        />
      </MemoryRouter>,
    );

    const removeButtons = screen.getAllByText("Remove");
    expect(removeButtons).toHaveLength(3);

    fireEvent.click(removeButtons[0]);
    expect(onRemove).toHaveBeenCalledWith("r1");
  });
});
