import type { ChartFieldInput } from "@rilldata/web-common/features/canvas/inspector/types";
import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";

export function hasComponentFilters(
  component: BaseCanvasComponent | null,
): boolean {
  if (!component) return false;
  return Object.keys(component.inputParams().filter).length > 0;
}

export function shouldShowPopover(
  chartFieldInput: ChartFieldInput | undefined,
): boolean {
  if (!chartFieldInput) return false;
  const popoverContributingProperties: Partial<keyof ChartFieldInput>[] = [
    "axisTitleSelector",
    "originSelector",
    "sortSelector",
    "limitSelector",
    "nullSelector",
    "labelAngleSelector",
    "axisRangeSelector",
    "defaultLegendOrientation",
    "totalSelector",
  ];

  const hasPopoverContent = popoverContributingProperties.some(
    (prop) => chartFieldInput[prop] !== undefined,
  );
  return hasPopoverContent;
}
