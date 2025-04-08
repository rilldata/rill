import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { type Writable } from "svelte/store";
import { Filters } from "./filters";
import type { CanvasResolvedSpec } from "./spec";
import { TimeControls } from "./time-control";
import type { ComponentSpec } from "../components/types";

export class CanvasComponentState<T = ComponentSpec> {
  localFilters: Filters;
  localTimeControls: TimeControls;

  constructor(
    name: string,
    specStore: CanvasSpecResponseStore,
    spec: CanvasResolvedSpec,
    localSpec: Writable<T>,
  ) {
    this.localFilters = new Filters(spec);
    this.localTimeControls = new TimeControls(specStore, name);

    localSpec.subscribe((spec) => {
      if (spec?.["dimension_filters"]) {
        this.localFilters.setFiltersFromText(spec?.["dimension_filters"]);
      }

      if (spec?.["time_filters"]) {
        this.localTimeControls.setTimeFiltersFromText(spec?.["time_filters"]);
      }
    });
  }
}
