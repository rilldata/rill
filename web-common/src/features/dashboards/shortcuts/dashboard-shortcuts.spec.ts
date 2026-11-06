import { setLocale } from "@rilldata/web-common/lib/i18n/gen/runtime";
import { afterEach, describe, expect, it, vi } from "vitest";
import {
  getDashboardShortcuts,
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
  it("shows Pivot navigation help only when its sidebar is available", () => {
    const root = document.createElement("div");
    expect(
      getDashboardShortcuts({ root, view: "pivot" }).some((item) =>
        item.id.startsWith("pivot-"),
      ),
    ).toBe(false);
    root.innerHTML =
      '<div data-dashboard-shortcut-context="pivot-sidebar"></div>';
    expect(
      getDashboardShortcuts({ root, view: "pivot" })
        .filter((item) => item.id.startsWith("pivot-"))
        .map((item) => item.id),
    ).toHaveLength(5);
    for (const view of ["canvas", "explore"] as const)
      expect(
        getDashboardShortcuts({ root, view }).some((item) =>
          item.id.startsWith("pivot-"),
        ),
      ).toBe(false);
  });
  it("does not advertise Explore-only shortcuts in Canvas or Pivot", () => {
    const root = document.createElement("div");
    root.innerHTML =
      '<div data-dashboard-shortcut-context="chart"></div><button data-dashboard-shortcut="open-metric-picker"></button><button data-dashboard-shortcut="open-dimension-picker"></button>';
    for (const view of ["canvas", "pivot"] as const) {
      const ids = getDashboardShortcuts({
        root,
        view,
        explainEnabled: true,
      }).map((shortcut) => shortcut.id);
      for (const id of [
        "open-metric-picker",
        "open-dimension-picker",
        "pan-left",
        "zoom",
        "explain",
      ])
        expect(ids).not.toContain(id);
    }
  });
  it("supports Canvas navigation without requiring Explore controls", () => {
    const root = document.createElement("div");
    for (const action of [
      DashboardShortcutAction.ToggleHelp,
      DashboardShortcutAction.GoHome,
      DashboardShortcutAction.OpenDashboardList,
      DashboardShortcutAction.ToggleComparison,
      DashboardShortcutAction.OpenFilter,
    ]) {
      const button = document.createElement("button");
      button.dataset.dashboardShortcut = action;
      root.append(button);
      expect(performDashboardShortcut(action, root)).toBe(true);
    }
    const ids = getDashboardShortcuts({ root, view: "canvas" }).map(
      (s) => s.id,
    );
    expect(ids).toEqual(
      expect.arrayContaining([
        "toggle-help",
        "go-home",
        "open-dashboard-list",
        "toggle-comparison",
        "open-filter",
      ]),
    );
    expect(ids).not.toContain("toggle-view");
    expect(ids).not.toContain("open-metric-picker");
    expect(ids).not.toContain("open-dimension-picker");
  });
  afterEach(() => setLocale("en", { reload: false }));

  it("omits help rows for absent pickers and chart contexts", () => {
    const ids = getDashboardShortcuts({
      root: document.createElement("div"),
    }).map((s) => s.id);
    for (const id of [
      "go-home",
      "open-filter",
      "open-metric-picker",
      "open-dimension-picker",
      "select-all",
      "pan-left",
      "pan-right",
      "zoom",
      "undo-zoom",
      "explain",
    ])
      expect(ids).not.toContain(id);
  });

  it("shows dimension and chart shortcuts only for their mounted contexts", () => {
    const root = document.createElement("div");
    root.innerHTML =
      '<div data-dashboard-shortcut-context="chart"></div><button data-dashboard-shortcut-context="dimension-selection"></button>';
    const ids = getDashboardShortcuts({ root }).map((s) => s.id);
    expect(ids).toContain("pan-left");
    expect(ids).toContain("pan-right");
    expect(ids).toContain("select-all");
    expect(ids).not.toContain("explain");
    expect(
      getDashboardShortcuts({ root, explainEnabled: true }).map((s) => s.id),
    ).toContain("explain");
  });

  it("reads the active locale when opening help", () => {
    setLocale("es", { reload: false });
    expect(
      getDashboardShortcuts({ root: document.createElement("div") }).find(
        (s) => s.id === "inspect-cell",
      )?.description,
    ).toBe("Mostrar u ocultar el inspector de celdas");
  });
  it("blocks picker shortcuts while help is open", () => {
    for (const key of ["/", ",", ".", "c", "p", "d", "h"]) {
      expect(
        getDashboardShortcutAction(keyboardEvent(key), { helpOpen: true }),
      ).toBeUndefined();
    }
    expect(
      getDashboardShortcutAction(keyboardEvent("?"), { helpOpen: true }),
    ).toBe(DashboardShortcutAction.ToggleHelp);
  });

  it.each(["dialog", "alertdialog"])(
    "blocks all dashboard shortcuts behind another %s",
    (role) => {
      const root = document.createElement("div");
      const modal = document.createElement("div");
      modal.setAttribute("role", role);
      modal.setAttribute("aria-modal", "true");
      root.append(modal);
      for (const key of ["?", "/", ",", ".", "c", "p", "d", "h"]) {
        expect(
          getDashboardShortcutAction(keyboardEvent(key), { root }),
        ).toBeUndefined();
      }
    },
  );
  it("does not block shortcuts for a closed mounted dialog", () => {
    const root = document.createElement("div");
    root.innerHTML =
      '<div role="dialog" aria-modal="true" data-state="closed"></div>';
    expect(getDashboardShortcutAction(keyboardEvent("/"), { root })).toBe(
      DashboardShortcutAction.OpenFilter,
    );
  });

  it("defines the dashboard menu shortcuts", () => {
    const root = document.createElement("div");
    for (const action of Object.values(DashboardShortcutAction)) {
      const button = document.createElement("button");
      button.dataset.dashboardShortcut = action;
      root.append(button);
    }
    expect(
      getDashboardShortcuts({ root }).map((shortcut) => shortcut.id),
    ).toEqual(expect.arrayContaining(Object.values(DashboardShortcutAction)));
    expect(getDashboardShortcuts({ root }).slice(0, 4)).toEqual(
      [
        {
          id: DashboardShortcutAction.ToggleHelp,
          keys: ["?"],
          description: "Show/hide the keyboard-shortcuts menu",
        },
        {
          id: DashboardShortcutAction.OpenFilter,
          keys: ["/"],
          description: "Open the Filter menu",
        },
        {
          id: DashboardShortcutAction.OpenMetricPicker,
          keys: [","],
          description: "Open the metric picker",
        },
        {
          id: DashboardShortcutAction.OpenDimensionPicker,
          keys: ["."],
          description: "Open the dimension picker",
        },
      ].map((shortcut, index) => ({
        ...shortcut,
        views: index < 2 ? ["canvas", "explore", "pivot"] : ["explore"],
      })),
    );
  });

  it.each([
    ["?", DashboardShortcutAction.ToggleHelp],
    ["/", DashboardShortcutAction.OpenFilter],
    [",", DashboardShortcutAction.OpenMetricPicker],
    [".", DashboardShortcutAction.OpenDimensionPicker],
    ["c", "toggle-comparison"],
    ["C", "toggle-comparison"],
    ["p", "toggle-view"],
    ["P", "toggle-view"],
    ["d", "open-dashboard-list"],
    ["D", "open-dashboard-list"],
    ["h", "go-home"],
    ["H", "go-home"],
  ])("maps %s to %s", (key, action) => {
    expect(getDashboardShortcutAction(keyboardEvent(key))).toBe(action);
  });

  it("ignores shortcuts while editing text", () => {
    for (const target of [
      document.createElement("input"),
      document.createElement("textarea"),
      document.createElement("select"),
    ]) {
      for (const key of ["/", "c", "p", "d", "h"]) {
        expect(
          getDashboardShortcutAction(keyboardEvent(key, {}, target)),
        ).toBeUndefined();
      }
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

  it.each(["disabled", "aria-disabled"])(
    "does not activate a %s comparison control",
    (attribute) => {
      const root = document.createElement("div");
      const target = document.createElement("button");
      target.dataset.dashboardShortcut = "toggle-comparison";
      target.setAttribute(attribute, attribute === "disabled" ? "" : "true");
      const click = vi.fn();
      target.addEventListener("click", click);
      root.append(target);
      expect(
        performDashboardShortcut(
          DashboardShortcutAction.ToggleComparison,
          root,
        ),
      ).toBe(false);
      expect(click).not.toHaveBeenCalled();
      expect(getDashboardShortcuts({ root }).map((s) => s.id)).not.toContain(
        "toggle-comparison",
      );
    },
  );

  it.each(Object.values(DashboardShortcutAction))(
    "clicks the element registered for %s",
    (action) => {
      const root = document.createElement("div");
      const target = document.createElement("button");
      const click = vi.fn();
      target.dataset.dashboardShortcut = action;
      target.addEventListener("click", click);
      root.appendChild(target);

      expect(performDashboardShortcut(action, root)).toBe(true);
      expect(click).toHaveBeenCalledOnce();
      expect(
        performDashboardShortcut(action, document.createElement("div")),
      ).toBe(false);
    },
  );
});
