import { render, screen } from "@testing-library/svelte";
import { afterEach, describe, expect, it, vi } from "vitest";
import InfiniteScrollTable from "./InfiniteScrollTable.svelte";

const props = {
  data: [{ email: "existing@example.com" }],
  columns: [{ accessorKey: "email", header: "Email" }],
  hasNextPage: false,
  isFetchingNextPage: false,
  onLoadMore: vi.fn(),
};

afterEach(() => vi.unstubAllGlobals());

describe("InfiniteScrollTable loading", () => {
  it("keeps existing rows and the table mounted while loading", async () => {
    const { rerender, container } = render(InfiniteScrollTable, props);
    const table = screen.getByRole("table");
    expect(screen.getByText("existing@example.com")).toBeInTheDocument();
    await rerender({ ...props, isFetchingNextPage: true });
    expect(screen.getByRole("table")).toBe(table);
    expect(screen.getByText("existing@example.com")).toBeInTheDocument();
    expect(container.querySelector('[aria-busy="true"]')).not.toBeNull();
  });

  it("shows no-results only after loading finishes", async () => {
    const { rerender } = render(InfiniteScrollTable, {
      ...props,
      data: [],
      isFetchingNextPage: true,
    });
    expect(screen.queryByText("No items found")).not.toBeInTheDocument();
    await rerender({ ...props, data: [] });
    expect(screen.getByText("No items found")).toBeInTheDocument();
  });

  it("continues loading when a page has no visible matches", async () => {
    const callbacks: IntersectionObserverCallback[] = [];
    const observe = vi.fn();
    vi.stubGlobal(
      "IntersectionObserver",
      class {
        observe = observe;
        disconnect = vi.fn();
        constructor(callback: IntersectionObserverCallback) {
          callbacks.push(callback);
        }
      },
    );
    const onLoadMore = vi.fn();
    const pagedProps = { ...props, data: [], hasNextPage: true, onLoadMore };
    const { rerender } = render(InfiniteScrollTable, pagedProps);
    expect(observe).toHaveBeenCalledTimes(1);
    const intersect = () =>
      callbacks.at(-1)!(
        [{ isIntersecting: true } as IntersectionObserverEntry],
        {} as IntersectionObserver,
      );
    intersect();
    expect(onLoadMore).toHaveBeenCalledTimes(1);
    await rerender({ ...pagedProps, isFetchingNextPage: true });
    intersect();
    expect(onLoadMore).toHaveBeenCalledTimes(1);
    await rerender(pagedProps);
    expect(observe).toHaveBeenCalledTimes(2);
    intersect();
    expect(onLoadMore).toHaveBeenCalledTimes(2);
    expect(screen.queryByText("No items found")).not.toBeInTheDocument();
  });
});
