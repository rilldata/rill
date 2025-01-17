import { useCanvas } from "@rilldata/web-common/features/canvas/selector";
import type { CanvasSpecResponseStore } from "@rilldata/web-common/features/canvas/types";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import { derived, writable, type Writable } from "svelte/store";
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
   * Spec store containing selectors derived from ResolveCanvas query
   */
  spec: CanvasResolvedSpec;

  /**
   * Index of the component higlighted or selected in the canvas
   */
  selectedComponentIndex: Writable<number | null>;

  constructor(name: string) {
    const validSpecStore: CanvasSpecResponseStore = derived(runtime, (r, set) =>
      useCanvas(r.instanceId, name, { queryClient }).subscribe(set),
    );

    this.name = name;

    this.selectedComponentIndex = writable(null);
    this.spec = new CanvasResolvedSpec(validSpecStore);
    this.timeControls = new CanvasTimeControls(validSpecStore);
  }

  setSelectedComponentIndex(index: number | null) {
    this.selectedComponentIndex.set(index);
  }
}
