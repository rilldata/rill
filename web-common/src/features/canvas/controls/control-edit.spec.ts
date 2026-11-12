import { describe, expect, it } from "vitest";
import { parseDocument } from "yaml";
import {
  addControl,
  controlLevelsForItem,
  controlToYAML,
  countItemsUsingField,
  draftFromYAML,
  moveControl,
  moveControlToLevel,
  readControls,
  removeControl,
  updateControl,
  type ControlDraft,
} from "./control-edit";

const YAML = `type: canvas
rows:
  - items:
      - width: 6
        bar_chart:
          metrics_view: bids
          x: { field: publisher }
          y: { field: total_bids }
      - width: 6
        name: bar_chart
        ignore_controls: true
        bar_chart:
          metrics_view: bids
          x: { field: publisher }
          y: { field: total_bids }
  - name: deep_dive
    rows:
      - items:
          - width: 12
            leaderboard:
              metrics_view: bids
              measures: [total_bids]
              dimensions: [domain]
  - name: views
    tabs:
      - label: Records
        rows:
          - items:
              - width: 12
                kpi_grid:
                  metrics_view: bids
                  measures: [total_bids, spend]
`;

function draft(overrides: Partial<ControlDraft>): ControlDraft {
  return {
    kind: "measure",
    field: "total_bids",
    options: [],
    default: [],
    multiple: false,
    display: "",
    label: "",
    name: "",
    ...overrides,
  };
}

describe("controlToYAML", () => {
  it("writes a swap with its field among the options and no default", () => {
    expect(
      controlToYAML(
        draft({ options: ["spend"], default: ["spend"], label: "Metric" }),
      ),
    ).toEqual({
      measure: "total_bids",
      label: "Metric",
      options: ["total_bids", "spend"],
    });
  });

  it("writes a filter's default as a list only when it takes several values", () => {
    const filter = draft({
      kind: "filter",
      field: "publisher",
      default: ["Acme", "Globex"],
    });
    expect(controlToYAML(filter)).toEqual({
      filter: "publisher",
      default: "Acme",
    });
    expect(controlToYAML({ ...filter, multiple: true })).toEqual({
      filter: "publisher",
      default: ["Acme", "Globex"],
      multiple: true,
    });
  });

  it("writes time controls and drops a default that isn't an option", () => {
    expect(
      controlToYAML(
        draft({
          kind: "time_range",
          options: ["P7D", "P4W"],
          default: ["P3M"],
          display: "buttons",
        }),
      ),
    ).toEqual({ time: "range", options: ["P7D", "P4W"], display: "buttons" });
    expect(
      controlToYAML(
        draft({
          kind: "time_grain",
          options: ["day", "week"],
          default: ["week"],
        }),
      ),
    ).toEqual({ time: "grain", options: ["day", "week"], default: "week" });
  });

  it("keeps keys the form doesn't own", () => {
    expect(
      controlToYAML(draft({ kind: "filter", field: "publisher" }), {
        filter: "domain",
        label: "Old",
        fields: { other_bids: "pub" },
      }),
    ).toEqual({ filter: "publisher", fields: { other_bids: "pub" } });
  });
});

describe("draftFromYAML", () => {
  it("round-trips a control", () => {
    const yaml = {
      filter: "publisher",
      name: "pub",
      label: "Publisher",
      options: ["Acme", "Globex"],
      default: ["Acme"],
      multiple: true,
      display: "dropdown",
    };
    const parsed = draftFromYAML(yaml);
    expect(parsed).not.toBeNull();
    expect(controlToYAML(parsed!)).toEqual(yaml);
  });

  it("reads a scalar default and can't edit a slot control", () => {
    expect(draftFromYAML({ time: "grain", default: "day" })?.default).toEqual([
      "day",
    ]);
    expect(draftFromYAML({ slot: "color.field", options: ["a"] })).toBeNull();
  });
});

describe("editing controls in the document", () => {
  it("adds a control to an item before its component, and names the item", () => {
    const doc = parseDocument(YAML);
    const index = addControl(
      doc,
      ["rows", 0, "items", 0],
      draft({ options: ["spend"] }),
    );
    expect(index).toBe(0);
    const item = doc.getIn(["rows", 0, "items", 0]) as {
      toJSON: () => Record<string, unknown>;
    };
    // "bar_chart" is taken by the item next to it.
    expect(Object.keys(item.toJSON())).toEqual([
      "width",
      "name",
      "controls",
      "bar_chart",
    ]);
    expect(item.toJSON().name).toBe("bar_chart_2");
    expect(item.toJSON().controls).toEqual([
      { measure: "total_bids", options: ["total_bids", "spend"] },
    ]);
  });

  it("keeps the name of an item that has one", () => {
    const doc = parseDocument(YAML);
    addControl(
      doc,
      ["rows", 0, "items", 1],
      draft({ kind: "filter", field: "publisher" }),
    );
    expect(doc.getIn(["rows", 0, "items", 1, "name"])).toBe("bar_chart");
  });

  it("names a control whose default name is taken at its level", () => {
    const doc = parseDocument(YAML);
    const path = ["rows", 1];
    addControl(
      doc,
      path,
      draft({ kind: "dimension", field: "domain", options: ["publisher"] }),
    );
    addControl(doc, path, draft({ kind: "filter", field: "domain" }));
    addControl(doc, path, draft({ kind: "filter", field: "domain" }));
    expect(readControls(doc, path)).toEqual([
      { dimension: "domain", options: ["domain", "publisher"] },
      { filter: "domain", name: "domain_filter" },
      { filter: "domain", name: "domain_2" },
    ]);
  });

  it("keeps a canvas control's name clear of item and block names", () => {
    const doc = parseDocument(YAML);
    addControl(doc, [], draft({ kind: "filter", field: "views" }));
    expect(readControls(doc, [])).toEqual([
      { filter: "views", name: "views_filter" },
    ]);
    // The canvas's controls go before its rows.
    const keys = Object.keys(doc.toJSON() as Record<string, unknown>);
    expect(keys.indexOf("controls")).toBeLessThan(keys.indexOf("rows"));
  });

  it("updates, reorders and removes controls", () => {
    const doc = parseDocument(YAML);
    const path = ["rows", 2, "tabs", 0];
    addControl(doc, path, draft({ options: ["spend"] }));
    addControl(doc, path, draft({ kind: "filter", field: "publisher" }));

    updateControl(
      doc,
      path,
      1,
      draft({ kind: "filter", field: "publisher", label: "Seller" }),
    );
    expect(readControls(doc, path)[1]).toEqual({
      filter: "publisher",
      label: "Seller",
    });

    moveControl(doc, path, 1, -1);
    expect(readControls(doc, path).map((c) => Object.keys(c)[0])).toEqual([
      "filter",
      "measure",
    ]);
    // Moving past the ends does nothing.
    moveControl(doc, path, 0, -1);
    expect(readControls(doc, path)).toHaveLength(2);

    removeControl(doc, path, 0);
    expect(readControls(doc, path)).toHaveLength(1);
    removeControl(doc, path, 0);
    expect(doc.hasIn([...path, "controls"])).toBe(false);
  });

  it("moves a control from an item to an enclosing level", () => {
    const doc = parseDocument(YAML);
    const item = ["rows", 1, "rows", 0, "items", 0];
    addControl(doc, item, draft({ options: ["spend"], label: "Metric" }));
    const control = draftFromYAML(readControls(doc, item)[0])!;

    moveControlToLevel(doc, item, 0, ["rows", 1], control);
    expect(doc.hasIn([...item, "controls"])).toBe(false);
    expect(readControls(doc, ["rows", 1])).toEqual([
      {
        measure: "total_bids",
        label: "Metric",
        options: ["total_bids", "spend"],
      },
    ]);
  });
});

describe("controlLevelsForItem", () => {
  it("lists an item's levels nearest first", () => {
    const doc = parseDocument(YAML);
    expect(controlLevelsForItem(doc, ["rows", 0, "items", 1])).toEqual([
      { kind: "item", path: ["rows", 0, "items", 1], name: "bar_chart" },
      { kind: "canvas", path: [], name: "" },
    ]);
    expect(
      controlLevelsForItem(doc, ["rows", 1, "rows", 0, "items", 0]),
    ).toEqual([
      { kind: "item", path: ["rows", 1, "rows", 0, "items", 0], name: "" },
      { kind: "container", path: ["rows", 1], name: "deep_dive" },
      { kind: "canvas", path: [], name: "" },
    ]);
    expect(
      controlLevelsForItem(doc, ["rows", 2, "tabs", 0, "rows", 0, "items", 0]),
    ).toEqual([
      {
        kind: "item",
        path: ["rows", 2, "tabs", 0, "rows", 0, "items", 0],
        name: "",
      },
      { kind: "tab", path: ["rows", 2, "tabs", 0], name: "Records" },
      { kind: "group", path: ["rows", 2], name: "views" },
      { kind: "canvas", path: [], name: "" },
    ]);
  });
});

describe("countItemsUsingField", () => {
  const doc = parseDocument(YAML);

  it("counts the items a swap at each level would change", () => {
    // The second chart ignores controls of enclosing levels.
    expect(countItemsUsingField(doc, [], "total_bids")).toBe(3);
    expect(countItemsUsingField(doc, [], "publisher")).toBe(1);
    expect(countItemsUsingField(doc, ["rows", 1], "domain")).toBe(1);
    expect(countItemsUsingField(doc, ["rows", 1], "publisher")).toBe(0);
    expect(countItemsUsingField(doc, ["rows", 2], "spend")).toBe(1);
    expect(countItemsUsingField(doc, ["rows", 2, "tabs", 0], "spend")).toBe(1);
  });

  it("counts an item for its own control even if it ignores outer ones", () => {
    expect(
      countItemsUsingField(doc, ["rows", 0, "items", 1], "total_bids"),
    ).toBe(1);
    expect(countItemsUsingField(doc, ["rows", 0, "items", 1], "spend")).toBe(0);
  });
});
