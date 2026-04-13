import { getRillTheme } from "@rilldata/web-common/components/vega/vega-config";
import { ScrubBoxColor } from "@rilldata/web-common/features/dashboards/time-series/chart-colors";
import type { VegaLiteSpec, VegaSpec, VisualizationSpec } from "svelte-vega";
import type { Signal } from "vega";
import { compile } from "vega-lite";
import type { SelectionParameter } from "vega-lite/types_unstable/selection.js";

/**
 * Creates a Vega-Lite brush parameter for interval selection on the x-axis.
 * Add this to a layer's `params` array to enable brush/scrub selection.
 */
export function createBrushParam(): SelectionParameter {
  return {
    name: "brush",
    select: {
      type: "interval",
      encodings: ["x"],
      mark: {
        fill: ScrubBoxColor,
        fillOpacity: 0.4,
        stroke: ScrubBoxColor,
        strokeWidth: 1,
        strokeOpacity: 0.8,
      },
    },
  };
}

/**
 * Compiles a Vega-Lite spec (that includes a brush param) to a Vega spec,
 * then injects custom signals for brush_end (pointerup) and brush_clear (Escape).
 *
 * Vega-Lite's interval selection doesn't natively support detecting when a brush
 * drag ends or when the user presses Escape to clear. These custom signals fill that gap.
 *
 * See: https://github.com/vega/vega-lite/issues/5341
 * See: https://github.com/vega/vega-lite/issues/3338
 */
export async function compileToBrushedVegaSpec(
  vlSpec: VisualizationSpec,
  isThemeModeDark: boolean,
  theme: Record<string, string> | undefined,
): Promise<{ spec: VegaSpec; temporalBrushSignal: string }> {
  const existingConfig =
    (vlSpec as { config?: Record<string, unknown> }).config ?? {};
  // Merge the Rill theme config so axis settings (e.g. grid: false) are baked
  // into the compiled Vega spec
  const rillThemeConfig = getRillTheme(isThemeModeDark, theme);
  const specWithConfig = {
    ...vlSpec,
    config: {
      ...rillThemeConfig,
      ...existingConfig,
      customFormatTypes: true,
    },
  };
  const compiledSpec = compile(specWithConfig as VegaLiteSpec).spec;
  const originalSignals = compiledSpec.signals || [];

  // Resolve the temporal brush signal name.
  // Vega-Lite generates `brush_[timeunit_]<fieldname>` as the reactive temporal
  // signal (e.g. brush_yearmonthdatehours___time for field __time with timeUnit).
  const usermetaField = (
    compiledSpec.usermeta as { brushTemporalField?: string } | undefined
  )?.brushTemporalField;

  const temporalBrushSignal =
    (usermetaField &&
      originalSignals.find(
        (s) => s.name?.startsWith("brush_") && s.name.endsWith(usermetaField),
      )?.name) ??
    "brush_ts";

  const updatedSignals = originalSignals.map((signal: Signal): Signal => {
    if (signal.name === "brush_x") {
      return {
        ...signal,
        value: [],
        on: [
          { events: { signal: "brush_clear" }, update: "[0, 0]" },
          ...(signal.on || []),
        ],
      };
    }

    // Add clear handler to the temporal brush signal (e.g. brush_ts, brush_event_time)
    if (signal.name === temporalBrushSignal) {
      return {
        ...signal,
        on: [
          { events: { signal: "brush_clear" }, update: "null" },
          ...(signal.on || []),
        ],
      };
    }

    return signal;
  });

  // Signal that fires on pointerup to detect brush drag end.
  // References the reactive temporal interval signal (e.g. brush_ts) instead of
  // brush, because in Vega-Lite 6 the `brush` signal is a static computed expression
  // with no `on` clause — it is evaluated once at init and never re-fires.
  // Uses source: "window" (not "scope") because Vega-Lite 6 registers brush_x's
  // drag-ending pointerup at the window level, which is where the event fires
  // even when the mouse is released outside the chart bounds.
  updatedSignals.push({
    name: "brush_end",
    on: [
      {
        events: { source: "window", type: "pointerup" },
        update: temporalBrushSignal,
      },
    ],
  });

  // Signal that fires on Escape key to clear the brush
  updatedSignals.push({
    name: "brush_clear",
    on: [
      {
        events: {
          source: "window",
          type: "keydown",
          filter: ["event.key === 'Escape'"],
        },
        update: temporalBrushSignal,
      },
    ],
  });

  return {
    spec: { ...compiledSpec, signals: updatedSignals },
    temporalBrushSignal,
  };
}

/**
 * Checks whether a Vega-Lite spec contains a brush parameter,
 * indicating it needs compilation to Vega for brush signal support.
 */
export function hasBrushParam(spec: unknown): boolean {
  if (!spec || typeof spec !== "object") return false;

  const layers = (spec as Record<string, unknown>).layer;
  if (Array.isArray(layers)) {
    return layers.some(
      (layer) =>
        layer?.params?.some?.((p: { name: string }) => p.name === "brush") ??
        false,
    );
  }

  const params = (spec as Record<string, unknown>).params;
  if (Array.isArray(params)) {
    return params.some((p: { name: string }) => p.name === "brush");
  }

  return false;
}

/**
 * Creates an adaptive scrub handler that throttles brush updates based on
 * rendering performance. Adjusts between 30-120fps dynamically.
 *
 * @param onBrush - Called with the brush interval on each throttled update
 * @returns Object with `update` method and `destroy` cleanup method
 */
export function createAdaptiveScrubHandler(
  onBrush: (interval: { start: Date; end: Date }) => void,
) {
  let rafId: number | null = null;
  let lastUpdateTime = 0;
  let currentInterval = 1000 / 60; // Start at 60fps

  const MIN_INTERVAL = 1000 / 120; // Max 120fps
  const MAX_INTERVAL = 1000 / 30; // Min 30fps
  const ADJUSTMENT_FACTOR = 1.2;

  function update(interval: { start: Date; end: Date }) {
    if (rafId) {
      cancelAnimationFrame(rafId);
    }

    rafId = requestAnimationFrame((timestamp) => {
      const elapsed = timestamp - lastUpdateTime;
      if (elapsed >= currentInterval) {
        onBrush(interval);
        lastUpdateTime = timestamp;

        // Adjust interval based on performance
        if (elapsed > currentInterval * ADJUSTMENT_FACTOR) {
          currentInterval = Math.min(
            currentInterval * ADJUSTMENT_FACTOR,
            MAX_INTERVAL,
          );
        } else {
          currentInterval = Math.max(
            currentInterval / ADJUSTMENT_FACTOR,
            MIN_INTERVAL,
          );
        }
      }
      rafId = null;
    });
  }

  function destroy() {
    if (rafId) {
      cancelAnimationFrame(rafId);
      rafId = null;
    }
  }

  return { update, destroy };
}
