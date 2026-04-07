import { describe, it, expect, vi, beforeEach } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { StateSelector } from "../StateSelector";

describe("StateSelector", () => {
  beforeEach(() => {
    window.__LATTICE_CONFIG__ = undefined;
  });

  it("shows forward transition as enabled for non-admin", () => {
    const onChange = vi.fn();
    render(<StateSelector current="NotDone" onChange={onChange} />);

    const inProgressBtn = screen.getByText("In Progress");
    expect(inProgressBtn).not.toBeDisabled();

    fireEvent.click(inProgressBtn);
    expect(onChange).toHaveBeenCalledWith("InProgress", false);
  });

  it("disables backward transition for non-admin", () => {
    render(<StateSelector current="InProgress" onChange={vi.fn()} />);
    expect(screen.getByText("Not Done")).toBeDisabled();
  });

  it("shows backward transition for admin", () => {
    window.__LATTICE_CONFIG__ = { role: "admin" };
    const onChange = vi.fn();
    render(<StateSelector current="InProgress" onChange={onChange} />);

    const notDoneBtn = screen.getByText("Not Done (Override)");
    expect(notDoneBtn).not.toBeDisabled();

    fireEvent.click(notDoneBtn);
    expect(onChange).toHaveBeenCalledWith("NotDone", true);
  });

  it("marks current state as selected", () => {
    render(<StateSelector current="Completed" onChange={vi.fn()} />);
    expect(screen.getByText("Completed")).toBeDisabled();
  });
});
