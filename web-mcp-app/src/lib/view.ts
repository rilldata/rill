import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
import {
  hasChartData,
  parseChartPayload,
  textContent,
  type ChartResult,
  type ToolResult,
} from "./contract";

export type InterpretedResult =
  | { kind: "chart"; result: ChartResult }
  | { kind: "error"; message: string; openUrl?: string };

/**
 * Decides what a tool result means for the UI before any data processing happens:
 * a renderable chart, or an error with the best available deep link.
 * A result without `data` is the shape Claude produces when a large result is written to a
 * file instead of being forwarded, so it gets an explicit message; a query without rows
 * carries `data: []` and renders as the empty state.
 */
export function interpretToolResult(result: ToolResult): InterpretedResult {
  if (result.isError) {
    return {
      kind: "error",
      message: textContent(result) || m.mcp_chart_error_tool_failed(),
    };
  }
  const parsed = parseChartPayload(result.structuredContent);
  switch (parsed.kind) {
    case "unsupported_widget":
      return {
        kind: "error",
        message: m.mcp_chart_error_unsupported_widget({
          widget: parsed.widget,
        }),
        openUrl: parsed.openUrl,
      };
    case "invalid":
      return {
        kind: "error",
        message: m.mcp_chart_error_invalid_result(),
        openUrl: parsed.openUrl,
      };
    case "chart":
      if (!hasChartData(parsed.result)) {
        return {
          kind: "error",
          message: m.mcp_chart_error_no_data(),
          openUrl: parsed.result.open_url,
        };
      }
      return parsed;
  }
}
