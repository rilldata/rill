import type { ChartDataResult } from "@rilldata/web-common/features/components/charts/types";
import chroma from "chroma-js";

/**
 * Minimal `ChartDataResult` for exercising the cartesian spec builders.
 * `post_title` is the category, `post_count` / `comment_count` are measures
 * and `region` is an optional color dimension.
 */
export function chartData(
  opts: { hasComparison?: boolean; colorValues?: string[] } = {},
): ChartDataResult {
  return {
    data: [],
    isFetching: false,
    fields: {
      post_title: { name: "post_title", displayName: "Post Title" },
      region: { name: "region", displayName: "Region" },
      post_count: { name: "post_count", displayName: "Post Count" },
      comment_count: { name: "comment_count", displayName: "Comment Count" },
    },
    domainValues: {
      post_title: ["A", "B"],
      ...(opts.colorValues && { region: opts.colorValues }),
    },
    isDarkMode: false,
    hasComparison: opts.hasComparison ?? false,
    theme: {
      primary: chroma("#1d4ed8"),
      secondary: chroma("#7c3aed"),
    },
  };
}

/** jsdom has no canvas; vega-lite's compile touches it when measuring text. */
export function stubCanvasContext() {
  if (typeof HTMLCanvasElement === "undefined") return;
  Object.defineProperty(HTMLCanvasElement.prototype, "getContext", {
    configurable: true,
    value: () => null,
  });
}

/**
 * Reads a dotted path (e.g. "layer.1.encoding.x.field") out of a generated
 * spec. Generated specs are typed as the broad Vega-Lite union, and reaching
 * into layers and encodings through that union needs a cast at every step;
 * this keeps the assertions readable without resorting to `any`.
 */
export function at(spec: unknown, path: string): unknown {
  return path.split(".").reduce<unknown>((value, key) => {
    if (value === null || typeof value !== "object") return undefined;
    return (value as Record<string, unknown>)[key];
  }, spec);
}
