import { getRangeLabel } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { V1CanvasControl } from "@rilldata/web-common/runtime-client";
import { ControlKind } from "./controls";

/** Looks up the display name of a measure or dimension. */
export type FieldLabel = (name: string) => string;

/** What viewers read for one value of a control: a field's display name, "Last 7 days", "Week". */
export function controlOptionLabel(
  control: V1CanvasControl,
  value: string,
  fieldLabel: FieldLabel,
): string {
  switch (control.kind) {
    case ControlKind.Measure:
    case ControlKind.Dimension:
    case ControlKind.Slot:
      return fieldLabel(value);
    case ControlKind.TimeRange:
      return timeRangeLabel(value);
    case ControlKind.TimeGrain:
      return value.charAt(0).toUpperCase() + value.slice(1);
    default:
      return value;
  }
}

/** "Last 7 days" for P7D, "Month to date" for rill-MTD; the raw value if it has no label. */
export function timeRangeLabel(range: string): string {
  const label = getRangeLabel(range);
  return label === m.time_custom() ? range : label;
}

/** The label shown next to a control. Empty for a swap without a label, whose options speak for themselves. */
export function controlLabel(
  control: V1CanvasControl,
  fieldLabel: FieldLabel,
): string {
  if (control.label) return control.label;
  return control.kind === ControlKind.Filter ||
    control.kind === ControlKind.Slot
    ? fieldLabel(control.target ?? "")
    : "";
}

/** A control's current values in a few words, e.g. "Acme, Globex" or "3 selected". */
export function controlSummary(
  control: V1CanvasControl,
  values: string[],
  fieldLabel: FieldLabel,
): string {
  if (values.length === 0) {
    return control.kind === ControlKind.Filter
      ? m.canvas_control_all()
      : m.canvas_control_same_as_canvas();
  }
  if (values.length <= 2) {
    return values
      .map((v) => controlOptionLabel(control, v, fieldLabel))
      .join(", ");
  }
  return m.canvas_control_selected_count({ count: values.length });
}
