import RemovableListMenu from "./RemovableListMenu.svelte";
import { describe, it, expect, vi } from "vitest";
import { render, waitFor, fireEvent, screen } from "@testing-library/svelte";
import { writable } from "svelte/store";

describe("RemovableListMenu", () => {
  it("renders selected values by default", async () => {
    const { unmount } = render(RemovableListMenu, {
      excludeStore: writable(false),
      selectedValues: ["foo", "bar"],
      searchedValues: null,
    });

    const foo = screen.getByText("foo");
    const bar = screen.getByText("bar");
    expect(foo).toBeDefined();
    expect(bar).toBeDefined();
    unmount();
  });

  it("renders selected values if search text is empty", async () => {
    const { unmount } = render(RemovableListMenu, {
      excludeStore: writable(false),
      selectedValues: ["foo", "bar"],
      searchedValues: ["x"],
    });

    const foo = screen.getByText("foo");
    const bar = screen.getByText("bar");
    expect(foo).toBeDefined();
    expect(bar).toBeDefined();
    unmount();
  });

  it("renders search values if search text is populated", async () => {
    const { unmount } = render(RemovableListMenu, {
      excludeStore: writable(false),
      selectedValues: ["foo", "bar"],
      searchedValues: ["x"],
    });

    const searchInput = screen.getByRole("textbox", { name: "Search list" });
    await fireEvent.input(searchInput, { target: { value: "x" } });

    const x = screen.getByText("x");
    const foo = screen.queryByText("foo");
    expect(x).toBeDefined();
    expect(foo).toBeNull();

    unmount();
  });

  it("should render switch based on exclude store", async () => {
    const excludeStore = writable(false);
    const { unmount } = render(RemovableListMenu, {
      excludeStore,
      selectedValues: ["foo", "bar"],
      searchedValues: ["x"],
    });

    const switchInput = screen.getByRole<HTMLInputElement>("switch");
    expect(switchInput.checked).toBe(false);

    excludeStore.set(true);
    await waitFor(() => {
      expect(switchInput.checked).toBe(true);
    });

    unmount();
  });

  it("should dispatch toggle, apply, and search events", async () => {
    const excludeStore = writable(false);
    const { unmount, component } = render(RemovableListMenu, {
      excludeStore,
      selectedValues: ["foo", "bar"],
      searchedValues: ["x"],
    });

    const toggleSpy = vi.fn();
    component.$on("toggle", toggleSpy);
    const switchInput = screen.getByRole<HTMLInputElement>("switch");
    await fireEvent.click(switchInput);
    expect(toggleSpy).toHaveBeenCalledOnce();

    const applySpy = vi.fn();
    component.$on("apply", (e) => applySpy(e.detail));
    const applyButton = screen.getByRole("menuitem", { name: "foo" });
    await fireEvent.click(applyButton);
    await waitFor(() => expect(applySpy).toHaveBeenCalledWith("foo"));

    const searchSpy = vi.fn();
    component.$on("search", (e) => searchSpy(e.detail));
    const searchInput = screen.getByRole("textbox", { name: "Search list" });
    await fireEvent.input(searchInput, { target: { value: "x" } });
    expect(searchSpy).toHaveBeenCalledWith("x");

    unmount();
  });
});
