import type { VisualizationSpec } from "svelte-vega";
import { describe, expect, it } from "vitest";
import type { CartesianChartSpec } from "./CartesianChartProvider";
import {
  chartOrientation,
  dimensionChannel,
  isHorizontal,
  measureChannel,
  supportsOrientation,
  swapAxes,
  swapSortChannel,
  toVerticalSpec,
  transposeCartesianSpec,
} from "./orientation";
import { at } from "./test-fixtures";

const vertical: CartesianChartSpec = {
  metrics_view: "reddit_posts",
  color: "primary",
  x: {
    field: "post_title",
    type: "nominal",
    sort: "-y",
    limit: 10,
    axisOrient: "top",
  },
  y: {
    field: "post_count",
    type: "quantitative",
    zeroBasedOrigin: true,
    axisOrient: "right",
  },
};

const horizontal: CartesianChartSpec = {
  metrics_view: "reddit_posts",
  color: "primary",
  x: {
    field: "post_count",
    type: "quantitative",
    zeroBasedOrigin: true,
    axisOrient: "top",
  },
  y: {
    field: "post_title",
    type: "nominal",
    sort: "-x",
    limit: 10,
    axisOrient: "right",
  },
};

describe("supportsOrientation", () => {
  it("is limited to the bar chart types", () => {
    expect(supportsOrientation("bar_chart")).toBe(true);
    expect(supportsOrientation("stacked_bar")).toBe(true);
    expect(supportsOrientation("stacked_bar_normalized")).toBe(true);
    expect(supportsOrientation("line_chart")).toBe(false);
    expect(supportsOrientation("area_chart")).toBe(false);
    expect(supportsOrientation("donut_chart")).toBe(false);
  });
});

describe("orientation of a spec", () => {
  it("is horizontal when the measure sits on x", () => {
    expect(isHorizontal(vertical)).toBe(false);
    expect(isHorizontal(horizontal)).toBe(true);
    expect(isHorizontal({})).toBe(false);
    expect(chartOrientation(vertical)).toBe("vertical");
    expect(chartOrientation(horizontal)).toBe("horizontal");
  });

  it("names the channel of each role", () => {
    expect(dimensionChannel(vertical)).toBe("x");
    expect(measureChannel(vertical)).toBe("y");
    expect(dimensionChannel(horizontal)).toBe("y");
    expect(measureChannel(horizontal)).toBe("x");
  });
});

describe("swapSortChannel", () => {
  it("rewrites channel-relative sort values and leaves the rest alone", () => {
    expect(swapSortChannel("x")).toBe("y");
    expect(swapSortChannel("-x")).toBe("-y");
    expect(swapSortChannel("y")).toBe("x");
    expect(swapSortChannel("-y")).toBe("-x");
    expect(swapSortChannel("y_delta")).toBe("x_delta");
    expect(swapSortChannel("-x_delta")).toBe("-y_delta");
    expect(swapSortChannel("color")).toBe("color");
    expect(swapSortChannel("custom")).toBe("custom");
  });
});

describe("swapAxes", () => {
  it("moves each field to the other axis with its sort and axis placement", () => {
    expect(swapAxes(vertical)).toEqual(horizontal);
    expect(swapAxes(horizontal)).toEqual(vertical);
  });

  it("does not mutate its input", () => {
    swapAxes(vertical);
    expect(vertical.x?.field).toBe("post_title");
    expect(vertical.x?.sort).toBe("-y");
  });

  it("keeps custom sort arrays and missing fields", () => {
    const custom: CartesianChartSpec = {
      metrics_view: "reddit_posts",
      x: { field: "post_title", type: "nominal", sort: ["b", "a"] },
    };
    expect(swapAxes(custom)).toEqual({
      metrics_view: "reddit_posts",
      x: undefined,
      y: { field: "post_title", type: "nominal", sort: ["b", "a"] },
    });
  });

  it("leaves a hidden axis hidden", () => {
    const hidden: CartesianChartSpec = {
      ...vertical,
      x: { field: "post_title", type: "nominal", axisOrient: "none" },
    };
    expect(swapAxes(hidden).y?.axisOrient).toBe("none");
  });
});

describe("toVerticalSpec", () => {
  it("returns a vertical spec unchanged", () => {
    expect(toVerticalSpec(vertical)).toBe(vertical);
  });

  it("swaps a horizontal spec into the vertical layout", () => {
    expect(toVerticalSpec(horizontal)).toEqual(vertical);
  });
});

describe("transposeCartesianSpec", () => {
  const verticalVega = {
    $schema: "https://vega.github.io/schema/vega-lite/v5.json",
    width: "container",
    autosize: { type: "fit" },
    encoding: {
      x: { field: "category", type: "nominal", bandPosition: 0, sort: "-y" },
    },
    layer: [
      {
        mark: { type: "bar", clip: true, opacity: 0.8 },
        encoding: {
          x: { field: "category", sort: ["b", "a"] },
          color: { value: "transparent" },
        },
        params: [
          {
            name: "hover",
            select: { type: "point", encodings: ["x"], on: "pointerover" },
          },
        ],
      },
      {
        mark: { type: "bar", clip: true, width: { band: 0.9 } },
        encoding: {
          x: {
            field: "category",
            type: "nominal",
            axis: { orient: "top", labelAngle: -45 },
          },
          y: {
            field: "measure",
            type: "quantitative",
            scale: { zero: false, domainMax: 1.1 },
            axis: { orient: "right", format: ".0%" },
            stack: "normalize",
          },
          xOffset: { field: "period", sort: { field: "sortOrder" } },
          color: { field: "period" },
        },
      },
      {
        encoding: { y: { field: "measure", type: "quantitative" } },
        layer: [{ mark: { type: "line" } }],
      },
    ],
  } as unknown as VisualizationSpec;

  const horizontalVega = transposeCartesianSpec(verticalVega);

  it("does not mutate its input", () => {
    expect(at(verticalVega, "encoding.x.field")).toBe("category");
    expect(at(verticalVega, "layer.1.encoding.xOffset")).toBeDefined();
    expect(at(verticalVega, "layer.1.mark.width")).toEqual({ band: 0.9 });
  });

  it("moves the shared category encoding to the y channel", () => {
    expect(at(horizontalVega, "encoding.x")).toBeUndefined();
    expect(at(horizontalVega, "encoding.y")).toMatchObject({
      field: "category",
      type: "nominal",
      bandPosition: 0,
    });
  });

  it("rewrites channel-relative sort strings and keeps array sorts", () => {
    expect(at(horizontalVega, "encoding.y.sort")).toBe("-x");
    expect(at(horizontalVega, "layer.0.encoding.y.sort")).toEqual(["b", "a"]);
  });

  it("swaps x/y and xOffset/yOffset in every layer", () => {
    expect(at(horizontalVega, "layer.1.encoding.y.field")).toBe("category");
    expect(at(horizontalVega, "layer.1.encoding.x.field")).toBe("measure");
    expect(at(horizontalVega, "layer.1.encoding.xOffset")).toBeUndefined();
    expect(at(horizontalVega, "layer.1.encoding.yOffset")).toEqual({
      field: "period",
      sort: { field: "sortOrder" },
    });
    expect(at(horizontalVega, "layer.1.encoding.color")).toEqual({
      field: "period",
    });
  });

  it("carries measure axis settings over to the x channel", () => {
    expect(at(horizontalVega, "layer.1.encoding.x.stack")).toBe("normalize");
    expect(at(horizontalVega, "layer.1.encoding.x.scale")).toEqual({
      zero: false,
      domainMax: 1.1,
    });
    expect(at(horizontalVega, "layer.1.encoding.x.axis.format")).toBe(".0%");
  });

  it("rotates axis orientation and pins the grid to the measure axis", () => {
    expect(at(horizontalVega, "layer.1.encoding.y.axis")).toEqual({
      orient: "right",
      labelAngle: -45,
      grid: false,
    });
    expect(at(horizontalVega, "layer.1.encoding.x.axis")).toEqual({
      orient: "top",
      format: ".0%",
      grid: true,
    });
  });

  it("swaps relative band sizes on the mark but not the top-level width", () => {
    expect(at(horizontalVega, "layer.1.mark")).toEqual({
      type: "bar",
      clip: true,
      height: { band: 0.9 },
    });
    expect(at(horizontalVega, "layer.0.mark")).toEqual({
      type: "bar",
      clip: true,
      opacity: 0.8,
    });
    expect(at(horizontalVega, "width")).toBe("container");
    expect(at(horizontalVega, "height")).toBeUndefined();
  });

  it("points selection params at the y channel", () => {
    expect(at(horizontalVega, "layer.0.params.0.select.encodings")).toEqual([
      "y",
    ]);
    expect(at(horizontalVega, "layer.0.params.0.select.on")).toBe(
      "pointerover",
    );
  });

  it("recurses into nested layers", () => {
    expect(at(horizontalVega, "layer.2.encoding.x.field")).toBe("measure");
    expect(at(horizontalVega, "layer.2.encoding.y")).toBeUndefined();
    expect(at(horizontalVega, "layer.2.layer.0.mark")).toEqual({
      type: "line",
    });
  });

  it("is an involution", () => {
    const roundTrip = transposeCartesianSpec(horizontalVega);
    expect(at(roundTrip, "encoding.x")).toMatchObject({
      field: "category",
      sort: "-y",
    });
    expect(at(roundTrip, "layer.1.encoding.xOffset")).toEqual({
      field: "period",
      sort: { field: "sortOrder" },
    });
    expect(at(roundTrip, "layer.1.encoding.x.axis.orient")).toBe("top");
    expect(at(roundTrip, "layer.1.mark.width")).toEqual({ band: 0.9 });
    expect(at(roundTrip, "layer.0.params.0.select.encodings")).toEqual(["x"]);
  });
});
