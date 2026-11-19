import { getRangeLabel } from "@rilldata/web-common/features/dashboards/time-controls/new-time-controls";
import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type {
  V1CanvasControl,
  V1MetricsView,
} from "@rilldata/web-common/runtime-client";
import { ControlKind, type ControlValue } from "./controls";

/** Looks up the display name of a measure or dimension. */
export type FieldLabel = (name: string) => string;

/** What viewers read for one value of a control: a field's display name, "Last 7 days", "Week". */
export function controlOptionLabel(
  control: V1CanvasControl,
  value: ControlValue,
  fieldLabel: FieldLabel,
): string {
  if (value === null) return m.canvas_control_null();
  const text = String(value);
  switch (control.kind) {
    case ControlKind.Measure:
    case ControlKind.Dimension:
    case ControlKind.Slot:
      return fieldLabel(text);
    case ControlKind.TimeRange:
      return timeRangeLabel(text);
    case ControlKind.TimeGrain:
      return text.charAt(0).toUpperCase() + text.slice(1);
    default:
      return text;
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
  values: ControlValue[],
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

/** First display name for each field in the canvas's metrics views. */
export function controlFieldLabels(
  metricsViews: Record<string, V1MetricsView | undefined>,
): Map<string, string> {
  const labels = new Map<string, string>();
  for (const view of Object.values(metricsViews)) {
    const spec = view?.state?.validSpec;
    for (const field of [
      ...(spec?.measures ?? []),
      ...(spec?.dimensions ?? []),
    ]) {
      if (field.name && !labels.has(field.name))
        labels.set(field.name, field.displayName || field.name);
    }
  }
  return labels;
}
