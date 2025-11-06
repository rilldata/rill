import type { ColorMapping } from "@rilldata/web-common/features/components/charts/types";
import { runtime } from "@rilldata/web-common/runtime-client/runtime-store";
import type { EmbedOptions } from "svelte-vega";
import { get } from "svelte/store";
import { expressionInterpreter } from "vega-interpreter";
import type { Config } from "vega-lite";
import type { ExpressionFunction } from "./types";
import { getRillTheme } from "./vega-config";

export interface CreateEmbedOptionsParams {
  canvasDashboard: boolean;
  width: number;
  height: number;
  config?: Config;
  renderer?: "canvas" | "svg";
  themeMode?: "light" | "dark";
  expressionFunctions?: ExpressionFunction;
  useExpressionInterpreter?: boolean;
  colorMapping: ColorMapping;
}

export function createEmbedOptions({
  canvasDashboard,
  width,
  height,
  config,
  renderer = "canvas",
  themeMode = "light",
  expressionFunctions = {},
  useExpressionInterpreter = true,
  colorMapping,
}: CreateEmbedOptionsParams): EmbedOptions {
  const jwt = get(runtime).jwt;

  return {
    config: config || getRillTheme(canvasDashboard, themeMode === "dark"),
    renderer,
    tooltip: {
      theme: themeMode,
      ...(colorMapping?.length
        ? { formatTooltip: getTooltipFormatter(colorMapping) }
        : {}),
    },
    actions: false,
    logLevel: 0, // only show errors
    width: canvasDashboard ? width : undefined,
    height: canvasDashboard ? height : undefined,
    ...(useExpressionInterpreter && {
      // Add interpreter so that vega expressions are CSP compliant
      ast: true,
      expr: expressionInterpreter,
    }),
    expressionFunctions,
    loader: {
      baseURL: `${get(runtime).host}/v1/instances/${get(runtime).instanceId}/assets/`,
      ...(jwt &&
        jwt.token && {
          http: {
            headers: {
              Authorization: `Bearer ${jwt.token}`,
            },
          },
        }),
    },
  };
}

export function getTooltipFormatter(colorMapping: ColorMapping) {
  return (
    items: Record<string, unknown>,
    sanitize: (value: unknown) => string,
  ) => {
    // Group items to combine current and previous values
    const groupedItems = new Map<
      string,
      { current?: unknown; previous?: unknown }
    >();
    const nonComparisonItems: Array<[string, unknown]> = [];

    for (const [key, val] of Object.entries(items)) {
      if (val === undefined) continue;

      // Check if this is a comparison field (ends with _prev)
      if (key.endsWith("_prev")) {
        const baseKey = key.slice(0, -1 * "_prev".length);
        const existing = groupedItems.get(baseKey) || {};
        groupedItems.set(baseKey, { ...existing, previous: val });
      } else {
        // Check if there's a corresponding _prev field in items
        const prevKey = key + "_prev";
        if (prevKey in items) {
          // This is a current value with a comparison
          const existing = groupedItems.get(key) || {};
          groupedItems.set(key, { ...existing, current: val });
        } else {
          // This is a standalone field (not part of comparison)
          nonComparisonItems.push([key, val]);
        }
      }
    }

    const rows: string[] = [];

    // Add non-comparison items first (like x-axis field)
    for (const [key, val] of nonComparisonItems) {
      const colorEntry = colorMapping?.find((mapping) => mapping.value === key);
      const keyColor = colorEntry
        ? `<svg class="key-color">
          <circle cx="6" cy="6" r="6" style="fill:${colorEntry.color};"/>
        </svg>`
        : "";
      rows.push(
        `<tr><td class="key">${keyColor}<span>${sanitize(key)}</span></td><td class="value">${sanitize(val)}</td></tr>`,
      );
    }

    // Check if any grouped items have comparison values
    const hasComparison = Array.from(groupedItems.values()).some(
      (v) => v.previous !== undefined,
    );

    // Add grouped comparison items
    for (const [key, values] of groupedItems.entries()) {
      const colorEntry = colorMapping?.find((mapping) => mapping.value === key);
      const keyColor = colorEntry
        ? `<svg class="key-color">
          <circle cx="6" cy="6" r="6" style="fill:${colorEntry.color};"/>
        </svg>`
        : "";

      if (hasComparison) {
        // Use separate columns for current and comparison values
        const currentValue =
          values.current !== undefined ? sanitize(values.current) : "";
        const previousValue =
          values.previous !== undefined ? sanitize(values.previous) : "";

        rows.push(
          `<tr><td class="key">${keyColor}<span>${sanitize(key)}</span></td><td class="value current-value">${currentValue}</td><td class="value previous-value">${previousValue}</td></tr>`,
        );
      } else {
        // Single value column when no comparison
        const valueHtml =
          values.current !== undefined ? sanitize(values.current) : "";
        rows.push(
          `<tr><td class="key">${keyColor}<span>${sanitize(key)}</span></td><td class="value">${valueHtml}</td></tr>`,
        );
      }
    }

    return `<table><tbody>${rows.join("")}</tbody></table>`;
  };
}
