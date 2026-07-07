import { render, screen } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";
import ConnectClientButtonsHarness from "./test/ConnectClientButtonsHarness.svelte";

describe("ConnectClientButtons", () => {
  it("renders a button for each branded provider plus a generic option", () => {
    render(ConnectClientButtonsHarness, { props: { open: vi.fn() } });

    expect(screen.getByRole("button", { name: "Claude" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "ChatGPT" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Gemini" })).toBeInTheDocument();
    expect(
      screen.getByRole("button", { name: "Other client" }),
    ).toBeInTheDocument();
  });

  it("calls open with the selected provider", () => {
    const open = vi.fn();
    render(ConnectClientButtonsHarness, { props: { open } });

    screen.getByRole("button", { name: "Claude" }).click();
    expect(open).toHaveBeenCalledWith("claude");

    screen.getByRole("button", { name: "ChatGPT" }).click();
    expect(open).toHaveBeenCalledWith("openai");

    screen.getByRole("button", { name: "Gemini" }).click();
    expect(open).toHaveBeenCalledWith("gemini");

    screen.getByRole("button", { name: "Other client" }).click();
    expect(open).toHaveBeenCalledWith("other");
  });
});
