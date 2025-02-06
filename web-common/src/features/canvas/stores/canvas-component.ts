import { get } from "svelte/store";
import { Filters } from "./filters";
import type { CanvasResolvedSpec } from "./spec";

export class CanvasComponentState {
  name: string;
  filters: Filters;

  constructor(name: string, spec: CanvasResolvedSpec) {
    this.name = name;
    this.filters = new Filters(spec);

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
