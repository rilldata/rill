import type { ColorMapping } from "@rilldata/web-common/features/components/charts/types";
import { ComparisonDeltaPreviousSuffix } from "@rilldata/web-common/features/dashboards/filters/measure-filters/measure-filter-entry";
import type { RuntimeClient } from "@rilldata/web-common/runtime-client/v2";
import type { EmbedOptions } from "svelte-vega";
import { expressionInterpreter } from "vega-interpreter";
import type { Config } from "vega-lite";
import type { ExpressionFunction } from "./types";
import { getRillTheme } from "./vega-config";

export interface CreateEmbedOptionsParams {
  client: RuntimeClient;
  canvasDashboard: boolean;
  width: number;
  height: number;
  config?: Config;
  renderer?: "canvas" | "svg";
  themeMode?: "light" | "dark";
  expressionFunctions?: ExpressionFunction;
  useExpressionInterpreter?: boolean;
  colorMapping: ColorMapping;
  hasComparison?: boolean;
}

export function createEmbedOptions({
  client,
  canvasDashboard,
  width,
  height,
  config,
  renderer = "canvas",
  themeMode = "light",
  expressionFunctions = {},
  useExpressionInterpreter = true,
  colorMapping,
  hasComparison,
}: CreateEmbedOptionsParams): EmbedOptions {
  const jwt = client.getJwt();

  return {
    config: config || getRillTheme(canvasDashboard, themeMode === "dark"),
    renderer,
    tooltip: {
      theme: themeMode,
      ...(hasComparison || colorMapping?.length
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
      baseURL: `${client.host}/v1/instances/${client.instanceId}/assets/`,
      ...(jwt && {
        http: {
          headers: {
            Authorization: `Bearer ${jwt}`,
          },
        },
      }),
    },
  };
}

export function getTooltipFormatter(colorMapping: ColorMapping) {
  const colorMap = new Map<string, string>(
    (colorMapping ?? []).map((m) => [m.value, m.color]),
  );

  return (
    items: Record<string, unknown>,
    sanitize: (value: unknown) => string,
  ) => {
    const groupedItems = new Map<
      string,
      { current?: unknown; previous?: unknown }
    >();
    const nonComparisonItems: Array<[string, unknown]> = [];
    let headerValue: string | null = null;
    let hasComparison = false;

    for (const [key, val] of Object.entries(items)) {
      if (val === undefined) continue;

      if (key.endsWith(ComparisonDeltaPreviousSuffix)) {
        const baseKey = key.slice(0, -ComparisonDeltaPreviousSuffix.length);
        const existing = groupedItems.get(baseKey) || {};
        groupedItems.set(baseKey, { ...existing, previous: val });
        hasComparison = true;
      } else {
        const prevKey = key + ComparisonDeltaPreviousSuffix;
        if (prevKey in items) {
          const existing = groupedItems.get(key) || {};
          groupedItems.set(key, { ...existing, current: val });
          hasComparison = true;
        } else {
          // Standalone field: first string becomes header, don't add it to rows
          if (headerValue === null && typeof val === "string") {
            headerValue = sanitize(val);
          } else {
            nonComparisonItems.push([key, val]);
          }
        }
      }
    }

    const rows: string[] = [];

    // Header row (if any)
    if (headerValue) {
      rows.push(
        `<tr><td colspan="10" style="text-align: left; font-weight: 600; padding-bottom: 4px;">${headerValue}</td></tr>`,
      );
    }

    // Helper: key color SVG (if present)
    const keyColorSvg = (color?: string) =>
      color
        ? `<svg class="key-color"><circle cx="6" cy="6" r="6" style="fill:${color};"/></svg>`
        : "";

    // Non-comparison rows first
    for (const [key, val] of nonComparisonItems) {
      const color = colorMap.get(key);
      const keyHtml = `<td class="key">${keyColorSvg(color)}<span>${sanitize(key)}</span></td>`;
      const valHtml = `<td class="value">${sanitize(val)}</td>`;
      if (hasComparison) {
        rows.push(
          `<tr>${keyHtml}${valHtml}<td class="value empty-cell"></td></tr>`,
        );
      } else {
        rows.push(`<tr>${keyHtml}${valHtml}</tr>`);
      }
    }

    // Grouped comparison (or single) rows
    for (const [key, values] of groupedItems.entries()) {
      const color = colorMap.get(key);
      const keyHtml = `<td class="key">${keyColorSvg(color)}<span>${sanitize(key)}</span></td>`;

      if (hasComparison) {
        const currentValue =
          values.current !== undefined ? sanitize(values.current) : "";
        const previousValue =
          values.previous !== undefined ? sanitize(values.previous) : "";
        rows.push(
          `<tr>${keyHtml}<td class="value current-value">${currentValue}</td><td class="value previous-value">${previousValue}</td></tr>`,
        );
      } else {
        const valueHtml =
          values.current !== undefined ? sanitize(values.current) : "";
        rows.push(`<tr>${keyHtml}<td class="value">${valueHtml}</td></tr>`);
      }
    }

    return `<table><tbody>${rows.join("")}</tbody></table>`;
  };
}
