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
 * The menu items carry their rilltime syntax alongside the label,
 * so `label` is a pattern rather than the accessible name in full.
 */
export async function selectTimeRange(label: RegExp) {
  await openTimeRangePicker();
  await act(() => screen.getByRole("menuitem", { name: label }).click());
  // Applying a range needs a network call to resolve its interval,
  // so the picker only shows the new label once the range has reached the dashboard.
  // TODO: fix this and make sure the dashboard has the new name immediately and resolved async
  await waitForTimeRangeLabel(label);
}

/** Waits for the picker to show `label`, which it does once the range is applied. */
export async function waitForTimeRangeLabel(label: RegExp) {
  await waitFor(() => expect(getTimeRangePicker()).toHaveTextContent(label));
}

export function getSnapPicker() {
  return screen.getByLabelText("Select reference time and grain");
}

export async function openSnapPicker() {
  await act(() => getSnapPicker().click());
  await waitFor(() =>
    expect(
      screen.getByRole("menuitemcheckbox", { name: /latest/ }),
    ).toBeVisible(),
  );
}

export async function selectSnapRefOrGrain(
  label: string,
  selectionLabel: string,
) {
  await openSnapPicker();
  await act(() =>
    screen.getByRole("menuitemcheckbox", { name: label }).click(),
  );
  // Applying a range needs a network call to resolve its interval,
  // so the picker only shows the new label once the range has reached the dashboard.
  // TODO: fix this and make sure the dashboard has the new name immediately and resolved async
  await waitFor(() =>
    expect(getSnapPicker()).toHaveTextContent(selectionLabel),
  );
}

function getSnapOffsetToggle() {
  return screen.getByLabelText("Anchor to period end");
}

export async function snapOffsetToggleIsDisabled() {
  await openSnapPicker();
  await waitFor(() => expect(getSnapOffsetToggle()).toBeVisible());
  await act(() => getSnapPicker().click());
}

export async function toggleSnapOffset(label: string) {
  await openSnapPicker();
  await act(() => getSnapOffsetToggle().click());
  await waitFor(() => expect(getSnapPicker()).toHaveTextContent(label));
  // Snap toggle doesnt auto toggle
  await act(() => getSnapPicker().click());
}

export function getComparisonTimeRangePicker() {
  return screen.getByLabelText("Select time comparison option");
}

export async function openComparisonTimeRangePicker() {
  await act(() => getComparisonTimeRangePicker().click());
  await waitFor(() =>
    expect(screen.getByRole("menuitem", { name: "Custom" })).toBeVisible(),
  );
}

export async function selectComparisonTimeRange(label: string) {
  await openComparisonTimeRangePicker();
  await act(() => screen.getByRole("menuitem", { name: label }).click());
  await waitFor(() =>
    expect(getComparisonTimeRangePicker()).toHaveTextContent(label),
  );
}

export function getToggleComparisonPicker() {
  return screen.getByLabelText("Toggle time comparison");
}

export async function toggleComparison() {
  await act(() => getToggleComparisonPicker().click());
}
