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

export function getTimeGrainPicker() {
  return screen.getByLabelText("Select aggregation grain");
}

function getAdaptiveTimeGrainOption() {
  return screen.getByRole("menuitemcheckbox", { name: /Adaptive/ });
}

export async function openTimeGrainPicker() {
  await act(() => getTimeGrainPicker().click());
  await waitFor(() => expect(getAdaptiveTimeGrainOption()).toBeVisible());
}

/** Closes the time grain dropdown if it is still open after a selection. */
async function closeTimeGrainPicker() {
  if (!screen.queryByRole("menuitemcheckbox", { name: /Adaptive/ })) return;
  await act(() => getTimeGrainPicker().click());
  await waitFor(() =>
    expect(
      screen.queryByRole("menuitemcheckbox", { name: /Adaptive/ }),
    ).toBeNull(),
  );
}

/**
 * The grains listed in the time grain dropdown, as the user reads them.
 * The adaptive option is left out since it is listed regardless of the time range.
 */
export async function getTimeGrainOptions() {
  await openTimeGrainPicker();
  const options = screen
    .getAllByRole("menuitemcheckbox")
    .map((option) => option.textContent?.trim() ?? "")
    .filter((label) => !label.startsWith("Adaptive"));
  await closeTimeGrainPicker();
  return options;
}

/** Picks a fixed grain, e.g. `hour`, and waits for the trigger to show it. */
export async function selectTimeGrain(grainLabel: string) {
  await openTimeGrainPicker();
  await act(() =>
    screen.getByRole("menuitemcheckbox", { name: grainLabel }).click(),
  );
  await closeTimeGrainPicker();
  await waitForTimeGrainLabel(grainLabel, false);
}

/** Picks the adaptive option, which lets the grain follow the time range. */
export async function selectAdaptiveTimeGrain() {
  await openTimeGrainPicker();
  await act(() => getAdaptiveTimeGrainOption().click());
  await closeTimeGrainPicker();
  await waitFor(() => expect(getTimeGrainPicker()).toHaveTextContent(/auto/));
}

/**
 * Waits for the trigger to show `grainLabel`,
 * followed by the `auto` hint when the grain is adaptive.
 */
export async function waitForTimeGrainLabel(
  grainLabel: string,
  adaptive: boolean,
) {
  await waitFor(() =>
    expect(getTimeGrainPicker().textContent?.replace(/\s+/g, " ").trim()).toBe(
      `by ${grainLabel}${adaptive ? " auto" : ""}`,
    ),
  );
}
