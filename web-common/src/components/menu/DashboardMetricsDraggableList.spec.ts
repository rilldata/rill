import { fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import { afterAll, beforeAll, describe, expect, it, vi } from "vitest";
import { mockAnimationsForComponentTesting } from "@rilldata/web-common/lib/test/mock-animations";
import { EmbedStore } from "@rilldata/web-common/features/embeds/embed-store";
import DashboardMetricsPickerHarness from "./__fixtures__/DashboardMetricsPickerHarness.svelte";

async function openPicker(type: "measure" | "dimension" = "measure") {
  await fireEvent.click(
    screen.getByRole("button", {
      name:
        type === "measure"
          ? "Choose measures to display"
          : "Choose dimensions to display",
    }),
  );
  const search = await screen.findByRole("textbox", { name: "Search list" });
  search.focus();
  return search;
}

describe("Dashboard picker keyboard navigation", () => {
  mockAnimationsForComponentTesting();
  afterAll(() => vi.unstubAllGlobals());
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

  it.each(["measure", "dimension"] as const)(
    "navigates %s items and tags with arrows",
    async (type) => {
      render(DashboardMetricsPickerHarness, { props: { type } });
      const search = await openPicker(type);
      await fireEvent.keyDown(search, { key: "ArrowDown" });
      const requests = screen.getByRole("button", { name: "Hide Requests" });
      expect(requests).toHaveFocus();
      await fireEvent.keyDown(requests, { key: "ArrowDown" });
      const spend = screen.getByRole("button", { name: "Hide Spend" });
      expect(spend).toHaveFocus();
      await fireEvent.keyDown(spend, { key: "ArrowLeft" });
      const revenue = screen.getByRole("button", { name: "Filter by Revenue" });
      expect(revenue).toHaveFocus();
      await fireEvent.keyDown(revenue, { key: "ArrowUp" });
      const traffic = screen.getByRole("button", { name: "Filter by Traffic" });
      expect(traffic).toHaveFocus();
      await fireEvent.keyDown(traffic, { key: "ArrowRight" });
      expect(requests).toHaveFocus();
      await fireEvent.keyDown(requests, { key: "ArrowUp" });
      expect(search).toHaveFocus();
    },
  );

  it("toggles a tag filter with Enter and Space", async () => {
    render(DashboardMetricsPickerHarness);
    const search = await openPicker();
    await fireEvent.keyDown(search, { key: "ArrowDown" });
    await fireEvent.keyDown(
      screen.getByRole("button", { name: "Hide Requests" }),
      { key: "ArrowLeft" },
    );
    const traffic = screen.getByRole("button", { name: "Filter by Traffic" });
    await fireEvent.keyDown(traffic, { key: "Enter" });
    const selectedTag = screen.getByRole("button", {
      name: "Clear filter Traffic",
    });
    expect(selectedTag).toHaveFocus();
    expect(selectedTag).toHaveAttribute("aria-pressed", "true");
    await fireEvent.keyDown(selectedTag, { key: " ", code: "Space" });
    expect(
      screen.getByRole("button", { name: "Filter by Traffic" }),
    ).toHaveFocus();
  });

  it("leaves search focused when neither column has matching results", async () => {
    render(DashboardMetricsPickerHarness);
    const search = await openPicker();
    await fireEvent.input(search, { target: { value: "no matching item" } });
    await fireEvent.keyDown(search, { key: "ArrowDown" });
    expect(search).toHaveFocus();
  });

  it("keeps focus on the next shown row and reaches hidden rows with Right", async () => {
    const onChange = vi.fn();
    render(DashboardMetricsPickerHarness, { props: { onChange } });
    const search = await openPicker();
    await fireEvent.keyDown(search, { key: "ArrowDown" });
    await fireEvent.keyDown(
      screen.getByRole("button", { name: "Hide Requests" }),
      { key: "Enter" },
    );
    await waitFor(() =>
      expect(screen.getByRole("button", { name: "Hide Spend" })).toHaveFocus(),
    );
    expect(onChange).toHaveBeenLastCalledWith(["spend", "bids"]);
    await fireEvent.keyDown(
      screen.getByRole("button", { name: "Hide Spend" }),
      { key: "ArrowRight" },
    );
    expect(screen.getByRole("button", { name: "Show Requests" })).toHaveFocus();
    await fireEvent.keyDown(
      screen.getByRole("button", { name: "Show Requests" }),
      { key: " ", code: "Space" },
    );
    await waitFor(() =>
      expect(
        screen.getByRole("button", { name: "Hide Requests" }),
      ).toHaveFocus(),
    );
    expect(onChange).toHaveBeenLastCalledWith(["spend", "bids", "requests"]);
  });

  it.each(["measure", "dimension"] as const)(
    "continues through %s rows after toggling and switches sections with Tab",
    async (type) => {
      render(DashboardMetricsPickerHarness, { props: { type } });
      const search = await openPicker(type);
      await fireEvent.keyDown(search, { key: "ArrowDown" });
      await fireEvent.keyDown(
        screen.getByRole("button", { name: "Hide Requests" }),
        { key: "Enter" },
      );
      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: "Hide Spend" }),
        ).toHaveFocus(),
      );
      await fireEvent.keyDown(
        screen.getByRole("button", { name: "Hide Spend" }),
        { key: " ", code: "Space" },
      );
      await waitFor(() =>
        expect(screen.getByRole("button", { name: "Hide Bids" })).toHaveFocus(),
      );
      await fireEvent.keyDown(
        screen.getByRole("button", { name: "Hide Bids" }),
        { key: "Tab" },
      );
      const requests = screen.getByRole("button", { name: "Show Requests" });
      expect(requests).toHaveFocus();
      await fireEvent.keyDown(requests, { key: "Enter" });
      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: "Show Spend" }),
        ).toHaveFocus(),
      );
      await fireEvent.keyDown(
        screen.getByRole("button", { name: "Show Spend" }),
        { key: "ArrowLeft" },
      );
      expect(screen.getByRole("button", { name: "Hide Bids" })).toHaveFocus();
    },
  );

  it("enters filtered results and keeps the last visible item selected", async () => {
    const onChange = vi.fn();
    render(DashboardMetricsPickerHarness, {
      props: { selectedItems: ["requests"], withTags: false, onChange },
    });
    const search = await openPicker();
    await fireEvent.keyDown(search, { key: "ArrowDown" });
    const requests = screen.getByRole("button", { name: "Hide Requests" });
    expect(requests).toHaveFocus();
    await fireEvent.keyDown(requests, { key: "Enter" });
    expect(onChange).not.toHaveBeenCalled();
    search.focus();
    await fireEvent.input(search, { target: { value: "Spend" } });
    await fireEvent.keyDown(search, { key: "ArrowDown" });
    expect(screen.getByRole("button", { name: "Show Spend" })).toHaveFocus();
  });
  it.each([
    "navigation=true",
    "navigation=true&hide_navigation_bar=true",
    "navigation=false",
  ])(
    "supports dimension picker navigation in an embed with %s",
    async (params) => {
      EmbedStore.init(
        new URL(
          `http://localhost/-/embed/explore/auction?instance_id=local&runtime_host=http%3A%2F%2Flocalhost&access_token=test&${params}`,
        ),
      );
      render(DashboardMetricsPickerHarness, { props: { type: "dimension" } });
      await fireEvent.keyDown(window, { key: "." });
      const search = await screen.findByRole("textbox", {
        name: "Search list",
      });
      await waitFor(() => expect(search).toHaveFocus());
      await fireEvent.keyDown(search, { key: "ArrowDown" });
      const requests = screen.getByRole("button", { name: "Hide Requests" });
      expect(requests).toHaveFocus();
      await fireEvent.keyDown(requests, { key: "ArrowLeft" });
      const traffic = screen.getByRole("button", { name: "Filter by Traffic" });
      expect(traffic).toHaveFocus();
      await fireEvent.keyDown(traffic, { key: " ", code: "Space" });
      const activeTag = screen.getByRole("button", {
        name: "Clear filter Traffic",
      });
      expect(activeTag).toHaveFocus();
      await fireEvent.keyDown(activeTag, { key: "Enter" });
      await fireEvent.keyDown(
        screen.getByRole("button", { name: "Filter by Traffic" }),
        { key: "ArrowRight" },
      );
      const currentRequests = screen.getByRole("button", {
        name: "Hide Requests",
      });
      expect(currentRequests).toHaveFocus();
      await fireEvent.keyDown(currentRequests, { key: "Enter" });
      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: "Hide Spend" }),
        ).toHaveFocus(),
      );
      await fireEvent.keyDown(
        screen.getByRole("button", { name: "Hide Spend" }),
        { key: " ", code: "Space" },
      );
      await waitFor(() =>
        expect(screen.getByRole("button", { name: "Hide Bids" })).toHaveFocus(),
      );
      await fireEvent.keyDown(
        screen.getByRole("button", { name: "Hide Bids" }),
        { key: "ArrowRight" },
      );
      const hiddenRequests = screen.getByRole("button", {
        name: "Show Requests",
      });
      expect(hiddenRequests).toHaveFocus();
      await fireEvent.keyDown(hiddenRequests, { key: " ", code: "Space" });
      await waitFor(() =>
        expect(
          screen.getByRole("button", { name: "Show Spend" }),
        ).toHaveFocus(),
      );
      await fireEvent.keyDown(
        screen.getByRole("button", { name: "Show Spend" }),
        { key: "Tab" },
      );
      expect(screen.getByRole("button", { name: "Hide Bids" })).toHaveFocus();
    },
  );
});
