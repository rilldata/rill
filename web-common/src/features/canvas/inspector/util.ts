import type { CanvasComponentObj } from "@rilldata/web-common/features/canvas/components/util";

export function hasComponentFilters(
  component: CanvasComponentObj | null,
): boolean {
  if (!component) return false;
  return Object.keys(component.inputParams().filter).length > 0;
}
