import type { V1CanvasSpec } from "@rilldata/web-common/runtime-client";
import { describe, expect, it } from "vitest";
import {
  collectControls,
  controlsForPath,
  filterValueViews,
  controlUrlKey,
  parseControlValue,
  serializeControlValue,
} from "./controls";

const spec: V1CanvasSpec = {
  controls: [{ kind: "measure", target: "spend", name: "metric" }],
  rows: [
    { items: [{ component: "top", name: "top" }] },
    {
      container: {
        name: "section",
        controls: [{ kind: "filter", target: "publisher", name: "publisher" }],
        rows: [
          {
            items: [
              {
                name: "own",
                controls: [
                  { kind: "filter", target: "publisher", name: "mine" },
                ],
              },
              { name: "plain" },
              {
                name: "ignore_all",
                ignoreAllControls: true,
                controls: [{ kind: "time_grain", name: "time_grain" }],
              },
              { name: "ignore_one", ignoreControls: ["metric"] },
            ],
          },
        ],
      },
    },
    {
      tabGroup: {
        name: "group",
        controls: [{ kind: "time_grain", name: "time_grain" }],
        tabs: [
          {
            name: "mobile",
            controls: [{ kind: "filter", target: "os", name: "os" }],
            rows: [{ items: [{ name: "in_tab" }] }],
          },
        ],
      },
    },
  ],
};

const keys = (path: (string | number)[]) =>
  controlsForPath(spec, path).map((c) => c.urlKey);

describe("controlsForPath", () => {
  it("resolves the nearest control per target, nearest level first", () => {
    expect(keys(["rows", 0, "items", 0, "kpi"])).toEqual(["c.metric"]);
    expect(keys(["rows", 1, "rows", 0, "items", 0, "bar_chart"])).toEqual([
      "c.own.mine",
      "c.metric",
    ]);
    expect(keys(["rows", 1, "rows", 0, "items", 1, "bar_chart"])).toEqual([
      "c.section.publisher",
      "c.metric",
    ]);
    expect(
      keys(["rows", 2, "tabs", 0, "rows", 0, "items", 0, "bar_chart"]),
    ).toEqual(["c.group.mobile.os", "c.group.time_grain", "c.metric"]);
  });

  it("applies ignore_controls to enclosing levels only", () => {
    expect(keys(["rows", 1, "rows", 0, "items", 2, "bar_chart"])).toEqual([
      "c.ignore_all.time_grain",
    ]);
    expect(keys(["rows", 1, "rows", 0, "items", 3, "bar_chart"])).toEqual([
      "c.section.publisher",
    ]);
  });

  it("returns nothing for a path that isn't on the canvas", () => {
    expect(keys(["rows", 9, "items", 0, "kpi"])).toEqual([]);
  });
});

describe("collectControls", () => {
  it("lists every control once", () => {
    expect(collectControls(spec).map((c) => c.urlKey)).toEqual([
      "c.metric",
      "c.section.publisher",
      "c.own.mine",
      "c.ignore_all.time_grain",
      "c.group.time_grain",
      "c.group.mobile.os",
    ]);
  });
});

describe("URL keys and values", () => {
  it("escapes dots and separators in key parts", () => {
    expect(controlUrlKey(["a.b", "c,d"], "x")).toBe("c.a%2Eb.c%2Cd.x");
  });

  const filter = { kind: "filter", target: "publisher", multiple: true };

  it("round-trips values, including separators", () => {
    const values = ["A,B", "C&D 100%"];
    const raw = serializeControlValue(filter, values);
    expect(raw).toBe("A%2CB,C%26D%20100%25");
    expect(parseControlValue(filter, raw!)).toEqual({ values, valid: true });
  });

  it("omits the default and keeps an explicit empty value", () => {
    const withDefault = { ...filter, default: ["Google"] };
    expect(serializeControlValue(withDefault, ["Google"])).toBeUndefined();
    expect(serializeControlValue(withDefault, [])).toBe("");
    expect(parseControlValue(withDefault, null).values).toEqual(["Google"]);
    expect(parseControlValue(withDefault, "").values).toEqual([]);
  });

  it("falls back to the default for a value that isn't allowed", () => {
    const metric = {
      kind: "measure",
      target: "spend",
      options: ["spend", "impressions"],
    };
    expect(parseControlValue(metric, "impressions")).toEqual({
      values: ["impressions"],
      valid: true,
    });
    expect(parseControlValue(metric, "nope")).toEqual({
      values: ["spend"],
      valid: false,
    });
    expect(parseControlValue(metric, "spend,impressions").valid).toBe(false);
    expect(parseControlValue(metric, "").valid).toBe(false);
    expect(serializeControlValue(metric, ["spend"])).toBeUndefined();
  });
});

describe("filterValueViews", () => {
  const dims: Record<string, string[]> = {
    bids: ["publisher", "country"],
    sales: ["billing_country"],
    visits: ["geo"],
  };
  const views = Object.keys(dims);
  const hasDimension = (view: string, dimension: string) =>
    dims[view].includes(dimension);

  it("reads values from the views that have the target dimension", () => {
    expect(
      filterValueViews({ target: "publisher" }, views, hasDimension),
    ).toEqual(["bids"]);
    expect(
      filterValueViews(
        { target: "country", fields: { sales: "billing_country" } },
        views,
        hasDimension,
      ),
    ).toEqual(["bids"]);
  });

  it("falls back to mapped views when no view has the target itself", () => {
    expect(
      filterValueViews(
        {
          target: "region",
          fields: { sales: "billing_country", visits: "geo" },
        },
        views,
        hasDimension,
      ),
    ).toEqual(["sales"]);
  });

  it("returns nothing when no view has the dimension", () => {
    expect(
      filterValueViews(
        { target: "region", fields: { sales: "missing" } },
        views,
        hasDimension,
      ),
    ).toEqual([]);
  });
});
