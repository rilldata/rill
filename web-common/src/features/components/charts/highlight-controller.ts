import type { View } from "svelte-vega";

/**
 * Programmatically sets the hover highlight on a Vega chart view.
 * This enables external components (e.g., a table) to highlight data points
 * in the chart by manipulating the Vega `hover_tuple` signal.
 *
 * @param view - The Vega View instance
 * @param time - The timestamp to highlight (null to clear)
 * @param dimensionValue - Optional dimension value for multi-series charts
 */
export function setExternalHover(
  view: View,
  time: Date | null | undefined,
  dimensionValue?: string | null,
): void {
  const epochTime = time ? time.getTime() : null;

  let fields: unknown[];
  try {
    fields = view.signal("hover_tuple_fields") || [];
  } catch {
    // Signal doesn't exist in this spec; nothing to highlight
    return;
  }

  // Only include dimension value when the hover selection actually has
  // a dimension field (fields.length > 1). Component charts use
  // encodings: ["x"] only, so fields has 1 entry; passing extra values
  // creates a mismatch that prevents Vega from matching data points.
  let values: unknown[] | null = null;
  if (epochTime !== null) {
    values =
      dimensionValue != null && fields.length > 1
        ? [epochTime, dimensionValue]
        : [epochTime];
  }

  const newValue = epochTime
    ? {
        unit: "",
        fields,
        values,
      }
    : null;

  // Check equality to avoid unnecessary re-renders
  let currentValues: unknown[];
  try {
    currentValues = (view.signal("hover_tuple") || { values: [] }).values;
  } catch {
    currentValues = [];
  }
  const newValues = values || [];

  if (isSignalEqual(currentValues, newValues)) {
    return;
  }

  view.signal("hover_tuple", newValue);
  void view.runAsync();
}

/**
 * Clears the hover highlight on a Vega chart view.
 */
export function clearExternalHover(view: View): void {
  view.signal("hover_tuple", null);
  void view.runAsync();
}

function isSignalEqual(
  currentValues: unknown[],
  newValues: unknown[],
): boolean {
  if (!Array.isArray(currentValues) || !Array.isArray(newValues)) {
    return false;
  }

  if (currentValues.length !== newValues.length) {
    return false;
  }

  for (let i = 0; i < currentValues.length; i++) {
    if (currentValues[i] !== newValues[i]) {
      return false;
    }
  }

  return true;
}
