import { act, screen, waitFor } from "@testing-library/svelte";
import { expect } from "vitest";

/** The picker trigger, whose text content is the time range as the user sees it. */
export function getTimeRangePicker() {
  return screen.getByLabelText("Select time range");
}

/** Opens the time range dropdown. */
export async function openTimeRangePicker() {
  await act(() => getTimeRangePicker().click());
  await waitFor(() =>
    expect(screen.getByRole("menuitem", { name: /All time/ })).toBeVisible(),
  );
}

/**
 * Picks the range labelled `label` in the time range dropdown and waits for it to be applied.
 * The menu items carry their rilltime syntax alongside the label, so `label` is a pattern rather
 * than the accessible name in full.
 *
 * Applying a range needs a network call to resolve its interval, so the picker only shows the new
 * label once the range has reached the dashboard.
 */
export async function selectTimeRange(label: RegExp) {
  await openTimeRangePicker();
  await act(() => screen.getByRole("menuitem", { name: label }).click());
  await waitForTimeRangeLabel(label);
}

/** Waits for the picker to show `label`, which it does once the range is applied. */
export async function waitForTimeRangeLabel(label: RegExp) {
  await waitFor(() => expect(getTimeRangePicker()).toHaveTextContent(label));
}
