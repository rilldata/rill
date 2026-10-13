import { maybeGetMetricsResolverQueryFromMessage } from "@rilldata/web-common/features/chat/core/citation-url-utils.ts";
import {
  MessageType,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";
import type { V1Message } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

function toolCallMessage(
  tool: string,
  input: Record<string, unknown> | undefined,
  type: string = MessageType.CALL,
): V1Message {
  return {
    id: "msg-1",
    role: "assistant",
    type,
    tool,
    contentType: "json",
    contentData: JSON.stringify(input),
    content: input ? [{ toolCall: { id: "call-1", name: tool, input } }] : [],
  };
}

const query = {
  metrics_view: "bids_metrics",
  dimensions: [{ name: "advertiser_name" }],
  measures: [{ name: "total_bids" }],
  time_range: { start: "2024-01-01T00:00:00Z", end: "2024-02-01T00:00:00Z" },
};

describe("maybeGetMetricsResolverQueryFromMessage", () => {
  it("returns the arguments of a query_metrics_view call as is", () => {
    const message = toolCallMessage(ToolName.QUERY_METRICS_VIEW, query);
    expect(maybeGetMetricsResolverQueryFromMessage(message)).toEqual(query);
  });

  it("maps the spec of a create_chart call onto a query", () => {
    const message = toolCallMessage(ToolName.CREATE_CHART, {
      chart_type: "line_chart",
      spec: {
        metrics_view: "bids_metrics",
        time_range: query.time_range,
        time_grain: "TIME_GRAIN_WEEK",
        where: {
          cond: { op: "eq", exprs: [{ name: "device" }, { val: "mobile" }] },
        },
        x: { field: "__time", type: "temporal" },
        y: { field: "total_bids", type: "quantitative" },
        color: { field: "advertiser_name", type: "nominal", limit: 5 },
      },
    });
    expect(maybeGetMetricsResolverQueryFromMessage(message)).toEqual({
      metrics_view: "bids_metrics",
      dimensions: [
        {
          name: "__time",
          compute: { time_floor: { dimension: "__time", grain: "week" } },
        },
        { name: "advertiser_name" },
      ],
      measures: [{ name: "total_bids" }],
      time_range: query.time_range,
      where: {
        cond: { op: "eq", exprs: [{ name: "device" }, { val: "mobile" }] },
      },
    });
  });

  it("rejects a create_chart call without a spec", () => {
    const message = toolCallMessage(ToolName.CREATE_CHART, {
      chart_type: "bar_chart",
    });
    expect(() => maybeGetMetricsResolverQueryFromMessage(message)).toThrow(
      expect.objectContaining({ status: 400 }),
    );
  });

  it("rejects calls to other tools", () => {
    const message = toolCallMessage(ToolName.GET_METRICS_VIEW, {
      metrics_view: "bids_metrics",
    });
    expect(() => maybeGetMetricsResolverQueryFromMessage(message)).toThrow(
      expect.objectContaining({ status: 400 }),
    );
  });

  it("rejects result messages", () => {
    const message = toolCallMessage(
      ToolName.QUERY_METRICS_VIEW,
      query,
      MessageType.RESULT,
    );
    expect(() => maybeGetMetricsResolverQueryFromMessage(message)).toThrow(
      expect.objectContaining({ status: 400 }),
    );
  });

  it("rejects calls without content", () => {
    const message = toolCallMessage(ToolName.QUERY_METRICS_VIEW, undefined);
    expect(() => maybeGetMetricsResolverQueryFromMessage(message)).toThrow(
      expect.objectContaining({ status: 400 }),
    );
  });
});
