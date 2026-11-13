import { cleanup, fireEvent, render, screen } from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";
import BreadcrumbItem from "./BreadcrumbItem.svelte";
import type { PathOption } from "./types";

vi.mock("$app/stores", async () => {
  const { readable } = await import("svelte/store");
  return {
    page: readable({
      url: new URL("http://localhost/org/project/explore/bids"),
    }),
  };
});
vi.mock("$app/state", () => ({
  page: { route: { id: "/[organization]/[project]/explore/[dashboard]" } },
}));

afterEach(cleanup);

const currentPath = ["org", "project", "bids"];

const dashboards = new Map<string, PathOption>([
  ["bids", { label: "Bids", section: "explore" }],
  ["auction", { label: "Auction", section: "explore" }],
]);

describe("BreadcrumbItem", () => {
  it("opens the dropdown when the current page's label is clicked", async () => {
    render(BreadcrumbItem, {
      pathOptions: { options: dashboards },
      current: "bids",
      currentPath,
      depth: 2,
      isCurrentPage: true,
    });

    const crumb = screen.getByRole("button", { name: "Bids" });
    expect(crumb).toHaveAttribute("aria-current", "page");
    expect(screen.queryByRole("link")).toBeNull();

    // bits-ui opens menus on a primary-button pointerdown, not on click.
    // jsdom has no PointerEvent, so dispatch the pointerdown as a MouseEvent.
    await fireEvent(
      screen.getByText("Bids"),
      new MouseEvent("pointerdown", { bubbles: true, button: 0 }),
    );

    expect(
      await screen.findByRole("menuitemcheckbox", { name: "Auction" }),
    ).toBeInTheDocument();
  });

  it("renders the current page's label as plain text when there is nothing to switch to", () => {
    render(BreadcrumbItem, {
      pathOptions: {
        options: new Map([["bids", { label: "Bids", section: "explore" }]]),
      },
      current: "bids",
      currentPath,
      depth: 2,
      isCurrentPage: true,
    });

    expect(screen.getByText("Bids")).toHaveAttribute("aria-current", "page");
    expect(screen.queryByRole("button")).toBeNull();
    expect(screen.queryByRole("link")).toBeNull();
  });

  it("keeps a link and a separate dropdown on crumbs above the current page", () => {
    render(BreadcrumbItem, {
      pathOptions: {
        options: new Map<string, PathOption>([
          ["project", { label: "Project" }],
          ["other", { label: "Other" }],
        ]),
      },
      current: "project",
      currentPath,
      depth: 1,
      isCurrentPage: false,
    });

    expect(screen.getByRole("link", { name: "Project" })).toHaveAttribute(
      "href",
      "/org/project",
    );
    expect(
      screen.getByRole("button", { name: "Breadcrumb dropdown" }),
    ).not.toHaveAttribute("aria-current");
  });
});
