import { DashboardFetchMocks } from "@rilldata/web-common/features/dashboards/dashboard-fetch-mocks";
import {
  act,
  fireEvent,
  screen,
  waitFor,
  within,
} from "@testing-library/svelte";
import { beforeAll, expect } from "vitest";

export function useDashboardFetchMocksForComponentTests() {
  const mocks = DashboardFetchMocks.useDashboardFetchMocks();
  mocks.mockMetricsViewAggregation(
    /BUILTIN_MEASURE_COUNT_DISTINCT.*"val":"%oo%"/,
    {
      data: [{ publisher__distinct_count: 3 }],
    },
  );
  mocks.mockMetricsViewAggregation(/"val":"%oo%"/, {
    data: [
      { publisher: "Facebook" },
      { publisher: "Google" },
      { publisher: "Yahoo" },
    ],
  });
  mocks.mockMetricsViewAggregation(
    /BUILTIN_MEASURE_COUNT_DISTINCT.*{"val":"Facebook"},{"val":"Google"}/,
    {
      data: [{ publisher__distinct_count: 2 }],
    },
  );
  mocks.mockMetricsViewAggregation(/{"val":"Facebook"},{"val":"Google"}/, {
    data: [{ publisher: "Facebook" }, { publisher: "Google" }],
  });
  mocks.mockMetricsViewAggregation(/publisher/, {
    data: [
      { publisher: null },
      { publisher: "Facebook" },
      { publisher: "Google" },
      { publisher: "Yahoo" },
      { publisher: "Microsoft" },
    ],
  });
  return mocks;
}

/**
 * bits-ui 2.x uses PointerEvent APIs that jsdom doesn't support.
 * Polyfill the missing types and methods so pointer-based interactions work in tests.
 */
export function mockPointerEventsForComponentTesting() {
  beforeAll(() => {
    if (typeof globalThis.PointerEvent === "undefined") {
      (globalThis as Record<string, unknown>).PointerEvent =
        class PointerEvent extends MouseEvent {
          readonly pointerId: number;
          readonly pointerType: string;
          constructor(
            type: string,
            init?: PointerEventInit & Record<string, unknown>,
          ) {
            super(type, init);
            this.pointerId = (init?.pointerId as number) ?? 0;
            this.pointerType = (init?.pointerType as string) ?? "mouse";
          }
        };
    }
    if (!HTMLElement.prototype.hasPointerCapture) {
      HTMLElement.prototype.hasPointerCapture = () => false;
    }
    if (!HTMLElement.prototype.releasePointerCapture) {
      HTMLElement.prototype.releasePointerCapture = () => {};
    }
    if (!HTMLElement.prototype.scrollIntoView) {
      HTMLElement.prototype.scrollIntoView = () => {};
    }
  });
}

/**
 * bits-ui schedules a body scroll lock cleanup that runs after the test ends, which throws
 * "document is not defined" once jsdom is torn down. Await this in an `afterAll` to let it run.
 */
export async function waitForBodyScrollCleanup() {
  await new Promise((resolve) => window.setTimeout(resolve, 30));
}

/** Adds a filter chip for `name` through the "add filter" menu. */
export async function addFilter(name: string) {
  await act(() => {
    screen.getByLabelText("Add filter button").click();
  });
  await waitFor(() =>
    expect(screen.getByRole("menuitem", { name })).toBeVisible(),
  );
  await act(() => {
    screen.getByRole("menuitem", { name }).click();
  });
  await waitFor(() =>
    expect(screen.queryByRole("menuitem", { name })).toBeNull(),
  );
}

/** The chip for `name`, whose text content is the filter as the user sees it. */
export function getFilterChip(name: string) {
  return screen.getByLabelText(`Open ${name} filter`);
}

/** Opens or closes the dropdown of the chip for `name`. */
export async function toggleFilter(name: string) {
  await act(() => getFilterChip(name).click());
}

export async function closeFilter(name: string) {
  await toggleFilter(name);
  await waitFor(() =>
    expect(screen.queryByRole("menu")).not.toBeInTheDocument(),
  );
}

/** The text of the mode selector trigger, which is the mode the open dropdown is in. */
export function getDimensionFilterModeText() {
  return document.getElementById("dimension-filter-mode-selector")!.textContent;
}

/** Picks a mode in the open dimension filter dropdown. */
export async function selectDimensionFilterMode(name: RegExp) {
  await selectFromDropdown("dimension-filter-mode-selector", name);
}

/**
 * Picks the option matching `name` in the `Select` with trigger id `triggerId`.
 *
 * bits-ui 2.x Select: open via keyboard (Space) on the trigger, then navigate with Home and
 * ArrowDown and select with Space. Pointer events don't work reliably in jsdom because bits-ui's
 * item ref tracking requires real layout.
 */
export async function selectFromDropdown(triggerId: string, name: RegExp) {
  const trigger = document.getElementById(triggerId)!;
  trigger.focus();
  await act(async () => {
    await fireEvent.keyDown(trigger, { key: " " });
  });
  await waitFor(() =>
    expect(screen.getByRole("option", { name })).toBeVisible(),
  );

  // Highlight the first option, then walk down to the target. bits-ui highlights the current
  // value on open, so navigating from a known position keeps this independent of the option
  // the select was on.
  await act(async () => {
    await fireEvent.keyDown(trigger, { key: "Home" });
  });
  const targetIndex = screen
    .getAllByRole("option")
    .findIndex((option) => name.test(option.textContent.trim()));
  for (let i = 0; i < targetIndex; i++) {
    await act(async () => {
      await fireEvent.keyDown(trigger, { key: "ArrowDown" });
    });
  }

  // Space commits the highlighted option. Enter would too, but it also reaches the window level
  // handler the dimension and measure filters have, which applies and closes the whole filter.
  await act(async () => {
    await fireEvent.keyDown(trigger, { key: " " });
  });
}

/** Types `text` into the search input of the open dimension filter dropdown. */
export async function typeInDimensionFilterSearch(name: string, text: string) {
  await act(() =>
    fireEvent.input(screen.getByLabelText(`${name} search list`), {
      target: { value: text },
    }),
  );
}

/**
 * The text of each item in the results of the open dimension filter dropdown.
 * bits-ui 2.x renders items as adjacent elements without whitespace, so checking individual
 * items is more reliable than toHaveTextContent.
 */
export function getDimensionFilterResults(name: string) {
  const group = screen.getByLabelText(`${name} results`);
  const items = group.querySelectorAll(
    "[data-dropdown-menu-item], [data-dropdown-menu-checkbox-item]",
  );
  return Array.from(items).map((el) => el.textContent?.trim() ?? "");
}

/** Waits for the result count of the open dimension filter dropdown to be `count`. */
export async function waitForDimensionFilterResultCount(
  name: string,
  count: string,
) {
  await waitFor(() =>
    expect(screen.getByLabelText(`${name} result count`)).toHaveTextContent(
      count,
    ),
  );
}

/**
 * The chip for the measure filter with `label`, whose text content is the filter as the user sees
 * it. Measure chips are labelled by the measure's display name, without the `filter` suffix the
 * dimension chips have.
 */
export function getMeasureFilterChip(label: string) {
  return screen.getByLabelText(`Open ${label}`);
}

/** Removes the measure chip for `label`. */
export async function removeMeasureFilter(label: string) {
  await act(() =>
    within(screen.getByLabelText(label)).getByLabelText("Remove").click(),
  );
}

/** Opens or closes the popover of the measure chip for `label`. */
export async function toggleMeasureFilter(label: string) {
  await act(() => getMeasureFilterChip(label).click());
}

/**
 * Closes the popover of the measure chip for `label`, discarding whatever the form holds.
 * The popover has no `role="menu"`, so `closeFilter` does not work for measure chips.
 */
export async function closeMeasureFilter(label: string) {
  await toggleMeasureFilter(label);
  await waitFor(() => expect(isMeasureFilterFormOpen()).toBe(false));
}

/**
 * Whether the measure filter form is open.
 *
 * bits-ui owns the id of the popover content, and jsdom gives the popover no layout to check
 * visibility against, so the form controls are what is left to key off.
 */
export function isMeasureFilterFormOpen() {
  return !!document.getElementById("value1");
}

/**
 * Fills the fields that are passed in on the measure filter form of the open popover.
 * Does not apply the filter; the form only reaches the dashboard through Apply.
 */
export async function fillMeasureFilterForm({
  dimension,
  operation,
  value1,
  value2,
}: {
  dimension?: RegExp;
  operation?: RegExp;
  value1?: string;
  value2?: string;
}) {
  if (dimension) await selectFromDropdown("dimension", dimension);
  if (operation) await selectFromDropdown("operation", operation);
  if (value1 !== undefined) await typeInMeasureFilterValue("value1", value1);
  if (value2 !== undefined) await typeInMeasureFilterValue("value2", value2);
}

/**
 * Submits the measure filter form of the open popover, which is the only way a measure filter
 * reaches the dashboard.
 *
 * superforms validates the form in a macrotask, so the click alone leaves nothing to assert on.
 * A valid form closes the popover; an invalid one keeps it open and renders its errors.
 */
export async function applyMeasureFilter() {
  await act(async () => {
    screen.getByRole("button", { name: "Apply" }).click();
    await new Promise((resolve) => window.setTimeout(resolve));
    await new Promise((resolve) => window.setTimeout(resolve));
  });
}

async function typeInMeasureFilterValue(id: string, value: string) {
  await act(() =>
    fireEvent.input(document.getElementById(id)!, { target: { value } }),
  );
}
