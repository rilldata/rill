import { cleanup, render, screen } from "@testing-library/svelte";
import { Interval } from "luxon";
import { afterEach, describe, expect, it } from "vitest";
import KPITest from "./KPITest.svelte";

afterEach(cleanup);

const measure = {
  name: "total_records",
  displayName: "Total records",
};

const idle = { isError: false, isLoading: false, isFetching: false };

function renderKPI(size?: unknown) {
  render(KPITest, {
    kpiProps: {
      measure,
      size,
      sparkline: "none",
      comparisonOptions: [],
      showTimeComparison: false,
      hasTimeSeries: true,
      hideTimeRange: true,
      comparisonLabel: undefined,
      timeGrain: undefined,
      timeZone: "UTC",
      interval: Interval.invalid("unset"),
      primaryTotalResult: {
        ...idle,
        data: { data: [{ total_records: 1234 }] },
      },
      comparisonTotalResult: { ...idle, data: undefined },
      primarySparklineResult: { ...idle, data: undefined },
      comparisonSparklineResult: { ...idle, data: undefined },
      // The component's query props are typed as full TanStack results.
    } as never,
  });

  const data = screen.getByRole("button", { name: "total_records KPI data" });
  const card = data.closest(".wrapper");
  return { data, card };
}

describe("KPI size", () => {
  it("renders at the medium size by default", () => {
    const { card, data } = renderKPI();
    expect(card).toHaveAttribute("data-size", "medium");
    expect(data).toHaveTextContent("1,234");
  });

  it.each(["small", "medium", "large", "xlarge"])(
    "renders the %s preset",
    (size) => {
      expect(renderKPI(size).card).toHaveAttribute("data-size", size);
    },
  );

  it("falls back to the medium size for an unknown value", () => {
    expect(renderKPI("huge").card).toHaveAttribute("data-size", "medium");
  });
});
