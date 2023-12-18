import RemovableListMenu from "./RemovableListMenu.svelte";
import { describe, it, expect, vi } from "vitest";
import { render, waitFor, fireEvent, screen } from "@testing-library/svelte";

describe("RemovableListMenu", () => {
  it("does not render selected values if not in all values", async () => {
    const { unmount } = render(RemovableListMenu, {
      excludeMode: false,
      selectedValues: ["x"],
      allValues: ["foo", "bar"],
    });

    const foo = screen.getByText("foo");
    const bar = screen.getByText("bar");
    expect(foo).toBeDefined();
    expect(bar).toBeDefined();

    const x = screen.queryByText("x");
    expect(x).toBeNull();

    unmount();
  });

  it("renders all values if search text is empty", async () => {
    const { unmount } = render(RemovableListMenu, {
      excludeMode: false,
      selectedValues: [],
      allValues: ["foo", "bar"],
    });

    const foo = screen.getByText("foo");
    const bar = screen.getByText("bar");
    expect(foo).toBeDefined();
    expect(bar).toBeDefined();
    unmount();
  });

  it("renders search values if search text is populated", async () => {
    const { unmount } = render(RemovableListMenu, {
      excludeMode: false,
      selectedValues: ["foo", "bar"],
      allValues: ["x"],
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
    const { unmount, component } = render(RemovableListMenu, {
      excludeMode: false,
      selectedValues: ["foo", "bar"],
      allValues: ["x"],
    });

    const switchInput = screen.getByText("Exclude");
    expect(switchInput).toBeDefined();

    await component.$set({ excludeMode: true });

    const includeButton = screen.getByText("Include");
    expect(includeButton).toBeDefined();

    unmount();
  });

  it("should dispatch toggle, apply, and search events", async () => {
    const { unmount, component } = render(RemovableListMenu, {
      excludeMode: false,
      selectedValues: [],
      allValues: ["foo", "bar"],
    });

    const toggleSpy = vi.fn();
    component.$on("toggle", toggleSpy);
    const switchInput = screen.getByText("Exclude");
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
