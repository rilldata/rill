import type { GridStack } from "gridstack";
import type { DashboardTimeControls } from "@rilldata/web-common/lib/time/types";

export interface CanvasEntity {
  name: string;

  /**
   * Index of the component higlighted or selected in the canvas
   */
  selectedComponentIndex: number | null;

  /**
   * user selected time range
   */
  selectedTimeRange?: DashboardTimeControls;

  /**
   * user selected timezone, should default to "UTC" if no other value is set
   */
  selectedTimezone: string;

  proto?: string;

  gridstack?: GridStack | null;
}
