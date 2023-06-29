// import Litepicker from "../date-picker/Litepicker.svelte";
// import { ButtonGroupTestingWrapper } from "./ButtonGroupTestingWrapper.svelte";
import ButtonGroupTestingWrapper from "./ButtonGroupTestingWrapper.svelte";
import { describe, it, expect, beforeEach, afterEach, vi } from "vitest";
import {
  render,
  waitFor,
  fireEvent,
  screen,
  getByText,
  getByRole,
} from "@testing-library/svelte";
// import {  } from "@testin";

describe("ButtonGroupTestingWrapper", () => {
  beforeEach(() => {});

  afterEach(() => {});

  it("buttons in test wrapper exist", async () => {
    const { component, unmount } = render(ButtonGroupTestingWrapper, {
      values: [1, 2, 3],
      selected: [1, 2, 3],
      disabled: [1, 2, 3],
    });

    await waitFor(() => {
      expect(screen.getByRole("button", { name: "button-1" })).toBeInstanceOf(
        HTMLButtonElement
      );
      expect(screen.getByRole("button", { name: "button-2" })).toBeInstanceOf(
        HTMLButtonElement
      );
      expect(screen.getByRole("button", { name: "button-3" })).toBeInstanceOf(
        HTMLButtonElement
      );
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
      let button = screen.getByRole("button", { name: "button-1" });
      expect(button).toBeInstanceOf(HTMLButtonElement);
      await fireEvent.click(button);
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
      let button = screen.getByRole("button", { name: "button-1" });
      expect(button).toBeInstanceOf(HTMLButtonElement);
      await fireEvent.click(button);
    });

    expect(onClick).toBeCalledTimes(0);
    unmount();
  });

  // it("correct tooltips", async () => {
  //   // const onClick = vi.fn();

  //   const { component, unmount } = render(ButtonGroupTestingWrapper, {
  //     values: [1, 2, 3],
  //     selected: [1],
  //     disabled: [3],
  //   });
  //   // component.$on("subbutton-click", onClick);

  //   // [1, 2, 3].forEach(async (i) => {
  //     let button = screen.getByRole("button", { name: "button-1" });
  //     expect(button).toBeInstanceOf(HTMLButtonElement);
  //     await userEven
  //   // });

  //   expect(onClick).toBeCalledTimes(0);
  //   unmount();
  // });
});
