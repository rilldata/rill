import type { Page } from "@sveltejs/kit";
import { describe, it, expect, beforeAll } from "vitest";
import type { MetricsViewAndExploreSpecs } from "@rilldata/web-common/features/entity-management/resource-selectors.ts";
import { mapMetricsResolverQueryToUrl } from "@rilldata/web-common/features/chat/core/messages/text/citation-url-mapper.ts";
import {
  getQueryServiceMetricsViewTimeRangeQueryKey,
  type V1Message,
} from "@rilldata/web-common/runtime-client";
import {
  AD_BIDS_EXPLORE_NAME,
  AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_3_MEASURES_DIMENSIONS,
  AD_BIDS_METRICS_NAME,
  AD_BIDS_TIME_RANGE_SUMMARY,
} from "@rilldata/web-common/features/dashboards/stores/test-data/data.ts";
import { queryClient } from "@rilldata/web-common/lib/svelte-query/globalQueryClient.ts";
import {
  MessageType,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types.ts";

const MetricsResolverQuery = {
  metrics_view: "AdBids_metrics",
  dimensions: [{ name: "publisher" }],
  measures: [{ name: "impressions" }],
};
const MetricsResolverQueryURLEncoded = encodeURIComponent(
  JSON.stringify(MetricsResolverQuery),
);

describe("mapMetricsResolverQueryToUrlParams", () => {
  beforeAll(() => {
    const metricsViewTimeRangeQueryKey =
      getQueryServiceMetricsViewTimeRangeQueryKey("", AD_BIDS_METRICS_NAME, {});
    queryClient.setQueryData(metricsViewTimeRangeQueryKey, {
      timeRangeSummary: AD_BIDS_TIME_RANGE_SUMMARY,
    });
  });

  const testCases: Array<{
    title: string;
    url: string;
    expectedUrl: string;
  }> = [
    {
      title: "Not a citation url",
      url: "https://random.citation.url/",
      expectedUrl: "https://random.citation.url/",
    },

    {
      title: "Legacy citation url without query param",
      url: "http://localhost:3000/-/open-query",
      expectedUrl: "http://localhost:3000/-/open-query",
    },
    {
      title: "Legacy citation url with valid query param",
      url:
        "http://localhost:3000/-/open-query?query=" +
        MetricsResolverQueryURLEncoded,
      expectedUrl:
        "http://localhost:3000/explore/AdBids_explore?view=explore&measures=impressions&dims=publisher&expand_dim=publisher",
    },
    {
      title: "Legacy citation url with additional bracket",
      url:
        "http://localhost:3000/-/open-query?query=" +
        MetricsResolverQueryURLEncoded +
        "%7D",
      expectedUrl:
        "http://localhost:3000/explore/AdBids_explore?view=explore&measures=impressions&dims=publisher&expand_dim=publisher",
    },

    {
      title: "Tool call based citation url with missing call",
      url: "http://localhost:3000/-/ai/sess/message/missing_call/-/open",
      expectedUrl:
        "http://localhost:3000/-/ai/sess/message/missing_call/-/open",
    },
    {
      title: "Tool call based citation url with wrong call",
      url: "http://localhost:3000/-/ai/sess/message/agent_call/-/open",
      expectedUrl: "http://localhost:3000/-/ai/sess/message/agent_call/-/open",
    },
    {
      title: "Tool call based citation url with result",
      url: "http://localhost:3000/-/ai/sess/message/query_res/-/open",
      expectedUrl: "http://localhost:3000/-/ai/sess/message/query_res/-/open",
    },
    {
      title: "Tool call based citation url with the correct call",
      url: "http://localhost:3000/-/ai/sess/message/query_call/-/open",
      expectedUrl:
        "http://localhost:3000/explore/AdBids_explore?view=explore&measures=impressions&dims=publisher&expand_dim=publisher",
    },
  ];

  const MockMetricsViewAndExploreSpecs: MetricsViewAndExploreSpecs = {
    metricsViewSpecsMap: new Map([
      [AD_BIDS_METRICS_NAME, AD_BIDS_METRICS_3_MEASURES_DIMENSIONS],
    ]),
    exploreSpecsMap: new Map([
      [AD_BIDS_EXPLORE_NAME, AD_BIDS_EXPLORE_WITH_3_MEASURES_DIMENSIONS],
    ]),
    exploreForMetricViewsMap: new Map([
      [AD_BIDS_METRICS_NAME, AD_BIDS_EXPLORE_NAME],
    ]),
  };

  const mockPage = {
    url: new URL("http://localhost:3000"),
    params: {},
    route: {},
  } as Page;

  const Messages: V1Message[] = [
    { id: "agent_call", type: MessageType.CALL, tool: ToolName.ANALYST_AGENT },
    {
      id: "query_call",
      type: MessageType.CALL,
      tool: ToolName.QUERY_METRICS_VIEW,
      content: [{ toolCall: { input: MetricsResolverQuery } }],
    },
    {
      id: "query_res",
      type: MessageType.RESULT,
      tool: ToolName.QUERY_METRICS_VIEW,
    },
    { id: "agent_res", type: MessageType.RESULT, tool: ToolName.ANALYST_AGENT },
  ];

  for (const { title, url, expectedUrl } of testCases) {
    it(title, () => {
      const result = mapMetricsResolverQueryToUrl(
        new URL(url),
        mockPage,
        MockMetricsViewAndExploreSpecs,
        Messages,
      );
      expect(result).toEqual(expectedUrl);
    });
  }
});
