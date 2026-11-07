import {
  fireEvent,
  render,
  screen,
  waitFor,
  within,
} from "@testing-library/svelte";
import { beforeAll, describe, expect, it, vi } from "vitest";
import { get } from "svelte/store";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import { metricsExplorerStore } from "../stores/dashboard-stores";
import { AD_BIDS_EXPLORE_NAME } from "../stores/test-data/data";
import { ephemeralMeasureDialog } from "../ephemeral-measures/dialog-store";
import PivotSidebarHarness from "./__fixtures__/PivotSidebarHarness.svelte";

function state() {
  return get(metricsExplorerStore).entities[AD_BIDS_EXPLORE_NAME].pivot;
}
function focused() {
  return document.activeElement;
}
async function key(key: string, shiftKey = false) {
  const target = focused();
  if (!(target instanceof HTMLElement)) throw new Error("No focused field");
  await fireEvent.keyDown(target, {
    key,
    shiftKey,
    code: key === " " ? "Space" : undefined,
  });
}
async function start() {
  const search = screen.getByRole("textbox");
  search.focus();
  await key("ArrowDown");
  return search;
}

describe("Pivot sidebar keyboard navigation", () => {
  it("reaches calculated measure actions from the field", async () => {
    render(PivotSidebarHarness, { withEphemeral: true });
    await start();
    await key("Tab");

    const row = focused()?.closest<HTMLElement>("[data-pivot-field]");
    if (!row) throw new Error("No focused pivot field");
    const edit = within(row).getByRole("button", {
      name: "Edit adhoc measure",
    });
    const remove = within(row).getByRole("button", { name: "Delete" });
    const addColumn = within(row).getByRole("button", { name: "Add Column" });
    await key("ArrowRight");
    expect(edit).toHaveFocus();
    await key("ArrowRight");
    expect(remove).toHaveFocus();
    await key("ArrowRight");
    expect(addColumn).toHaveFocus();
    await key("ArrowLeft");
    expect(remove).toHaveFocus();
    await key("ArrowLeft");
    expect(edit).toHaveFocus();
    await key("Enter");

    expect(state().columns).toEqual([]);
    expect(get(ephemeralMeasureDialog)?.def?.name).toBe("requests");
    ephemeralMeasureDialog.set(null);
  });
  mockAnimationsForComponentTesting();
  beforeAll(() => {
    vi.stubGlobal(
      "ResizeObserver",
      class {
        observe() {}
        unobserve() {}
        disconnect() {}
      },
    );
    HTMLElement.prototype.scrollIntoView ??= () => {};
  });

  it("enters Time from search, cycles sections, and returns to search", async () => {
    render(PivotSidebarHarness);
    const search = await start();
    expect(focused()?.getAttribute("aria-label")).toContain("Time");
    await key("Tab");
    expect(focused()?.getAttribute("aria-label")).toContain("Requests");
    await key("Tab");
    expect(focused()?.getAttribute("aria-label")).toContain("Country");
    await key("Tab", true);
    expect(focused()?.getAttribute("aria-label")).toContain("Requests");
    await key("Escape");
    expect(search).toHaveFocus();
  });

  it.each(["Enter", " "])(
    "adds measures to columns with %s and keeps the next field focused",
    async (activation) => {
      render(PivotSidebarHarness);
      await start();
      await key("Tab");
      await key(activation);
      await waitFor(() =>
        expect(state().columns.map((item) => item.id)).toEqual(["requests"]),
      );
      expect(focused()?.getAttribute("aria-label")).toContain("Spend");
    },
  );

  it.each(["rows", "columns"] as const)(
    "lets the user place a dimension in %s",
    async (zone) => {
      render(PivotSidebarHarness);
      await start();
      await key("Tab");
      await key("Tab");
      await key("Enter");
      expect(focused()).toHaveAttribute("data-pivot-placement", "rows");
      if (zone === "columns") await key("ArrowRight");
      await key(" ");
      await waitFor(() =>
        expect(state()[zone].map((item) => item.id)).toEqual(["country"]),
      );
      expect(focused()?.getAttribute("aria-label")).toContain("Region");
    },
  );

  it.each(["rows", "columns"] as const)(
    "adds Time to %s with an available grain",
    async (zone) => {
      render(PivotSidebarHarness);
      await start();
      await key("Enter");
      if (zone === "columns") await key("ArrowRight");
      await key("Enter");
      expect(state()[zone][0]?.type).toBe("time");
      expect(state()[zone][0]?.id).not.toBe("time");
    },
  );

  it("cancels the placement choice without adding a dimension", async () => {
    render(PivotSidebarHarness);
    await start();
    await key("Tab");
    await key("Tab");
    await key("Enter");
    await key("Escape");
    expect(focused()?.getAttribute("aria-label")).toContain("Country");
    expect(state().rows).toEqual([]);
    expect(state().columns).toEqual([]);
  });

  it("uses columns directly in flat mode", async () => {
    render(PivotSidebarHarness, { flat: true });
    await start();
    await key("Tab");
    await key("Tab");
    await key("Enter");
    expect(state().rows).toEqual([]);
    expect(state().columns.map((item) => item.id)).toEqual(["country"]);
  });

  it("navigates and selects tags with arrows, Enter and Space", async () => {
    render(PivotSidebarHarness);
    await start();
    await key("ArrowLeft");
    expect(focused()).toHaveAttribute("data-pivot-tag", "Traffic");
    await key("ArrowDown");
    expect(focused()).toHaveAttribute("data-pivot-tag", "Location");
    await key("Enter");
    expect(focused()).toHaveAttribute("aria-pressed", "true");
    await key(" ");
    expect(focused()).toHaveAttribute("aria-pressed", "false");
    await key("ArrowRight");
    expect(focused()?.getAttribute("aria-label")).toContain("Time");
  });

  it("skips sections without matching fields", async () => {
    render(PivotSidebarHarness);
    const search = screen.getByRole("textbox");
    await fireEvent.input(search, { target: { value: "Region" } });
    await start();
    await key("Tab");
    expect(focused()?.getAttribute("aria-label")).toContain("Region");
  });
});
