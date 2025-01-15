import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { writable, type Writable } from "svelte/store";
import { CanvasFilters } from "./canvas-filters";
import { CanvasResolvedSpec } from "./canvas-spec";
import { CanvasTimeControls } from "./canvas-time-control";

export class CanvasEntity {
  name: string;

  /**
   * Time controls for the canvas entity containing various
   * time related writables
   */
  timeControls: CanvasTimeControls;

  /**
   * Dimension and measure filters for the canvas entity
   */
  filters: CanvasFilters;

  /**
   * Spec store containing selectors derived from ResolveCanvas query
   */
  spec: CanvasResolvedSpec;

  /**
   * Index of the component higlighted or selected in the canvas
   */
  selectedComponentIndex: Writable<number | null>;

  constructor(name: string, validSpecStore: CanvasSpecResponseStore) {
    this.name = name;
    this.selectedComponentIndex = writable(null);
    this.spec = new CanvasResolvedSpec(validSpecStore);
    this.filters = new CanvasFilters();
    this.timeControls = new CanvasTimeControls(validSpecStore);
  }

  setSelectedComponentIndex(index: number | null) {
    this.selectedComponentIndex.set(index);
  }
}
