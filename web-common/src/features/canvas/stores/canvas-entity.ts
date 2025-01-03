import { writable, type Writable } from "svelte/store";
import { CanvasTimeControls } from "./canvas-time-control";

export class CanvasEntity {
  name: string;
  /**
   * Time controls for the canvas entity containing various
   * time related writables
   */
  timeControls: CanvasTimeControls;
  /**
   * Index of the component higlighted or selected in the canvas
   */
  selectedComponentIndex: Writable<number | null>;

  constructor(name: string) {
    this.name = name;
    this.timeControls = new CanvasTimeControls();
    this.selectedComponentIndex = writable(null);
  }

  setSelectedComponentIndex(index: number | null) {
    this.selectedComponentIndex.set(index);
  }
}
