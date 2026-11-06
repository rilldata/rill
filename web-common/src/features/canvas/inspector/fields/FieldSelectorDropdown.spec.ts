import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";
import { readable } from "svelte/store";
import Harness from "../../../../../tests/fixtures/FieldSelectorDropdownHarness.svelte";

vi.mock("@rilldata/web-common/runtime-client/v2", async (original) => ({
  ...(await original<
    typeof import("@rilldata/web-common/runtime-client/v2")
  >()),
  useRuntimeClient: () => ({ instanceId: "default" }),
}));
vi.mock(
  "@rilldata/web-common/features/canvas/state-managers/state-managers",
  () => ({ getCanvasStore: () => ({}) }),
);
vi.mock("@rilldata/web-common/features/canvas/inspector/selectors", () => ({
  useMetricFieldData: () =>
    readable({
      filteredItems: ["a", "b", "c"],
      displayMap: {
        a: { type: "measure", label: "A" },
        b: { type: "measure", label: "B" },
        c: { type: "measure", label: "C" },
      },
    }),
}));
afterEach(cleanup);

async function openMenu() {
  const trigger = screen.getByRole("button", { name: "Choose fields" });
  await fireEvent.keyDown(trigger, { key: "Enter" });
  return screen.findByRole("menuitemcheckbox", { name: "B", exact: true });
}

describe("field selector optimistic selection", () => {
  it("preserves the latest toggle through older server snapshots", async () => {
    const onMultiSelect = vi.fn();
    const view = render(Harness, { selectedItems: ["a", "b"], onMultiSelect });
    const b = await openMenu();
    expect(b).toHaveAttribute("aria-checked", "true");
    await fireEvent.click(b);
    expect(onMultiSelect).toHaveBeenLastCalledWith(["a"]);
    expect(b).toHaveAttribute("aria-checked", "false");
    await view.rerender({ selectedItems: ["a", "b"] });
    expect.soft(b).toHaveAttribute("aria-checked", "false");
    await fireEvent.click(b);
    expect(onMultiSelect).toHaveBeenLastCalledWith(["a", "b"]);
    expect(b).toHaveAttribute("aria-checked", "true");
    await view.rerender({ selectedItems: ["a"] });
    expect(b).toHaveAttribute("aria-checked", "true");
    expect(onMultiSelect).toHaveBeenCalledTimes(2);
  });

  it("preserves unrelated external additions and removals while open", async () => {
    const onMultiSelect = vi.fn();
    const view = render(Harness, { selectedItems: ["a", "b"], onMultiSelect });
    const b = await openMenu();
    await view.rerender({ selectedItems: ["a", "b", "c"] });
    expect(onMultiSelect).not.toHaveBeenCalled();
    await fireEvent.click(b);
    expect(onMultiSelect).toHaveBeenLastCalledWith(["a", "c"]);
    await view.rerender({ selectedItems: ["b", "c"] });
    expect(b).toHaveAttribute("aria-checked", "false");
    expect(
      screen.getByRole("menuitemcheckbox", { name: "A", exact: true }),
    ).toHaveAttribute("aria-checked", "false");
    await fireEvent.click(b);
    expect(onMultiSelect).toHaveBeenLastCalledWith(["b", "c"]);
  });

  it("clears session intent when closed and reopens with the current props", async () => {
    const onMultiSelect = vi.fn();
    const view = render(Harness, { selectedItems: ["a", "b"], onMultiSelect });
    const b = await openMenu();
    await fireEvent.click(b);
    await view.rerender({ open: false, selectedItems: ["b", "c"] });
    const reopened = await openMenu();
    expect(reopened).toHaveAttribute("aria-checked", "true");
    expect(
      screen.getByRole("menuitemcheckbox", { name: "A", exact: true }),
    ).toHaveAttribute("aria-checked", "false");
    expect(
      screen.getByRole("menuitemcheckbox", { name: "C", exact: true }),
    ).toHaveAttribute("aria-checked", "true");
    expect(onMultiSelect).toHaveBeenCalledTimes(1);
  });

  it("keeps single-select emit and close behavior", async () => {
    const onMultiSelect = vi.fn();
    render(Harness, {
      selectedItems: ["a"],
      onMultiSelect,
      allowMultiSelect: false,
    });
    await fireEvent.keyDown(
      screen.getByRole("button", { name: "Choose fields" }),
      { key: "Enter" },
    );
    await fireEvent.click(
      await screen.findByRole("menuitem", { name: "B", exact: true }),
    );
    expect(onMultiSelect).toHaveBeenCalledExactlyOnceWith(["b"]);
    expect(screen.queryByRole("menu")).not.toBeInTheDocument();
  });
});
