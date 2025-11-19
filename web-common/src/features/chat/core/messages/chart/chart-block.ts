import type { V1Message } from "@rilldata/web-common/runtime-client";
import { MessageContentType } from "../../types";
import { parseChartData } from "../../utils";

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
