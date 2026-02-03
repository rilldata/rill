import { derived, writable, get, type Readable, type Writable } from "svelte/store";
import type {
  ChartScales,
  PlotBounds,
  HoverState,
  ScrubState,
  BisectedPoint,
  InteractionState,
  InteractionHandlers,
} from "./types";

/**
 * Create an IntersectionObserver-based visibility store.
 */
export function createVisibilityObserver(
  rootMargin = "120px",
): {
  visible: Writable<boolean>;
  observe: (element: HTMLElement, root?: HTMLElement | null) => () => void;
} {
  const visible = writable(false);

  function observe(
    element: HTMLElement,
    root: HTMLElement | null = null,
  ): () => void {
    const observer = new IntersectionObserver(
      ([entry]) => {
        if (entry.isIntersecting) {
          visible.set(true);
          observer.unobserve(element);
        }
      },
      { root, rootMargin, threshold: 0 },
    );
    observer.observe(element);
    return () => observer.disconnect();
  }

  return { visible, observe };
}

const EMPTY_HOVER_STATE: HoverState = {
  index: null,
  screenX: null,
  isHovered: false,
};

const EMPTY_SCRUB_STATE: ScrubState = {
  startIndex: null,
  endIndex: null,
  isScrubbing: false,
};

const EMPTY_BISECTED: BisectedPoint = { index: -1 };

type ScrubMode = "none" | "create" | "resize-start" | "resize-end" | "move";

/**
 * Index-based chart interactions.
 * x scale is linear (domain = [0, N-1]), so bisection is just Math.round(xScale.invert(px)).
 */
export function createChartInteractions(
  scalesStore: Readable<ChartScales>,
  visibleRangeStore: Readable<[number, number]>,
  plotBoundsStore: Readable<PlotBounds>,
  externalScrubState?: Writable<ScrubState>,
): {
  state: Readable<InteractionState>;
  handlers: InteractionHandlers;
  resetScrub: () => void;
} {
  const hoverState = writable<HoverState>(EMPTY_HOVER_STATE);
  const internalScrubState = writable<ScrubState>(EMPTY_SCRUB_STATE);
  const scrubState = externalScrubState ?? internalScrubState;

  let scrubMode: ScrubMode = "none";
  let scrubMoveStartX: number | null = null;
  let scrubMoveStartIndices: { start: number | null; end: number | null } | null = null;

  const EDGE_THRESHOLD = 5;

  /** Snap fractional index to nearest valid visible index. */
  function snap(fractionalIndex: number, range: [number, number]): number {
    return Math.max(range[0], Math.min(range[1], Math.round(fractionalIndex)));
  }

  const bisectedPoint = derived(
    [hoverState, scalesStore, visibleRangeStore],
    ([$hover, _$scales, $range]) => {
      if ($hover.index === null || $range[0] === $range[1] && $range[0] === 0) return EMPTY_BISECTED;
      return { index: snap($hover.index, $range) };
    },
  );

  const cursorStyle = derived(
    [scrubState, hoverState, scalesStore],
    ([$scrub, $hover, $scales]) => {
      if ($scrub.isScrubbing) return "cursor-ew-resize";

      if ($scrub.startIndex !== null && $scrub.endIndex !== null && $hover.screenX !== null) {
        const startX = $scales.x($scrub.startIndex);
        const endX = $scales.x($scrub.endIndex);
        const hx = $hover.screenX;

        if (Math.abs(hx - startX) <= EDGE_THRESHOLD || Math.abs(hx - endX) <= EDGE_THRESHOLD) {
          return "cursor-ew-resize";
        }
        const minX = Math.min(startX, endX);
        const maxX = Math.max(startX, endX);
        if (hx > minX + EDGE_THRESHOLD && hx < maxX - EDGE_THRESHOLD) {
          return "cursor-grab";
        }
      }
      return "cursor-crosshair";
    },
  );

  const state = derived(
    [hoverState, scrubState, bisectedPoint, cursorStyle],
    ([$hover, $scrub, $bisected, $cursor]) => ({
      hover: $hover,
      scrub: $scrub,
      bisectedPoint: $bisected,
      cursorStyle: $cursor,
    }),
  );

  function getScrubMode(hoverX: number): ScrubMode {
    const $scrub = get(scrubState);
    const $scales = get(scalesStore);

    if ($scrub.startIndex === null || $scrub.endIndex === null) return "create";

    const startX = $scales.x($scrub.startIndex);
    const endX = $scales.x($scrub.endIndex);

    if (Math.abs(hoverX - startX) <= EDGE_THRESHOLD) return "resize-start";
    if (Math.abs(hoverX - endX) <= EDGE_THRESHOLD) return "resize-end";

    const minX = Math.min(startX, endX);
    const maxX = Math.max(startX, endX);
    if (hoverX > minX + EDGE_THRESHOLD && hoverX < maxX - EDGE_THRESHOLD) return "move";

    return "create";
  }

  function resetScrub(): void {
    scrubState.set(EMPTY_SCRUB_STATE);
    scrubMode = "none";
    scrubMoveStartX = null;
    scrubMoveStartIndices = null;
  }

  const handlers: InteractionHandlers = {
    onMouseMove(event: MouseEvent) {
      const $scales = get(scalesStore);
      const $bounds = get(plotBoundsStore);

      const x = Math.max($bounds.left, Math.min($bounds.left + $bounds.width, event.offsetX));
      const fractionalIndex = $scales.x.invert(x);

      hoverState.set({
        index: fractionalIndex,
        screenX: x,
        isHovered: true,
      });

      const $scrub = get(scrubState);
      const $range = get(visibleRangeStore);
      if ($scrub.isScrubbing) {
        const snappedIndex = snap(fractionalIndex, $range);
        switch (scrubMode) {
          case "create":
          case "resize-end":
            scrubState.update((s) => ({ ...s, endIndex: snappedIndex }));
            break;
          case "resize-start":
            scrubState.update((s) => ({ ...s, startIndex: snappedIndex }));
            break;
          case "move":
            if (scrubMoveStartX !== null && scrubMoveStartIndices) {
              const deltaX = x - scrubMoveStartX;
              const startPx = $scales.x(scrubMoveStartIndices.start!);
              const endPx = $scales.x(scrubMoveStartIndices.end!);
              scrubState.update((s) => ({
                ...s,
                startIndex: snap($scales.x.invert(startPx + deltaX), $range),
                endIndex: snap($scales.x.invert(endPx + deltaX), $range),
              }));
            }
            break;
        }
      }
    },

    onMouseLeave() {
      hoverState.set(EMPTY_HOVER_STATE);
    },

    onMouseDown(event: MouseEvent) {
      if (event.button !== 0) return;
      const $scales = get(scalesStore);
      const $scrub = get(scrubState);
      const $range = get(visibleRangeStore);
      const x = event.offsetX;
      const idx = snap($scales.x.invert(x), $range);

      scrubMode = getScrubMode(x);

      if (scrubMode === "move") {
        scrubMoveStartX = x;
        scrubMoveStartIndices = { start: $scrub.startIndex, end: $scrub.endIndex };
        scrubState.update((s) => ({ ...s, isScrubbing: true }));
      } else if (scrubMode === "create") {
        scrubState.set({ startIndex: idx, endIndex: idx, isScrubbing: true });
      } else {
        scrubState.update((s) => ({ ...s, isScrubbing: true }));
      }
    },

    onMouseUp() {
      const $scrub = get(scrubState);
      if ($scrub.isScrubbing) {
        if (
          $scrub.startIndex !== null &&
          $scrub.endIndex !== null &&
          Math.abs($scrub.startIndex - $scrub.endIndex) < 0.5
        ) {
          resetScrub();
        } else {
          scrubState.update((s) => ({ ...s, isScrubbing: false }));
        }
      }
      scrubMode = "none";
      scrubMoveStartX = null;
      scrubMoveStartIndices = null;
    },

    onClick(event: MouseEvent) {
      const $scrub = get(scrubState);
      const $scales = get(scalesStore);

      if ($scrub.startIndex !== null && $scrub.endIndex !== null && !$scrub.isScrubbing) {
        const clickIdx = $scales.x.invert(event.offsetX);
        const [min, max] =
          $scrub.startIndex < $scrub.endIndex
            ? [$scrub.startIndex, $scrub.endIndex]
            : [$scrub.endIndex, $scrub.startIndex];
        if (clickIdx < min || clickIdx > max) {
          resetScrub();
        }
      }
    },
  };

  return { state, handlers, resetScrub };
}

/**
 * Helper to get ordered start/end dates.
 */
export function getOrderedDates(
  start: Date | null,
  end: Date | null,
): { start: Date | null; end: Date | null } {
  if (!start || !end) return { start, end };
  return start.getTime() > end.getTime() ? { start: end, end: start } : { start, end };
}
