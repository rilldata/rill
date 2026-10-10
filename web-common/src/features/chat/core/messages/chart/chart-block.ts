import type { ChartType } from "@rilldata/web-common/features/components/charts";
import { getDefaultTimeZone } from "@rilldata/web-common/features/dashboards/stores/get-rill-default-explore-state";
import { getUTCIANA } from "@rilldata/web-common/lib/time/timezone";
import type {
  V1ExploreSpec,
  V1Message,
} from "@rilldata/web-common/runtime-client";
import { MessageContentType } from "../../types";

// =============================================================================
// BACKEND TYPES (mirror runtime/ai tool definitions)
// =============================================================================

/** Arguments for the create_chart tool call */
interface CreateChartCallData {
  chart_type: ChartType;
  spec: unknown;
}

// =============================================================================
// BLOCK TYPE
// =============================================================================

/**
 * Chart block representation.
 * Contains a chart visualization extracted from a create_chart tool call.
 */
export type ChartBlock = {
  type: "chart";
  id: string;
  message: V1Message;
  resultMessage: V1Message;
  chartType: ChartType;
  chartSpec: unknown;
};

/**
 * Creates a chart block from a create_chart tool call message.
 * Returns null if the chart data is invalid or the result indicates an error.
 */
export function createChartBlock(
  message: V1Message,
  resultMessage: V1Message | undefined,
): ChartBlock | null {
  if (!resultMessage) return null;
  if (resultMessage.contentType === MessageContentType.ERROR) return null;

  const callData = parseCallData(message.contentData);
  if (!callData) return null;

  return {
    type: "chart",
    id: `chart-${message.id}`,
    message,
    resultMessage,
    chartType: callData.chart_type,
    chartSpec: callData.spec,
  };
}

/**
 * Resolves the timezone used to bin and label a chat chart.
 *
 * Explicit `time_range.time_zone` in the spec is authoritative.
 * Otherwise inherit the explore default (`defaultPreset.timezone`, else
 * `timeZones[0]`, else UTC). Missing explore context (embed, still loading,
 * no dashboard for the metrics view) degrades to UTC.
 */
export function resolveChartTimeZone(
  explicitTimeZone: string | undefined,
  explore: V1ExploreSpec | undefined,
): string {
  if (explicitTimeZone) return explicitTimeZone;
  if (!explore) return getUTCIANA();
  return getDefaultTimeZone(explore);
}

// =============================================================================
// HELPERS
// =============================================================================

/**
 * Parses the create_chart tool call data.
 */
function parseCallData(
  contentData: string | undefined,
): CreateChartCallData | null {
  try {
    const parsed =
      typeof contentData === "string" ? JSON.parse(contentData) : contentData;
    if (!parsed?.chart_type) return null;
    return parsed as CreateChartCallData;
  } catch {
    return null;
  }
}
