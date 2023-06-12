import Litepicker from "./Litepicker.svelte";
import { describe, it, expect, beforeEach, afterEach } from "vitest";
import { render, fireEvent, waitFor } from "@testing-library/svelte";

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
    render(Litepicker, {
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
  });

  it.skip("calls change in reaction to day click events", () => {});

  it.skip("calls change in reaction to input change events", () => {});

  it.skip("calls editing event when the date being edited is changed", () => {});

  it("renders by default the starting date", () => {
    const { container } = render(Litepicker, {
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
  });
});

// <Litepicker
//       {startEl}
//       {endEl}
//       defaultStart={start}
//       defaultEnd={end}
//       openOnMount
//       on:change={handleDatePickerChange}
//       on:editing={handleEditingChange}
//       on:toggle={handleToggle}
//     />
