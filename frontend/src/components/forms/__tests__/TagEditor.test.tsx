import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { TagEditor } from "../TagEditor";

describe("TagEditor", () => {
  it("renders existing tags", () => {
    render(<TagEditor tags={["bug", "urgent"]} onUpdate={vi.fn()} />);
    expect(screen.getByText("bug")).toBeInTheDocument();
    expect(screen.getByText("urgent")).toBeInTheDocument();
  });

  it("adds a new tag on Enter", () => {
    const onUpdate = vi.fn();
    render(<TagEditor tags={["existing"]} onUpdate={onUpdate} />);

    const input = screen.getByPlaceholderText("Add tag...");
    fireEvent.change(input, { target: { value: "newtag" } });
    fireEvent.keyDown(input, { key: "Enter" });

    expect(onUpdate).toHaveBeenCalledWith(["existing", "newtag"]);
  });

  it("does not add duplicate tag", () => {
    const onUpdate = vi.fn();
    render(<TagEditor tags={["existing"]} onUpdate={onUpdate} />);

    const input = screen.getByPlaceholderText("Add tag...");
    fireEvent.change(input, { target: { value: "existing" } });
    fireEvent.keyDown(input, { key: "Enter" });

    expect(onUpdate).not.toHaveBeenCalled();
  });

  it("removes a tag on click", () => {
    const onUpdate = vi.fn();
    render(<TagEditor tags={["bug", "urgent"]} onUpdate={onUpdate} />);

    fireEvent.click(screen.getByLabelText("Remove tag bug"));
    expect(onUpdate).toHaveBeenCalledWith(["urgent"]);
  });
});
