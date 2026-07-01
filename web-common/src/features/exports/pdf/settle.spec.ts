import { describe, expect, it } from "vitest";
import type { Readable } from "svelte/store";
import { isCanvasExportQuery, waitForStoreValue } from "./settle";

describe("waitForStoreValue", () => {
  it("handles stores that synchronously reach the target on subscribe", async () => {
    let subscriptions = 0;
    let unsubscriptions = 0;
    const store: Readable<boolean> = {
      subscribe(run) {
        subscriptions += 1;
        run(subscriptions > 1);
        return () => {
          unsubscriptions += 1;
        };
      },
    };

    await expect(waitForStoreValue(store, true, 100)).resolves.toBe(true);
    expect(subscriptions).toBe(2);
    expect(unsubscriptions).toBe(2);
  });
});

describe("isCanvasExportQuery", () => {
  it("matches query service queries for the active runtime instance", () => {
    expect(
      isCanvasExportQuery(
        ["QueryService", "metricsViewAggregation", "instance-1", {}],
        "instance-1",
      ),
    ).toBe(true);
  });

  it("matches custom chart metrics SQL queries", () => {
    expect(
      isCanvasExportQuery(["metrics_sql", "chart", 0, "select 1", "{}"], "i"),
    ).toBe(true);
  });

  it("ignores unrelated admin and other runtime queries", () => {
    expect(
      isCanvasExportQuery(
        ["AdminService", "getProject", { project: "p" }],
        "instance-1",
      ),
    ).toBe(false);
    expect(
      isCanvasExportQuery(
        ["RuntimeService", "gitStatus", "instance-1", {}],
        "instance-1",
      ),
    ).toBe(false);
    expect(
      isCanvasExportQuery(
        ["QueryService", "metricsViewAggregation", "instance-2", {}],
        "instance-1",
      ),
    ).toBe(false);
  });
});
