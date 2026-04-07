import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent } from "@testing-library/react";
import { RelationshipEditor } from "../RelationshipEditor";

describe("RelationshipEditor", () => {
  it("shows add button initially", () => {
    render(
      <RelationshipEditor sourceId="source-id" onAdd={vi.fn()} />,
    );
    expect(screen.getByText("Add relationship")).toBeInTheDocument();
  });

  it("opens form on add button click", () => {
    render(
      <RelationshipEditor sourceId="source-id" onAdd={vi.fn()} />,
    );
    fireEvent.click(screen.getByText("Add relationship"));
    expect(screen.getByText("Type")).toBeInTheDocument();
    expect(screen.getByText("Target Item ID")).toBeInTheDocument();
  });

  it("prevents self-link", () => {
    const onAdd = vi.fn();
    render(
      <RelationshipEditor sourceId="source-id" onAdd={onAdd} />,
    );

    fireEvent.click(screen.getByText("Add relationship"));
    fireEvent.change(screen.getByPlaceholderText("Enter work item ID..."), {
      target: { value: "source-id" },
    });
    fireEvent.click(screen.getByText("Add"));

    expect(onAdd).not.toHaveBeenCalled();
    expect(
      screen.getByText("Cannot create a relationship to itself"),
    ).toBeInTheDocument();
  });

  it("calls onAdd with correct values", () => {
    const onAdd = vi.fn();
    render(
      <RelationshipEditor sourceId="source-id" onAdd={onAdd} />,
    );

    fireEvent.click(screen.getByText("Add relationship"));
    fireEvent.change(screen.getByPlaceholderText("Enter work item ID..."), {
      target: { value: "target-id" },
    });
    fireEvent.click(screen.getByText("Add"));

    expect(onAdd).toHaveBeenCalledWith("depends_on", "target-id");
  });

  it("shows direction preview", () => {
    render(
      <RelationshipEditor sourceId="source-id" onAdd={vi.fn()} />,
    );

    fireEvent.click(screen.getByText("Add relationship"));
    fireEvent.change(screen.getByPlaceholderText("Enter work item ID..."), {
      target: { value: "target-id" },
    });

    expect(
      screen.getByText(/This item depends on: target-i/),
    ).toBeInTheDocument();
  });
});
