// @vitest-environment jsdom
import { render, screen } from "@testing-library/svelte";
import { V1AssertionStatus } from "@rilldata/web-common/runtime-client/gen/index.schemas";
import { describe, expect, it } from "vitest";
import AlertHistoryTableCompositeCell from "./AlertHistoryTableCompositeCell.svelte";

const props = {
  alertTime: "2026-10-12T05:00:00Z",
  timeZone: "UTC",
  currentExecution: null,
  result: { status: V1AssertionStatus.ASSERTION_STATUS_FAIL },
};

describe("AlertHistoryTableCompositeCell", () => {
  it("is a link to the check's data when it has one", () => {
    render(AlertHistoryTableCompositeCell, {
      props: { ...props, openHref: "/org/proj/-/alerts/a/open" },
    });
    expect(screen.getByRole("link").getAttribute("href")).toBe(
      "/org/proj/-/alerts/a/open",
    );
  });

  it("is plain text otherwise", () => {
    render(AlertHistoryTableCompositeCell, { props });
    expect(screen.queryByRole("link")).toBeNull();
  });
});
