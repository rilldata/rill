import type { BaseCanvasComponent } from "../components/BaseCanvasComponent";

export function hasComponentFilters(
  component: BaseCanvasComponent | null,
): boolean {
  if (!component) return false;
  return Object.keys(component.inputParams().filter).length > 0;
}
