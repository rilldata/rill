import { fireEvent, render, screen, waitFor } from "@testing-library/svelte";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import ButtonGroupTestingWrapper from "./ButtonGroupTestingWrapper.svelte";

// Helper function to simulate hover with intent
async function hoverWithIntent(element: HTMLElement) {
  // Initial mouse enter
  await fireEvent.mouseEnter(element);
  // Small movement to simulate intent
  await fireEvent.mouseMove(element, {
    clientX: 10,
    clientY: 10,
  });
  // Wait for hover intent delay
  await new Promise((resolve) => setTimeout(resolve, 300)); // activeDelay (200) + timeout (100)
}

// Helper function to wait for tooltip
async function waitForTooltip(text: string) {
  return waitFor(
    () => {
      const tooltip = screen.getByText(text);
      expect(tooltip).toBeInTheDocument();
      return tooltip;
    },
    { timeout: 1000 },
  );
}

describe("ButtonGroup", () => {
  it("ButtonGroupTestingWrapper -- buttons in test wrapper exist", () => {
    const { unmount } = render(ButtonGroupTestingWrapper, {
      values: [1, 2, 3],
      selected: [1, 2, 3],
      disabled: [1, 2, 3],
    });

    const buttons = [1, 2, 3];

    for (const i of buttons) {
      const button = screen.getByRole("button", { name: `button-${i}` });
      expect(button).toBeInstanceOf(HTMLButtonElement);
    }

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

    const buttons = [1, 2, 3];

    for (const i of buttons) {
      const button = screen.getByRole("button", { name: `button-${i}` });
      await fireEvent.click(button);
      expect(onClick).toBeCalledWith(expect.objectContaining({ detail: i }));
    }

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

    const buttons = [1, 2, 3];

    for (const i of buttons) {
      const button = screen.getByRole("button", { name: `button-${i}` });
      await fireEvent.click(button);
    }

    expect(onClick).toBeCalledTimes(0);
    unmount();
  });

  it("correct tooltips", async () => {
    const { unmount } = render(ButtonGroupTestingWrapper, {
      values: [1, 2, 3],
      selected: [1],
      disabled: [3],
    });

    const buttons = [
      [1, "selected tt"],
      [2, "unselected tt"],
      [3, "disabled tt"],
    ] as const;

    for (const [i, tt] of buttons) {
      const button = screen.getByRole("button", { name: `button-${i}` });
      if (!button?.parentElement) return;

      await hoverWithIntent(button.parentElement);
      const toolTip = await waitForTooltip(tt);
      expect(toolTip).toBeTruthy();
      await fireEvent.mouseLeave(button.parentElement);
      // Wait for tooltip to disappear
      await waitFor(() => {
        expect(screen.queryByText(tt)).not.toBeInTheDocument();
      });
    }

    unmount();
  });
});

describe("ButtonGroup - adding buttons", () => {
  let component: ButtonGroupTestingWrapper;
  let unmount: () => void;
  let onClick: (event: CustomEvent<unknown>) => void;

  beforeEach(() => {
    const { component: component_before, unmount: unmount_before } = render(
      ButtonGroupTestingWrapper,
      {
        values: [1, 2, 3],
        selected: [1, 2, 3],
        disabled: [],
      },
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
    component.$set({ values: [1, 2, 3, 4] });

    const button = await screen.findByRole("button", { name: `button-${4}` });

    // mock console.error to avoid irrelevant errors about
    // `scrollTo` not being implemented in jsdom
    const errorObject = console.error;
    console.error = vi.fn();

    if (!button?.parentElement) return;

    await hoverWithIntent(button.parentElement);
    let toolTip = await waitForTooltip("unselected tt");
    expect(toolTip).toBeTruthy();
    await fireEvent.mouseLeave(button.parentElement);
    await waitFor(() => {
      expect(screen.queryByText("unselected tt")).not.toBeInTheDocument();
    });

    component.$set({ values: [1, 2, 3, 4], selected: [4] });
    await hoverWithIntent(button.parentElement);
    toolTip = await waitForTooltip("selected tt");
    expect(toolTip).toBeTruthy();
    await fireEvent.mouseLeave(button.parentElement);
    await waitFor(() => {
      expect(screen.queryByText("selected tt")).not.toBeInTheDocument();
    });

    component.$set({ values: [1, 2, 3, 4], disabled: [4] });
    await hoverWithIntent(button.parentElement);
    toolTip = await waitForTooltip("disabled tt");
    expect(toolTip).toBeTruthy();
    await fireEvent.mouseLeave(button.parentElement);
    await waitFor(() => {
      expect(screen.queryByText("disabled tt")).not.toBeInTheDocument();
    });

    // unmock console.error
    console.error = errorObject;
  });
});

describe("ButtonGroup - removing buttons", () => {
  let component: ButtonGroupTestingWrapper;
  let unmount: () => void;
  let onClick: (event: CustomEvent<unknown>) => void;

  beforeEach(() => {
    const { component: component_before, unmount: unmount_before } = render(
      ButtonGroupTestingWrapper,
      {
        values: [1, 2, 3, 4, 5],
        selected: [1, 2],
        disabled: [4],
      },
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
    component.$set({ values: [1, 3, 4] });

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
    component.$set({ values: [1, 3, 4] });

    // mock console.error to avoid irrelevant errors about
    // `scrollTo` not being implemented in jsdom
    const errorObject = console.error;
    console.error = vi.fn();

    let button = await screen.findByRole("button", { name: `button-1` });

    if (!button?.parentElement) return;

    await hoverWithIntent(button.parentElement);
    let toolTip = await waitForTooltip("selected tt");
    expect(toolTip).toBeTruthy();
    await fireEvent.mouseLeave(button.parentElement);
    await waitFor(() => {
      expect(screen.queryByText("selected tt")).not.toBeInTheDocument();
    });

    button = await screen.findByRole("button", { name: `button-3` });
    if (!button?.parentElement) return;
    await hoverWithIntent(button.parentElement);
    toolTip = await waitForTooltip("unselected tt");
    expect(toolTip).toBeTruthy();
    await fireEvent.mouseLeave(button.parentElement);
    await waitFor(() => {
      expect(screen.queryByText("unselected tt")).not.toBeInTheDocument();
    });

    button = await screen.findByRole("button", { name: `button-4` });
    if (!button?.parentElement) return;
    await hoverWithIntent(button.parentElement);
    toolTip = await waitForTooltip("disabled tt");
    expect(toolTip).toBeTruthy();
    await fireEvent.mouseLeave(button.parentElement);
    await waitFor(() => {
      expect(screen.queryByText("disabled tt")).not.toBeInTheDocument();
    });

    // unmock console.error
    console.error = errorObject;
  });
});
