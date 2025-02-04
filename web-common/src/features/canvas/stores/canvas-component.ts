import { get } from "svelte/store";
import { CanvasFilters } from "./canvas-filters";
import type { CanvasResolvedSpec } from "./canvas-spec";

export class CanvasComponentState {
  name: string;
  filters: CanvasFilters;
  private spec: CanvasResolvedSpec;

  constructor(name: string, spec: CanvasResolvedSpec) {
    this.name = name;
    this.filters = new CanvasFilters(spec);
    this.spec = spec;

    const componentResourceStore = spec.getComponentResourceFromName(name);
    const componentResource = get(componentResourceStore);

    if (
      componentResource &&
      componentResource.rendererProperties?.dimension_filters
    ) {
      this.filters.setFiltersFromText(
        componentResource.rendererProperties?.dimension_filters as string,
      );
    }
  }
}
