import { describe, expect, it, vi } from "vitest";
import {
  DASHBOARD_SHORTCUTS,
  DashboardShortcutAction,
  getDashboardShortcutAction,
  performDashboardShortcut,
} from "./dashboard-shortcuts";

function keyboardEvent(
  key: string,
  options: KeyboardEventInit = {},
  target: EventTarget = document.body,
) {
  const event = new KeyboardEvent("keydown", { key, ...options });
  Object.defineProperty(event, "target", { value: target });
  return event;
}

describe("dashboard shortcuts", () => {
  it("defines the dashboard menu shortcuts", () => {
    expect(DASHBOARD_SHORTCUTS.slice(0, 4)).toEqual([
      {
        action: DashboardShortcutAction.ToggleHelp,
        keys: ["?"],
        description: "Show/hide the keyboard-shortcuts menu",
      },
      {
        action: DashboardShortcutAction.OpenFilter,
        keys: ["/"],
        description: "Open the Filter menu",
      },
      {
        action: DashboardShortcutAction.OpenMetricPicker,
        keys: [","],
        description: "Open the metric picker",
      },
      {
        action: DashboardShortcutAction.OpenDimensionPicker,
        keys: ["."],
        description: "Open the dimension picker",
      },
    ]);
  });

  it.each([
    ["?", DashboardShortcutAction.ToggleHelp],
    ["/", DashboardShortcutAction.OpenFilter],
    [",", DashboardShortcutAction.OpenMetricPicker],
    [".", DashboardShortcutAction.OpenDimensionPicker],
  ])("maps %s to %s", (key, action) => {
    expect(getDashboardShortcutAction(keyboardEvent(key))).toBe(action);
  });

  it("ignores shortcuts while editing text", () => {
    for (const target of [
      document.createElement("input"),
      document.createElement("textarea"),
      document.createElement("select"),
    ]) {
      expect(getDashboardShortcutAction(keyboardEvent("/", {}, target))).toBe(
        undefined,
      );
    }

    const editable = document.createElement("div");
    editable.setAttribute("contenteditable", "true");
    expect(getDashboardShortcutAction(keyboardEvent("/", {}, editable))).toBe(
      undefined,
    );
  });

  it("ignores repeated and modified shortcuts", () => {
    expect(
      getDashboardShortcutAction(keyboardEvent("/", { repeat: true })),
    ).toBe(undefined);
    expect(
      getDashboardShortcutAction(keyboardEvent("/", { metaKey: true })),
    ).toBe(undefined);
    expect(
      getDashboardShortcutAction(keyboardEvent("/", { ctrlKey: true })),
    ).toBe(undefined);
    expect(
      getDashboardShortcutAction(keyboardEvent("/", { altKey: true })),
    ).toBe(undefined);
  });

  it("allows Shift for punctuation keys", () => {
    expect(
      getDashboardShortcutAction(keyboardEvent("?", { shiftKey: true })),
    ).toBe(DashboardShortcutAction.ToggleHelp);
  });

  it("clicks the element registered for an action", () => {
    const root = document.createElement("div");
    const target = document.createElement("button");
    const click = vi.fn();
    target.dataset.dashboardShortcut = DashboardShortcutAction.OpenFilter;
    target.addEventListener("click", click);
    root.appendChild(target);

    expect(
      performDashboardShortcut(DashboardShortcutAction.OpenFilter, root),
    ).toBe(true);
    expect(click).toHaveBeenCalledOnce();
    expect(
      performDashboardShortcut(
        DashboardShortcutAction.OpenDimensionPicker,
        root,
      ),
    ).toBe(false);
  });
});
