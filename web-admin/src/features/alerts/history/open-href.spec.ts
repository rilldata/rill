import { V1AssertionStatus } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { describe, expect, it } from "vitest";
import { alertExecutionOpenHref } from "./open-href";

describe("alertExecutionOpenHref", () => {
  it("opens a triggered check at its execution time, like the email link", () => {
    expect(
      alertExecutionOpenHref("org", "proj", "devoluciones-altas", {
        executionTime: "2026-10-12T05:00:00Z",
        result: { status: V1AssertionStatus.ASSERTION_STATUS_FAIL },
      }),
    ).toBe(
      "/org/proj/-/alerts/devoluciones-altas/open?execution_time=2026-10-12T05%3A00%3A00Z",
    );
  });

  it("leaves passed and failed-to-run checks without a link", () => {
    for (const status of [
      V1AssertionStatus.ASSERTION_STATUS_PASS,
      V1AssertionStatus.ASSERTION_STATUS_ERROR,
    ]) {
      expect(
        alertExecutionOpenHref("org", "proj", "a", {
          executionTime: "2026-10-12T05:00:00Z",
          result: { status },
        }),
      ).toBeUndefined();
    }
  });
});
