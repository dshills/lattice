import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import { describe, it, expect, vi } from "vitest";
import { ErrorState } from "../ErrorState";

describe("ErrorState", () => {
  it("renders error message", () => {
    render(<ErrorState message="Something broke" />);
    expect(screen.getByText("Something broke")).toBeInTheDocument();
    expect(screen.getByText("Something went wrong")).toBeInTheDocument();
  });

  it("renders retry button when onRetry provided", async () => {
    const onRetry = vi.fn();
    const user = userEvent.setup();

    render(<ErrorState message="Failure" onRetry={onRetry} />);

    await user.click(screen.getByText("Try again"));
    expect(onRetry).toHaveBeenCalledOnce();
  });

  it("hides retry button when no onRetry", () => {
    render(<ErrorState message="Failure" />);
    expect(screen.queryByText("Try again")).not.toBeInTheDocument();
  });
});
