import {
  cleanup,
  fireEvent,
  render,
  screen,
  waitFor,
} from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";
import EmbeddedDashboardNavigation from "./EmbeddedDashboardNavigation.svelte";
import DashboardShortcuts from "../dashboards/shortcuts/DashboardShortcuts.svelte";
import {
  DashboardShortcutAction,
  getDashboardShortcuts,
  performDashboardShortcut,
} from "../dashboards/shortcuts/dashboard-shortcuts";

vi.mock("$app/stores", async () => {
  const { readable } = await import("svelte/store");
  return {
    page: readable({
      url: new URL("http://localhost/-/embed/explore/auction"),
      route: { id: "/-/embed/explore/[name]" },
      params: { name: "auction" },
    }),
  };
});

const breadcrumbOptions = {
  options: new Map([
    ["auction", { label: "Auction", href: "/-/embed/explore/auction" }],
    ["overview", { label: "Overview", href: "/-/embed/canvas/overview" }],
  ]),
};

afterEach(cleanup);

describe("embedded dashboard navigation shortcuts", () => {
  it.each([false, true])(
    "opens the embedded dashboard list with two-tier navigation %s",
    async (twoTiered) => {
      render(EmbeddedDashboardNavigation, {
        breadcrumbOptions,
        currentResourceName: "auction",
        twoTiered,
      });
      render(DashboardShortcuts, { view: "canvas" });
      await fireEvent.keyDown(window, { key: "d" });
      await waitFor(() => expect(screen.getByRole("menu")).toBeVisible());
      expect(screen.getByText("Overview").closest("a")).toHaveAttribute(
        "href",
        "/-/embed/canvas/overview",
      );
    },
  );

  it("keeps home navigation inside the embed", async () => {
    render(EmbeddedDashboardNavigation, {
      breadcrumbOptions,
      currentResourceName: "auction",
    });
    const home = screen.getByRole("link", { name: "Home" });
    const click = vi.fn((event: Event) => event.preventDefault());
    home.addEventListener("click", click);
    render(DashboardShortcuts, { view: "pivot" });
    await fireEvent.keyDown(window, { key: "h" });
    expect(click).toHaveBeenCalledOnce();
    expect(home).toHaveAttribute("href", "/-/embed");
  });

  it.each(["canvas", "explore", "pivot"] as const)(
    "opens shortcut help in the embedded %s view",
    async (view) => {
      render(EmbeddedDashboardNavigation, {
        breadcrumbOptions,
        currentResourceName: "auction",
      });
      render(DashboardShortcuts, { view });
      await fireEvent.keyDown(window, { key: "?" });
      await waitFor(() => expect(screen.getByRole("dialog")).toBeVisible());
      const dialog = screen.getByRole("dialog");
      expect(dialog).toHaveTextContent("Canvas");
      expect(dialog).toHaveTextContent("Explore");
      expect(dialog).toHaveTextContent("Pivot");
      expect(dialog).toHaveTextContent("Go to home");
    },
  );

  it("retains keyboard navigation when only the navigation bar is hidden", async () => {
    render(EmbeddedDashboardNavigation, {
      breadcrumbOptions,
      currentResourceName: "auction",
      navigationBarEnabled: false,
    });
    expect(screen.getByRole("navigation")).toHaveClass("sr-only");
    expect(
      performDashboardShortcut(DashboardShortcutAction.OpenDashboardList),
    ).toBe(true);
    await waitFor(() => expect(screen.getByRole("menu")).toBeVisible());
  });

  it("does not expose navigation when embed navigation is disabled", () => {
    render(EmbeddedDashboardNavigation, {
      breadcrumbOptions,
      currentResourceName: "auction",
      navigationEnabled: false,
    });
    expect(performDashboardShortcut(DashboardShortcutAction.GoHome)).toBe(
      false,
    );
    expect(
      performDashboardShortcut(DashboardShortcutAction.OpenDashboardList),
    ).toBe(false);
    for (const view of ["canvas", "explore", "pivot"] as const) {
      expect(
        getDashboardShortcuts({ view }).map((shortcut) => shortcut.id),
      ).not.toContain("go-home");
      expect(
        getDashboardShortcuts({ view }).map((shortcut) => shortcut.id),
      ).not.toContain("open-dashboard-list");
    }
  });
});
