import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import type {
  V1GetResourceResponse,
  V1MetricsViewSpec,
  V1TimeRangeSummary,
} from "@rilldata/web-common/runtime-client";
import { afterAll, beforeAll, vi } from "vitest";
import { asyncWait } from "../../lib/waitUtils";

export class DashboardFetchMocks {
  private responses = new Map<string, any>();

  public static useDashboardFetchMocks(): DashboardFetchMocks {
    const mock = new DashboardFetchMocks();

    beforeAll(() => {
      vi.stubGlobal("fetch", (url) => mock.fetchMock(url));
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

  private async fetchMock(url: string) {
    const u = new URL(url);
    const [, , , , type, ...parts] = u.pathname.split("/");
    let key: string;

    switch (type) {
      case "resource":
        key = type + "__" + u.searchParams.get("name.name");
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
      json: () => this.responses.get(key),
    };
  }
}
