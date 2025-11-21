import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageContentType } from "../../types";

/**
 * Chart block representation.
 * Contains a chart visualization extracted from a create_chart tool call.
 */
export type ChartBlock = {
  type: "chart";
  id: string;
  message: V1Message;
  chartData: any;
};

/**
 * Creates a chart block from a create_chart tool call message.
 * Returns null if the chart data is invalid or the result indicates an error.
 */
export function createChartBlock(
  message: V1Message,
  resultMessage: V1Message | undefined,
): ChartBlock | null {
  const hasResult = !!resultMessage;
  const isError = resultMessage?.contentType === MessageContentType.ERROR;
  const chartData = parseChartData({ input: message.contentData });

  if (!chartData || !hasResult || isError) {
    return null;
  }

  return {
    type: "chart",
    id: `chart-${message.id}`,
    message,
    chartData,
  };
}

/**
 * Parses chart data from a tool call.
 * Returns null if parsing fails.
 */
function parseChartData(toolCall: any) {
  try {
    // Check if input is already an object or needs parsing
    const parsed =
      typeof toolCall?.input === "string"
        ? JSON.parse(toolCall.input)
        : toolCall?.input;

    return {
      chartType: parsed.chart_type,
      chartSpec: parsed.spec,
    };
  } catch (error) {
    console.error("Failed to parse chart data:", error);
    return null;
  }
}
