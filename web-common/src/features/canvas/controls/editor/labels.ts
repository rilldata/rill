import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import type { ControlLevelRef } from "../control-edit";
import { ControlKind } from "../controls";

/** What a control of a kind lets viewers choose, e.g. "Measure". */
export function controlKindLabel(kind: string): string {
  switch (kind) {
    case ControlKind.Measure:
      return m.canvas_control_kind_measure();
    case ControlKind.Dimension:
      return m.canvas_control_kind_dimension();
    case ControlKind.Filter:
      return m.canvas_control_kind_filter();
    case ControlKind.TimeRange:
      return m.canvas_control_kind_time_range();
    case ControlKind.TimeGrain:
      return m.canvas_control_kind_time_grain();
    default:
      return m.canvas_control_kind_slot();
  }
}

/** The name of a level in the "Applies to" choice, e.g. "Container deep_dive". */
export function controlLevelLabel(level: ControlLevelRef): string {
  switch (level.kind) {
    case "item":
      return m.canvas_control_level_item();
    case "tab":
      return m.canvas_control_level_tab({ name: level.name });
    case "group":
      return m.canvas_control_level_group({ name: level.name });
    case "container":
      return m.canvas_control_level_container({ name: level.name });
    default:
      return m.canvas_control_level_canvas();
  }
}
