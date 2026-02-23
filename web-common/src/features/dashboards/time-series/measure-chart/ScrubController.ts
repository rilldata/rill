import { writable, get, type Readable, type Writable } from "svelte/store";
import type { ScrubState } from "./types";
import type { ScaleLinear } from "d3-scale";

type ScrubMode = "none" | "create" | "resize-start" | "resize-end" | "move";

const EMPTY_SCRUB: ScrubState = {
  startIndex: null,
  endIndex: null,
  isScrubbing: false,
};

const EDGE_THRESHOLD_PX = 5;

type XScale = ScaleLinear<number, number>;

/**
 * Controller for chart scrub/brush selection interactions.
 * Designed as a singleton shared across all measure charts.
 *
 * All charts share the same x-scale domain (0 to dataLength-1),
 * so scrub state is shared and any chart can drive interactions.
 */
export class ScrubController {
  private _state: Writable<ScrubState>;
  private mode: ScrubMode = "none";
  private moveStartX: number | null = null;
  private moveStartIndices: { start: number; end: number } | null = null;
  private dataLength: number = 0;

  readonly state: Readable<ScrubState>;

  constructor(externalState?: Writable<ScrubState>) {
    this._state = externalState ?? writable<ScrubState>(EMPTY_SCRUB);
    this.state = this._state;
  }

  /** Update the data length (used for clamping indices). */
  setDataLength(length: number): void {
    this.dataLength = length;
  }

  /** Initialize controller state from external selection (e.g., from URL/props). */
  initFromExternal(startIndex: number, endIndex: number): void {
    this._state.set({
      startIndex,
      endIndex,
      isScrubbing: false,
    });
  }

  /** Clamp index to valid range [0, dataLength-1]. */
  private clamp(index: number): number {
    if (this.dataLength === 0) return 0;
    return Math.max(0, Math.min(this.dataLength - 1, Math.round(index)));
  }

  /** Determine scrub mode based on click position relative to existing selection. */
  private detectMode(screenX: number, xScale: XScale): ScrubMode {
    const state = get(this._state);

    if (state.startIndex === null || state.endIndex === null) {
      return "create";
    }

    const startX = xScale(state.startIndex);
    const endX = xScale(state.endIndex);

    if (Math.abs(screenX - startX) <= EDGE_THRESHOLD_PX) return "resize-start";
    if (Math.abs(screenX - endX) <= EDGE_THRESHOLD_PX) return "resize-end";

    const minX = Math.min(startX, endX);
    const maxX = Math.max(startX, endX);
    if (
      screenX > minX + EDGE_THRESHOLD_PX &&
      screenX < maxX - EDGE_THRESHOLD_PX
    ) {
      return "move";
    }

    return "create";
  }

  /** Get cursor style based on hover position relative to selection. */
  getCursorStyle(screenX: number | null, xScale: XScale): string {
    const state = get(this._state);

    if (state.isScrubbing) return "cursor-ew-resize";

    if (
      state.startIndex === null ||
      state.endIndex === null ||
      screenX === null
    ) {
      return "cursor-crosshair";
    }

    const startX = xScale(state.startIndex);
    const endX = xScale(state.endIndex);

    if (
      Math.abs(screenX - startX) <= EDGE_THRESHOLD_PX ||
      Math.abs(screenX - endX) <= EDGE_THRESHOLD_PX
    ) {
      return "cursor-ew-resize";
    }

    const minX = Math.min(startX, endX);
    const maxX = Math.max(startX, endX);
    if (
      screenX > minX + EDGE_THRESHOLD_PX &&
      screenX < maxX - EDGE_THRESHOLD_PX
    ) {
      return "cursor-grab";
    }

    return "cursor-crosshair";
  }

  /** Start a scrub interaction. */
  start(screenX: number, xScale: XScale): void {
    const state = get(this._state);
    const index = this.clamp(xScale.invert(screenX));

    this.mode = this.detectMode(screenX, xScale);

    if (
      this.mode === "move" &&
      state.startIndex !== null &&
      state.endIndex !== null
    ) {
      this.moveStartX = screenX;
      this.moveStartIndices = {
        start: state.startIndex,
        end: state.endIndex,
      };
      this._state.update((s) => ({ ...s, isScrubbing: true }));
    } else if (this.mode === "create") {
      this._state.set({
        startIndex: index,
        endIndex: index,
        isScrubbing: true,
      });
    } else {
      this._state.update((s) => ({ ...s, isScrubbing: true }));
    }
  }

  /** Update scrub position during drag. */
  update(screenX: number, xScale: XScale): void {
    const state = get(this._state);
    if (!state.isScrubbing) return;

    const index = this.clamp(xScale.invert(screenX));

    switch (this.mode) {
      case "create":
      case "resize-end":
        this._state.update((s) => ({ ...s, endIndex: index }));
        break;

      case "resize-start":
        this._state.update((s) => ({ ...s, startIndex: index }));
        break;

      case "move":
        if (this.moveStartX !== null && this.moveStartIndices) {
          const deltaX = screenX - this.moveStartX;
          const startPx = xScale(this.moveStartIndices.start);
          const endPx = xScale(this.moveStartIndices.end);
          this._state.update((s) => ({
            ...s,
            startIndex: this.clamp(xScale.invert(startPx + deltaX)),
            endIndex: this.clamp(xScale.invert(endPx + deltaX)),
          }));
        }
        break;
    }
  }

  /** End scrub interaction. Returns true if selection was kept, false if cleared. */
  end(): boolean {
    const state = get(this._state);

    if (!state.isScrubbing) {
      this.clearMoveState();
      return state.startIndex !== null;
    }

    // Clear if selection is too small (single point click)
    if (
      state.startIndex !== null &&
      state.endIndex !== null &&
      Math.abs(state.startIndex - state.endIndex) < 0.5
    ) {
      this.reset();
      return false;
    }

    this._state.update((s) => ({ ...s, isScrubbing: false }));
    this.clearMoveState();
    return true;
  }

  /** Reset scrub state completely. */
  reset(): void {
    this._state.set(EMPTY_SCRUB);
    this.clearMoveState();
  }

  private clearMoveState(): void {
    this.mode = "none";
    this.moveStartX = null;
    this.moveStartIndices = null;
  }
}
