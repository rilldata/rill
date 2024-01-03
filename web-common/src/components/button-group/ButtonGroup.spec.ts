import { fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import ButtonGroupTestingWrapper from "./ButtonGroupTestingWrapper.svelte";

describe("ButtonGroup", () => {
  it("ButtonGroupTestingWrapper -- buttons in test wrapper exist", async () => {
    const { unmount } = render(ButtonGroupTestingWrapper, {
      values: [1, 2, 3],
      selected: [1, 2, 3],
      disabled: [1, 2, 3],
    });

    [1, 2, 3].forEach(async (i) => {
      const button = screen.getByRole("button", { name: `button-${i}` });
      expect(button).toBeInstanceOf(HTMLButtonElement);
    });

    unmount();
  });

  it("button clicks dispatch correctly for non-disabled buttons", async () => {
    const onClick = vi.fn();

    const { component, unmount } = render(ButtonGroupTestingWrapper, {
      values: [1, 2, 3],
      selected: [1, 2, 3],
      disabled: [],
    });
    component.$on("subbutton-click", onClick);

    [1, 2, 3].forEach(async (i) => {
      const button = screen.getByRole("button", { name: `button-${i}` });
      await fireEvent.click(button);
      expect(onClick).toBeCalledWith(expect.objectContaining({ detail: i }));
    });

    expect(onClick).toBeCalledTimes(3);
    unmount();
  });

  it("no button clicks dispatch for disabled buttons", async () => {
    const onClick = vi.fn();

    const { component, unmount } = render(ButtonGroupTestingWrapper, {
      values: [1, 2, 3],
      selected: [1, 2, 3],
      disabled: [1, 2, 3],
    });
    component.$on("subbutton-click", onClick);

    [1, 2, 3].forEach(async (i) => {
      const button = screen.getByRole("button", { name: `button-${i}` });
      await fireEvent.click(button);
    });

    expect(onClick).toBeCalledTimes(0);
    unmount();
  });

  it("correct tooltips", async () => {
    const { unmount } = render(ButtonGroupTestingWrapper, {
      values: [1, 2, 3],
      selected: [1],
      disabled: [3],
    });

    [
      [1, "selected tt"],
      [2, "unselected tt"],
      [3, "disabled tt"],
    ].forEach(async ([i, tt]) => {
      const button = screen.getByRole("button", { name: `button-${i}` });
      await fireEvent.mouseEnter(button.parentElement);
      const toolTip = await waitFor(() => screen.getByText(tt));
      expect(toolTip).toBeTruthy();
      await fireEvent.mouseLeave(button.parentElement);
    });

    unmount();
  });
});

describe("ButtonGroup - adding buttons", () => {
  let component;
  let unmount;
  let onClick;

  beforeEach(() => {
    const { component: component_before, unmount: unmount_before } = render(
      ButtonGroupTestingWrapper,
      {
        values: [1, 2, 3],
        selected: [1, 2, 3],
        disabled: [],
      }
    );
    component = component_before;
    unmount = unmount_before;
    onClick = vi.fn();
    component.$on("subbutton-click", onClick);
  });

  afterEach(() => {
    unmount();
  });

  it("added buttons found", async () => {
    component.$set({ values: [1, 2, 3, 4] });
    const button = await screen.findByRole("button", { name: `button-${4}` });
    expect(button).toBeInstanceOf(HTMLButtonElement);
  });

  it("added buttons clickable if not disabled", async () => {
    component.$set({ values: [1, 2, 3, 4] });
    const button = await screen.findByRole("button", { name: `button-${4}` });
    await fireEvent.click(button);
    expect(onClick).toBeCalledWith(expect.objectContaining({ detail: 4 }));
  });

  it("added buttons not clickable if disabled", async () => {
    component.$set({ values: [1, 2, 3, 4], disabled: [4] });
    const button = await screen.findByRole("button", { name: `button-${4}` });
    await fireEvent.click(button);
    expect(onClick).toBeCalledTimes(0);
  });

  it("added has correct tooltip, including on props change", async () => {
    // window.scrollTo = vi.fn();
    component.$set({ values: [1, 2, 3, 4] });

    const button = await screen.findByRole("button", { name: `button-${4}` });

    // mock console.error to avoid irrelevant errors about
    // `scrollTo` not being implemented in jsdom
    const errorObject = console.error;
    console.error = vi.fn();

    await fireEvent.mouseEnter(button.parentElement);
    let toolTip = await screen.findByText("unselected tt");
    expect(toolTip).toBeTruthy();
    await fireEvent.mouseLeave(button.parentElement);

    component.$set({ values: [1, 2, 3, 4], selected: [4] });
    await fireEvent.mouseEnter(button.parentElement);
    toolTip = await waitFor(() => screen.getByText("selected tt"));
    expect(toolTip).toBeTruthy();
    await fireEvent.mouseLeave(button.parentElement);

    component.$set({ values: [1, 2, 3, 4], disabled: [4] });
    await fireEvent.mouseEnter(button.parentElement);
    toolTip = await waitFor(() => screen.getByText("disabled tt"));
    expect(toolTip).toBeTruthy();
    await fireEvent.mouseLeave(button.parentElement);

    // unmock console.error
    console.error = errorObject;
  });
});

describe("ButtonGroup - removing buttons", () => {
  let component;
  let unmount;
  let onClick;

  beforeEach(() => {
    const { component: component_before, unmount: unmount_before } = render(
      ButtonGroupTestingWrapper,
      {
        values: [1, 2, 3, 4, 5],
        selected: [1, 2],
        disabled: [4],
      }
    );
    component = component_before;
    unmount = unmount_before;
    onClick = vi.fn();
    component.$on("subbutton-click", onClick);
  });

  afterEach(() => {
    unmount();
  });

  it("removed buttons not found", async () => {
    await component.$set({ values: [1, 3, 4] });
    const button2 = await screen.queryByRole("button", { name: `button-2` });
    expect(button2).toBeNull();
    const button5 = await screen.queryByRole("button", { name: `button-5` });
    expect(button5).toBeNull();
  });

  it("after removal, remaining buttons dispatch correct actions", async () => {
    await component.$set({ values: [1, 3, 4] });

    let button = screen.getByRole("button", { name: `button-1` });
    await fireEvent.click(button);
    expect(onClick).toBeCalledWith(expect.objectContaining({ detail: 1 }));

    button = screen.getByRole("button", { name: `button-3` });
    await fireEvent.click(button);
    expect(onClick).toBeCalledWith(expect.objectContaining({ detail: 1 }));

    button = screen.getByRole("button", { name: `button-4` });
    // disabled, should not be clickable
    await fireEvent.click(button);

    expect(onClick).toBeCalledTimes(2);
  });

  it("after removal, remaining buttons have correct tooltips", async () => {
    await component.$set({ values: [1, 3, 4] });

    // mock console.error to avoid irrelevant errors about
    // `scrollTo` not being implemented in jsdom
    const errorObject = console.error;
    console.error = vi.fn();

    let button = await screen.findByRole("button", { name: `button-1` });
    await fireEvent.mouseEnter(button.parentElement);
    let toolTip = await screen.findByText("selected tt");
    expect(toolTip).toBeTruthy();
    await fireEvent.mouseLeave(button.parentElement);

    button = await screen.findByRole("button", { name: `button-3` });
    await fireEvent.mouseEnter(button.parentElement);
    toolTip = await waitFor(() => screen.getByText("unselected tt"));
    expect(toolTip).toBeTruthy();
    await fireEvent.mouseLeave(button.parentElement);

    button = await screen.findByRole("button", { name: `button-4` });
    await fireEvent.mouseEnter(button.parentElement);
    toolTip = await waitFor(() => screen.getByText("disabled tt"));
    expect(toolTip).toBeTruthy();
    await fireEvent.mouseLeave(button.parentElement);

    // unmock console.error
    console.error = errorObject;
  });
});
