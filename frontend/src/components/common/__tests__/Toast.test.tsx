import { render, screen, act, fireEvent } from "@testing-library/react";
import { describe, it, expect, vi, beforeEach, afterEach } from "vitest";
import { ToastProvider, useToast } from "../Toast";

function TestTrigger() {
  const { addToast } = useToast();
  return (
    <>
      <button onClick={() => addToast("Success!", "success")}>
        Show success
      </button>
      <button onClick={() => addToast("Error!", "error")}>Show error</button>
    </>
  );
}

describe("Toast", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("shows and auto-dismisses a toast", () => {
    render(
      <ToastProvider>
        <TestTrigger />
      </ToastProvider>,
    );

    fireEvent.click(screen.getByText("Show success"));
    expect(screen.getByText("Success!")).toBeInTheDocument();

    act(() => {
      vi.advanceTimersByTime(4000);
    });

    expect(screen.queryByText("Success!")).not.toBeInTheDocument();
  });

  it("dismisses on click", () => {
    render(
      <ToastProvider>
        <TestTrigger />
      </ToastProvider>,
    );

    fireEvent.click(screen.getByText("Show error"));
    expect(screen.getByText("Error!")).toBeInTheDocument();

    fireEvent.click(screen.getByLabelText("Dismiss"));
    expect(screen.queryByText("Error!")).not.toBeInTheDocument();
  });

  it("renders correct type styles", () => {
    render(
      <ToastProvider>
        <TestTrigger />
      </ToastProvider>,
    );

    fireEvent.click(screen.getByText("Show error"));
    const toast = screen.getByText("Error!").closest("[role='status']");
    expect(toast?.className).toContain("bg-red-50");
  });
});
