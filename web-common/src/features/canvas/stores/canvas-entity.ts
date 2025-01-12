import { writable, type Writable } from "svelte/store";
import { CanvasFilters } from "./canvas-filters";
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
   * Index of the component higlighted or selected in the canvas
   */
  selectedComponentIndex: Writable<number | null>;

  constructor(name: string) {
    this.name = name;
    this.timeControls = new CanvasTimeControls();
    this.filters = new CanvasFilters();
    this.selectedComponentIndex = writable(null);
  }

  setSelectedComponentIndex(index: number | null) {
    this.selectedComponentIndex.set(index);
  }
}
