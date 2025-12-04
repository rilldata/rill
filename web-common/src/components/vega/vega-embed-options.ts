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
  return (items: any, sanitize: (value: any) => string) => {
    const rows = Object.entries(items)
      .map(([key, val]) => {
        if (val === undefined) return "";
        const colorEntry = colorMapping?.find(
          (mapping) => mapping.value === key,
        );
        const keyColor = colorEntry
          ? `<svg  class="key-color">
            <circle cx="6" cy="6" r="6" style="fill:${colorEntry.color};"/>
          </svg>`
          : "";
        return `<tr><td class="key">${keyColor}<span>${sanitize(key)}</span></td><td class="value">${sanitize(val)}</td></tr>`;
      })
      .join("");
    return `<table><tbody>${rows}</tbody></table>`;
  };
}
