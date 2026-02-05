import type {
  ComponentCommonProperties,
  ComponentComparisonOptions,
  ComponentFilterProperties,
} from "../types";

export { default as KPI } from "./KPI.svelte";

export const SPARKLINE_MIN_WIDTH = 128;
export const BIG_NUMBER_MIN_WIDTH = 160;
export const padding = 32;
export const SPARK_RIGHT_MIN =
  SPARKLINE_MIN_WIDTH + 8 + BIG_NUMBER_MIN_WIDTH + padding;

export function getMinWidth(
  sparkline: "none" | "bottom" | "right" | undefined,
): number {
  switch (sparkline) {
    case "right":
      return SPARK_RIGHT_MIN;
    case "none":
    case "bottom":
    default:
      return BIG_NUMBER_MIN_WIDTH + padding;
  }
}

export interface KPISpec
  extends ComponentCommonProperties,
    ComponentFilterProperties {
  metrics_view: string;
  measure: string;
  // Defaults to "bottom"
  sparkline?: "none" | "bottom" | "right";
  // Defaults to "delta" and "percent_change"
  comparison?: ComponentComparisonOptions[];
  hide_time_range?: boolean;
}
