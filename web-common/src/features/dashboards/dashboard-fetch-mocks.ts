import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type {
  V1ExploreSpec,
  V1GetExploreResponse,
  V1GetResourceResponse,
  V1MetricsViewAggregationResponse,
  V1MetricsViewSpec,
  V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { afterAll, beforeAll, vi } from "vitest";
import { asyncWait } from "../../lib/waitUtils";

export class DashboardFetchMocks {
  private responses = new Map<string, any>();
  private aggregationRequestMocks: {
    regex: RegExp;
    response: V1MetricsViewAggregationResponse;
  }[] = [];

  public static useDashboardFetchMocks(): DashboardFetchMocks {
    const mock = new DashboardFetchMocks();

    beforeAll(() => {
      vi.stubGlobal("fetch", (url, { body }) => mock.fetchMock(url, body));
    });

    afterAll(() => {
      vi.unstubAllGlobals();
    });

    return mock;
  }

  public mockMetricsView(name: string, resp: V1MetricsViewSpec) {
    this.responses.set(`resource__${name}`, {
      resource: {
        meta: {
          name: {
            kind: ResourceKind.MetricsView,
            name,
          },
        },
        metricsView: {
          state: {
            validSpec: resp,
          },
        },
      },
    } as V1GetResourceResponse);
  }

  public mockMetricsExplore(
    name: string,
    metricsView: V1MetricsViewSpec,
    explore: V1ExploreSpec,
  ) {
    this.responses.set(`resources__explore__${name}`, {
      metricsView: {
        meta: {
          name: {
            kind: ResourceKind.MetricsView,
            name,
          },
        },
        metricsView: {
          state: {
            validSpec: metricsView,
          },
        },
      },
      explore: {
        meta: {
          name: {
            kind: ResourceKind.Explore,
            name,
          },
        },
        explore: {
          state: {
            validSpec: explore,
          },
        },
      },
    } as V1GetExploreResponse);
  }

  public mockTimeRangeSummary(
    metricsViewName: string,
    resp: V1TimeRangeSummary,
  ) {
    this.responses.set(
      `queries__metrics-views__time-range-summary__${metricsViewName}`,
      {
        timeRangeSummary: resp,
      },
    );
  }

  public mockMetricsViewAggregation(
    regex: RegExp,
    response: V1MetricsViewAggregationResponse,
  ) {
    this.aggregationRequestMocks.push({ regex, response });
  }

  public mockMetricsViewTimeRanges(
    metricsViewName: string,
    start: string,
    end: string,
  ) {
    this.responses.set(
      `queries__metrics-views__time-ranges__${metricsViewName}`,
      {
        timeRanges: [{ start, end }],
        resolvedTimeRanges: [{ expression: "PT6H", start, end }],
      },
    );
  }

  private async fetchMock(url: string, body: string | Uint8Array | undefined) {
    const u = new URL(url);

    // ConnectRPC routes: POST to /rill.runtime.v1.{Service}/{Method}
    const connectMatch = u.pathname.match(
      /^\/rill\.runtime\.v1\.(\w+)\/(\w+)$/,
    );
    if (connectMatch) {
      return this.handleConnectRequest(connectMatch[1], connectMatch[2], body);
    }

    const [, , , , type, ...parts] = u.pathname.split("/");
    let key: string;

    switch (type) {
      case "resource":
        key = type + "__" + u.searchParams.get("name.name");
        break;

      case "resources":
        key = type + "__" + parts[0] + "__" + u.searchParams.get("name");
        break;

      case "catalog":
        key = type + "__" + parts[0];
        break;

      case "queries":
        key = type + "__" + parts[0] + "__" + parts[2] + "__" + parts[1];
        break;

      case "metrics-views":
        key = type + "__" + parts[0] + "__" + parts[1];
        break;

      default:
        key = url;
        break;
    }

    // wait a tick
    await asyncWait(1);

    return {
      ready: true,
      ok: true,
      json: () => {
        if (body && parts[2] === "aggregation") {
          const bodyText = typeof body === "string" ? body : "";
          for (const { regex, response } of this.aggregationRequestMocks) {
            if (!regex.test(bodyText)) continue;
            return response;
          }
        }
        return this.responses.get(key);
      },
      headers: new Map([["content-type", "application/json"]]),
    };
  }

  private async handleConnectRequest(
    service: string,
    method: string,
    body: string | Uint8Array | undefined,
  ) {
    await asyncWait(1);

    // ConnectRPC may send body as Uint8Array or other typed array
    let bodyStr: string;
    if (typeof body === "string") {
      bodyStr = body;
    } else if (body && typeof body === "object") {
      bodyStr = new TextDecoder().decode(body as ArrayBufferView);
    } else {
      bodyStr = "";
    }
    const parsed = bodyStr ? JSON.parse(bodyStr) : {};
    let responseData: unknown;

    if (service === "RuntimeService" && method === "GetExplore") {
      responseData = this.responses.get(`resources__explore__${parsed.name}`);
    } else if (
      service === "QueryService" &&
      method === "MetricsViewTimeRange"
    ) {
      const stored = this.responses.get(
        `queries__metrics-views__time-range-summary__${parsed.metricsViewName}`,
      );
      if (stored?.timeRangeSummary) {
        // Convert short dates to RFC 3339 for proto Timestamp compatibility
        const s = stored.timeRangeSummary;
        responseData = {
          timeRangeSummary: {
            ...s,
            min: s.min && !s.min.includes("T") ? s.min + "T00:00:00Z" : s.min,
            max: s.max && !s.max.includes("T") ? s.max + "T00:00:00Z" : s.max,
          },
        };
      } else {
        responseData = stored;
      }
    } else if (
      service === "QueryService" &&
      method === "MetricsViewAggregation"
    ) {
      for (const { regex, response } of this.aggregationRequestMocks) {
        if (regex.test(bodyStr)) {
          responseData = response;
          break;
        }
      }
    }

    // ConnectRPC expects status, ok, headers (Headers), json(), arrayBuffer()
    const jsonBody = responseData ?? {};
    return {
      status: 200,
      ok: true,
      headers: new Headers({ "content-type": "application/json" }),
      json: async () => jsonBody,
      arrayBuffer: async () =>
        new TextEncoder().encode(JSON.stringify(jsonBody)).buffer,
    };
  }
}
