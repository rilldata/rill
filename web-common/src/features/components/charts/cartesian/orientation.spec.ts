import type { VisualizationSpec } from "svelte-vega";
import { describe, expect, it } from "vitest";
import { supportsOrientation, transposeCartesianSpec } from "./orientation";
import { at } from "./test-fixtures";

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

describe("transposeCartesianSpec", () => {
  const vertical = {
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

  const horizontal = transposeCartesianSpec(vertical);

  it("does not mutate its input", () => {
    expect(at(vertical, "encoding.x.field")).toBe("category");
    expect(at(vertical, "layer.1.encoding.xOffset")).toBeDefined();
    expect(at(vertical, "layer.1.mark.width")).toEqual({ band: 0.9 });
  });

  it("moves the shared category encoding to the y channel", () => {
    expect(at(horizontal, "encoding.x")).toBeUndefined();
    expect(at(horizontal, "encoding.y")).toMatchObject({
      field: "category",
      type: "nominal",
      bandPosition: 0,
    });
  });

  it("rewrites channel-relative sort strings and keeps array sorts", () => {
    expect(at(horizontal, "encoding.y.sort")).toBe("-x");
    expect(at(horizontal, "layer.0.encoding.y.sort")).toEqual(["b", "a"]);
  });

  it("swaps x/y and xOffset/yOffset in every layer", () => {
    expect(at(horizontal, "layer.1.encoding.y.field")).toBe("category");
    expect(at(horizontal, "layer.1.encoding.x.field")).toBe("measure");
    expect(at(horizontal, "layer.1.encoding.xOffset")).toBeUndefined();
    expect(at(horizontal, "layer.1.encoding.yOffset")).toEqual({
      field: "period",
      sort: { field: "sortOrder" },
    });
    expect(at(horizontal, "layer.1.encoding.color")).toEqual({
      field: "period",
    });
  });

  it("carries measure axis settings over to the x channel", () => {
    expect(at(horizontal, "layer.1.encoding.x.stack")).toBe("normalize");
    expect(at(horizontal, "layer.1.encoding.x.scale")).toEqual({
      zero: false,
      domainMax: 1.1,
    });
    expect(at(horizontal, "layer.1.encoding.x.axis.format")).toBe(".0%");
  });

  it("rotates axis orientation and pins the grid to the measure axis", () => {
    expect(at(horizontal, "layer.1.encoding.y.axis")).toEqual({
      orient: "right",
      labelAngle: -45,
      grid: false,
    });
    expect(at(horizontal, "layer.1.encoding.x.axis")).toEqual({
      orient: "top",
      format: ".0%",
      grid: true,
    });
  });

  it("swaps relative band sizes on the mark but not the top-level width", () => {
    expect(at(horizontal, "layer.1.mark")).toEqual({
      type: "bar",
      clip: true,
      height: { band: 0.9 },
    });
    expect(at(horizontal, "layer.0.mark")).toEqual({
      type: "bar",
      clip: true,
      opacity: 0.8,
    });
    expect(at(horizontal, "width")).toBe("container");
    expect(at(horizontal, "height")).toBeUndefined();
  });

  it("points selection params at the y channel", () => {
    expect(at(horizontal, "layer.0.params.0.select.encodings")).toEqual(["y"]);
    expect(at(horizontal, "layer.0.params.0.select.on")).toBe("pointerover");
  });

  it("recurses into nested layers", () => {
    expect(at(horizontal, "layer.2.encoding.x.field")).toBe("measure");
    expect(at(horizontal, "layer.2.encoding.y")).toBeUndefined();
    expect(at(horizontal, "layer.2.layer.0.mark")).toEqual({ type: "line" });
  });

  it("is an involution", () => {
    const roundTrip = transposeCartesianSpec(horizontal);
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
