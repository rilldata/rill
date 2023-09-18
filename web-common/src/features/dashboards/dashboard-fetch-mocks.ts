import type { V1TimeRangeSummary } from "@rilldata/web-common/runtime-client";
import type { V1MetricsView } from "@rilldata/web-common/runtime-client";
import { wait } from "@testing-library/user-event/dist/utils";
import { afterAll, beforeAll, vi } from "vitest";

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

  public mockMetricsView(name: string, resp: V1MetricsView) {
    this.responses.set(`catalog__${name}`, {
      entry: {
        metricsView: resp,
      },
    });
  }

  public mockTimeRangeSummary(
    tableName: string,
    columnName: string,
    resp: V1TimeRangeSummary
  ) {
    this.responses.set(
      `queries__time-range-summary__${tableName}__${columnName}`,
      {
        timeRangeSummary: resp,
      }
    );
  }

  private async fetchMock(url: string) {
    const u = new URL(url);
    const [, , , , type, ...parts] = u.pathname.split("/");
    let key: string;

    switch (type) {
      case "catalog":
        key = type + "__" + parts[0];
        break;

      case "queries":
        key =
          type +
          "__" +
          parts[0] +
          "__" +
          parts[2] +
          "__" +
          (u.searchParams.get("columnName") ?? "");
        break;

      case "metrics-views":
        key = type + "__" + parts[0] + "__" + parts[1];
        break;

      default:
        key = url;
        break;
    }

    await wait(1);

    return {
      ready: true,
      ok: true,
      json: () => this.responses.get(key),
    };
  }
}
