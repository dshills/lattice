import { describe, it, expect, vi } from "vitest";
import { render, screen, fireEvent, act } from "@testing-library/react";
import { InlineEditableText } from "../InlineEditableText";

describe("InlineEditableText", () => {
  it("renders initial value", () => {
    render(<InlineEditableText value="Hello" onSave={vi.fn()} />);
    expect(screen.getByDisplayValue("Hello")).toBeInTheDocument();
  });

  it("calls onSave after debounce on change", async () => {
    vi.useFakeTimers();
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(<InlineEditableText value="Hello" onSave={onSave} />);

    const input = screen.getByDisplayValue("Hello");
    fireEvent.change(input, { target: { value: "World" } });

    expect(onSave).not.toHaveBeenCalled();

    await act(async () => {
      vi.advanceTimersByTime(500);
    });

    expect(onSave).toHaveBeenCalledWith("World");
    vi.useRealTimers();
  });

  it("shows 'Saved' indicator on success", async () => {
    vi.useFakeTimers();
    const onSave = vi.fn().mockResolvedValue(undefined);
    render(<InlineEditableText value="Hello" onSave={onSave} />);

    const input = screen.getByDisplayValue("Hello");
    fireEvent.change(input, { target: { value: "World" } });

    await act(async () => {
      vi.advanceTimersByTime(500);
    });

    expect(screen.getByText("Saved")).toBeInTheDocument();
    vi.useRealTimers();
  });

  it("shows error indicator on failure", async () => {
    vi.useFakeTimers();
    const onSave = vi.fn().mockRejectedValue(new Error("fail"));
    render(<InlineEditableText value="Hello" onSave={onSave} />);

    const input = screen.getByDisplayValue("Hello");
    fireEvent.change(input, { target: { value: "World" } });

    await act(async () => {
      vi.advanceTimersByTime(500);
    });

    expect(screen.getByText("Failed to save")).toBeInTheDocument();
    vi.useRealTimers();
  });

  it("renders as textarea when as='textarea'", () => {
    render(
      <InlineEditableText value="Text" onSave={vi.fn()} as="textarea" />,
    );
    expect(screen.getByDisplayValue("Text").tagName).toBe("TEXTAREA");
  });
});
