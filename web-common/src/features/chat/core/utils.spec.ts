import { extractMessageText } from "@rilldata/web-common/features/chat/core/utils";
import {
  MessageContentType,
  ToolName,
} from "@rilldata/web-common/features/chat/core/types";
import type { V1Message } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";

function routerCall(content: object): V1Message {
  return {
    tool: ToolName.ROUTER_AGENT,
    contentType: MessageContentType.JSON,
    contentData: JSON.stringify(content),
  };
}

describe("extractMessageText", () => {
  it("returns the user prompt for router calls", () => {
    expect(
      extractMessageText(routerCall({ prompt: "Why did revenue drop?" })),
    ).toEqual("Why did revenue drop?");
  });

  it("describes a report's opening call by its explore and time range", () => {
    const text = extractMessageText(
      routerCall({
        prompt: "",
        agent: "analyst_agent",
        analyst_agent_args: {
          prompt: "",
          explore: "requests",
          time_start: "2026-09-09T00:00:00Z",
          time_end: "2026-09-10T00:00:00Z",
          comparison_time_start: "2026-09-08T00:00:00Z",
          comparison_time_end: "2026-09-09T00:00:00Z",
          is_report: true,
        },
      }),
    );
    expect(text).toContain("AI report for requests");
    expect(text).toContain("2026");
    expect(text).toContain("compared with");
    expect(text).not.toContain("analyst_agent_args");
  });

  it("describes the report before its configured prompt", () => {
    const text = extractMessageText(
      routerCall({
        prompt: "Analyze key metrics",
        analyst_agent_args: { explore: "requests", is_report: true },
      }),
    );
    expect(text).toEqual("AI report for requests\n\nAnalyze key metrics");
  });

  it("keeps day boundaries in the report's time zone", () => {
    const text = extractMessageText(
      routerCall({
        prompt: "",
        analyst_agent_args: {
          is_report: true,
          time_start: "2026-09-09T00:00:00-04:00",
          time_end: "2026-09-10T00:00:00-04:00",
        },
      }),
    );
    // Locale-independent: the days stay 9 and 10 and no time suffix is added.
    expect(text).toMatch(/^AI report covering .*9.*10.*2026$/);
    expect(text).not.toContain("(");
  });

  it("describes a report without an explore or time range", () => {
    expect(
      extractMessageText(
        routerCall({ prompt: "", analyst_agent_args: { is_report: true } }),
      ),
    ).toEqual("AI report");
  });

  it("falls back to the raw content for non-report calls without a prompt", () => {
    const content = { prompt: "", analyst_agent_args: { explore: "requests" } };
    expect(extractMessageText(routerCall(content))).toEqual(
      JSON.stringify(content),
    );
  });
});
