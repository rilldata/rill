import { CanvasFilters } from "./canvas-filters";
import type { CanvasResolvedSpec } from "./canvas-spec";

export class CanvasComponentState {
  filters: CanvasFilters;

  constructor(spec: CanvasResolvedSpec) {
    this.filters = new CanvasFilters(spec);
  }
}
