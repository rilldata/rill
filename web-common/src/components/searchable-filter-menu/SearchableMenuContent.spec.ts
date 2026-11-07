import { fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import { describe, expect, it, vi } from "vitest";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import SearchableMenuHarness from "./__fixtures__/SearchableMenuHarness.svelte";

describe("Searchable menu keyboard navigation", () => {
  mockAnimationsForComponentTesting();

  it("moves from search into the results and selects with Enter", async () => {
    const onSelect = vi.fn();
    render(SearchableMenuHarness, { props: { onSelect } });
    const search = screen.getByRole("textbox");
    await waitFor(() => expect(screen.getByRole("menu")).toHaveFocus());
    search.focus();
    await fireEvent.keyDown(search, { key: "ArrowDown" });
    expect(screen.getByRole("menuitem", { name: "Country" })).toHaveFocus();
    await fireEvent.keyDown(screen.getByRole("menuitem", { name: "Country" }), {
      key: "ArrowDown",
    });
    const device = screen.getByRole("menuitem", { name: "Device" });
    expect(device).toHaveFocus();
    await fireEvent.keyDown(device, { key: "Enter" });
    expect(onSelect).toHaveBeenCalledWith("device");
  });

  it("selects a focused filter with Space instead of opening the cell inspector", async () => {
    const onSelect = vi.fn();
    render(SearchableMenuHarness, { props: { onSelect } });
    const search = screen.getByRole("textbox");
    await waitFor(() => expect(screen.getByRole("menu")).toHaveFocus());
    search.focus();
    await fireEvent.keyDown(search, { key: "ArrowDown" });
    const country = screen.getByRole("menuitem", { name: "Country" });
    expect(country).toHaveFocus();
    await fireEvent.keyDown(country, { key: " ", code: "Space" });
    expect(onSelect).toHaveBeenCalledWith("country");
  });

  it("closes the containing menu when Escape leaves search", async () => {
    render(SearchableMenuHarness, { props: { onSelect: vi.fn() } });
    const search = screen.getByRole("textbox");
    await waitFor(() => expect(screen.getByRole("menu")).toHaveFocus());
    search.focus();
    await fireEvent.keyDown(search, { key: "Escape", code: "Escape" });
    await waitFor(() =>
      expect(screen.queryByRole("menu")).not.toBeInTheDocument(),
    );
  });

  it("moves up from search to the last filtered result", async () => {
    render(SearchableMenuHarness, { props: { onSelect: vi.fn() } });
    const search = screen.getByRole("textbox");
    await waitFor(() => expect(screen.getByRole("menu")).toHaveFocus());
    search.focus();
    await fireEvent.input(search, { target: { value: "Device" } });
    await fireEvent.keyDown(search, { key: "ArrowUp" });
    expect(screen.getByRole("menuitem", { name: "Device" })).toHaveFocus();
  });
});
