import Chart from "@rilldata/web-common/features/components/charts/Chart.svelte";
import type { ChartDataResult } from "@rilldata/web-common/features/components/charts";
import { render } from "@testing-library/svelte";
import chroma from "chroma-js";
import { readable } from "svelte/store";
import { describe, expect, it } from "vitest";

const measure = { name: "operation_jx_cnt", displayName: "JX count" };

function chartData(partial: Partial<ChartDataResult>): ChartDataResult {
  return {
    data: [],
    isFetching: false,
    fields: {},
    isDarkMode: false,
    hasComparison: false,
    theme: {
      primary: chroma("#1d4ed8"),
      secondary: chroma("#7c3aed"),
    },
    ...partial,
  };
}

const chartSpec = {
  metrics_view: "ncar_jx_detail_metrics",
  x: { field: "cal_date", type: "temporal" as const },
  y: { field: "operation_jx_cnt", type: "quantitative" as const },
};

function renderChart(data: ChartDataResult, measures = [measure]) {
  return render(Chart, {
    props: {
      chartType: "line_chart",
      chartSpec,
      chartData: readable(data),
      measures,
      isCanvas: false,
      view: undefined,
    },
  });
}

describe("Chart loading and error states", () => {
  it("shows the query error instead of the spinner", () => {
    const { container, queryByText } = renderChart(
      chartData({
        isFetching: true,
        error: new Error("time range has too many bins"),
      }),
    );

    expect(queryByText("time range has too many bins")).toBeInTheDocument();
    expect(container.querySelector(".status")).toBeNull();
  });

  it("shows the spinner while fetching without an error", () => {
    const { container, queryByText } = renderChart(
      chartData({ isFetching: true }),
    );

    expect(container.querySelector(".status")).not.toBeNull();
    expect(queryByText("No Data to Display")).toBeNull();
  });

  it("shows an empty state when the query resolves with no rows", () => {
    const { container, queryByText } = renderChart(
      chartData({ isFetching: false, data: [] }),
    );

    expect(queryByText("No Data to Display")).toBeInTheDocument();
    expect(container.querySelector(".status")).toBeNull();
  });
});
