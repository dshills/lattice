import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { FilterPanel } from "../FilterPanel";

describe("FilterPanel", () => {
  const defaultProps = {
    filters: {},
    setFilter: vi.fn(),
    clearFilters: vi.fn(),
    activeFilterCount: 0,
  };

  it("renders state filter buttons", () => {
    render(<FilterPanel {...defaultProps} />);
    expect(screen.getByText("Not Done")).toBeInTheDocument();
    expect(screen.getByText("In Progress")).toBeInTheDocument();
    expect(screen.getByText("Completed")).toBeInTheDocument();
  });

  it("calls setFilter when state button clicked", () => {
    const setFilter = vi.fn();
    render(<FilterPanel {...defaultProps} setFilter={setFilter} />);
    fireEvent.click(screen.getByText("In Progress"));
    expect(setFilter).toHaveBeenCalledWith("state", "InProgress");
  });

  it("shows clear all when filters active", () => {
    render(<FilterPanel {...defaultProps} activeFilterCount={2} />);
    expect(screen.getByText("Clear all (2)")).toBeInTheDocument();
  });

  it("disables Ready when Blocked is active", () => {
    render(
      <FilterPanel
        {...defaultProps}
        filters={{ is_blocked: true }}
      />,
    );
    expect(screen.getByText("Ready")).toBeDisabled();
  });

  it("disables Blocked when Ready is active", () => {
    render(
      <FilterPanel
        {...defaultProps}
        filters={{ is_ready: true }}
      />,
    );
    expect(screen.getByText("Blocked")).toBeDisabled();
  });
});
