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
  expressionFunctions?: ExpressionFunction;
  useExpressionInterpreter?: boolean;
}

export function createEmbedOptions({
  canvasDashboard,
  width,
  height,
  config,
  renderer = "canvas",
  expressionFunctions = {},
  useExpressionInterpreter = true,
}: CreateEmbedOptionsParams): EmbedOptions {
  const jwt = get(runtime).jwt;

  return {
    config: config || getRillTheme(canvasDashboard),
    renderer,
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
