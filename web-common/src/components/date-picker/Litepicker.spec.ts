import Litepicker from "./Litepicker.svelte";
import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, waitFor, fireEvent } from "@testing-library/svelte";

// TODO: jsdom isn't firing our focus event listener in `onOpenMount`

// Utility to get the day element from the Litepicker UI
function getDayElement(container, dateString) {
  return container.querySelector(
    `[data-time="${new Date(dateString).valueOf()}"]`
  );
}

describe("Litepicker", () => {
  let startEl, endEl;
  beforeEach(() => {
    startEl = document.createElement("input");
    endEl = document.createElement("input");
  });

  afterEach(() => {
    startEl.remove();
    endEl.remove();
  });

  it("autofills the inputs with the default values", async () => {
    const { unmount } = render(Litepicker, {
      startEl,
      endEl,
      defaultStart: "1/1/2023",
      defaultEnd: "1/15/2023",
      openOnMount: true,
    });

    await waitFor(() => {
      expect(startEl.value).toBe("1/1/2023");
      expect(endEl.value).toBe("1/15/2023");
    });

    unmount();
  });

  it("calls change in reaction to day click events", async () => {
    const { container, component, unmount } = render(Litepicker, {
      startEl,
      endEl,
      defaultStart: "1/1/2023",
      defaultEnd: "1/15/2023",
      openOnMount: true,
    });

    let dates = null;

    component.$on("change", (e) => {
      dates = e.detail;
    });

    const altDay = getDayElement(container, "1/2/2023");
    await fireEvent.click(altDay);

    expect(dates).toEqual({
      start: new Date("1/2/2023"),
      end: new Date("1/15/2023"),
    });

    const altDay2 = getDayElement(container, "1/17/2023");
    await fireEvent.click(altDay2);

    expect(dates).toEqual({
      start: new Date("1/2/2023"),
      end: new Date("1/17/2023"),
    });

    unmount();
  });

  it("calls change in reaction to input change events", async () => {
    const { component, unmount } = render(Litepicker, {
      startEl,
      endEl,
      defaultStart: "1/1/2023",
      defaultEnd: "1/15/2023",
      openOnMount: true,
    });

    let dates = null;

    component.$on("change", (e) => {
      dates = e.detail;
    });

    await fireEvent.change(startEl, { target: { value: "1/2/2023" } });
    await fireEvent.blur(startEl);

    expect(dates).toEqual({
      start: new Date("1/2/2023"),
      end: new Date("1/15/2023"),
    });

    await fireEvent.change(endEl, { target: { value: "1/17/2023" } });
    await fireEvent.blur(endEl);

    expect(dates).toEqual({
      start: new Date("1/2/2023"),
      end: new Date("1/17/2023"),
    });

    unmount();
  });

  it("calls editing event when the date being edited is changed", async () => {
    const { container, component, unmount } = render(Litepicker, {
      startEl,
      endEl,
      defaultStart: "1/1/2023",
      defaultEnd: "1/15/2023",
      openOnMount: true,
    });

    let editing = null;

    component.$on("editing", (e) => {
      editing = e.detail;
    });

    // Focus the endEl to switch to editing the end date
    fireEvent.focus(endEl);

    await waitFor(() => {
      expect(editing).toBe(1);
    });

    // Focus the startEl to switch to editing the start date
    fireEvent.focus(startEl);

    await waitFor(() => {
      expect(editing).toBe(0);
    });

    // Select a start date to switch back to editing the end date
    const altDay = getDayElement(container, "1/2/2023");
    await fireEvent.click(altDay);
    await waitFor(() => {
      expect(editing).toBe(1);
    });

    // Select an end date to switch back to editing the start date
    const altDay2 = getDayElement(container, "1/17/2023");
    await fireEvent.click(altDay2);
    await waitFor(() => {
      expect(editing).toBe(0);
    });

    unmount();
  });

  it("renders by default the starting date", () => {
    const { container, unmount } = render(Litepicker, {
      startEl,
      endEl,
      defaultStart: "1/1/2023",
      defaultEnd: "1/15/2023",
      openOnMount: true,
    });

    const startDate = container.querySelector(
      `[data-time="${new Date("1/1/2023").valueOf()}"]`
    );
    expect(startDate).toBeDefined();

    unmount();
  });

  it("resets the range to 1 day if clicking an end date before start, or a start date after end date", async () => {
    const { container, component, unmount } = render(Litepicker, {
      startEl,
      endEl,
      defaultStart: "1/1/2023",
      defaultEnd: "1/15/2023",
      openOnMount: true,
    });

    let dates = null;

    component.$on("change", (e) => {
      dates = e.detail;
    });

    const altDay = getDayElement(container, "1/17/2023");
    await fireEvent.click(altDay);

    expect(dates).toEqual({
      start: new Date("1/17/2023"),
      end: new Date("1/17/2023"),
    });

    const altDay2 = getDayElement(container, "1/11/2023");
    await fireEvent.click(altDay2);

    expect(dates).toEqual({
      start: new Date("1/11/2023"),
      end: new Date("1/11/2023"),
    });

    unmount();
  });

  it("calls on:toggle when the picker shows/hides", async () => {
    const { component, unmount } = render(Litepicker, {
      startEl,
      endEl,
      defaultStart: "1/1/2023",
      defaultEnd: "1/15/2023",
    });

    let showState = null;

    component.$on("toggle", (e) => {
      showState = e.detail;
    });

    await fireEvent.focus(startEl);
    expect(showState).toBe(true);
    await fireEvent.click(document.body);
    expect(showState).toBe(false);

    unmount();
  });
});
