import type { V1MetricsViewSpec } from "@rilldata/web-common/runtime-client";
import { describe, it, expect } from "vitest";
import { parseDocument } from "yaml";
import {
  forEachRowsList,
  generateNewAssets,
  getItemPropertiesExcept,
  isHiddenPath,
  generateArrayRearrangeFunction,
  mapGuard,
  positionKeyPrefix,
  rowsGuard,
  specRowsAt,
  targetOfComponent,
} from "./layout-util";

// A canvas with one free row followed by a tab group that contains a widget.
const CANVAS = `type: canvas
rows:
  - items:
      - component: header
  - name: deep_dive
    tabs:
      - label: Overview
        rows:
          - items:
              - component: d1--component-g1-t0-0-0
`;

describe("mapGuard", () => {
  it("keeps the `if` condition of plain rows", () => {
    const doc = parseDocument(`type: canvas
rows:
  - if: '{{ has "finance" .user.groups }}'
    height: 320px
    items:
      - component: margin
`);
    const rows = mapGuard(rowsGuard(doc.getIn(["rows"])));

    expect(rows[0]).toEqual({
      if: '{{ has "finance" .user.groups }}',
      height: "320px",
      items: [{ component: "margin" }],
    });
  });

  it("preserves tab group rows instead of coercing them to empty items", () => {
    const doc = parseDocument(CANVAS);
    const rows = mapGuard(rowsGuard(doc.getIn(["rows"])));

    expect(rows).toHaveLength(2);
    expect(rows[0].items).toEqual([{ component: "header" }]);
    // The tab group row keeps its tabs and has no coerced `items`.
    expect(rows[1].items).toBeUndefined();
    expect(rows[1].name).toBe("deep_dive");
    expect(rows[1].tabs).toBeDefined();
  });
});

describe("mapGuard with containers", () => {
  it("preserves containers and unknown blocks, but cleans up empty plain rows", () => {
    const doc = parseDocument(`type: canvas
rows:
  - name: story
    if: '{{ .user.admin }}'
    rows:
      - items:
          - component: a
  - future_block:
      key: value
  - height: 40px
`);
    const rows = mapGuard(rowsGuard(doc.getIn(["rows"])));

    expect(rows[0]).toEqual({
      name: "story",
      if: "{{ .user.admin }}",
      rows: [{ items: [{ component: "a" }] }],
    });
    // A kind of block this editor doesn't know is passed through verbatim too.
    expect(rows[1]).toEqual({ future_block: { key: "value" } });
    // A plain row without items is still coerced, so the empty-row cleanup removes it.
    expect(rows[2]).toEqual({ height: "40px", items: [] });
  });

  it("keeps blocks written with an empty items list next to their content", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items: []
    rows:
      - items:
          - component: a
  - items: []
    tabs:
      - label: A
        rows: []
  - items:
      - component: b
`);
    const yamlRows = mapGuard(rowsGuard(doc.getIn(["rows"])));
    // The empty items are dropped, so the empty-row cleanup leaves the blocks alone.
    expect(yamlRows[0]).toEqual({ rows: [{ items: [{ component: "a" }] }] });
    expect(yamlRows[1]).toEqual({ tabs: [{ label: "A", rows: [] }] });

    const { newYamlRows } = generateNewAssets({
      transaction: {
        operations: [
          {
            type: "add",
            insertRow: true,
            componentType: "markdown",
            destination: { row: 3, col: 0 },
          },
        ],
      },
      yamlRows,
      specRows: [
        {
          items: [],
          container: {
            name: "container-0",
            rows: [{ items: [{ component: "a" }] }],
          },
        },
        { items: [], tabGroup: { name: "group-1", tabs: [] } },
        { items: [{ component: "b", definedInCanvas: false }] },
      ],
      resolvedComponents: {},
      canvasName: "d1",
      defaultMetrics: { metricsViewName: "foo", metricsViewSpec: undefined },
    });
    expect(newYamlRows).toHaveLength(4);
    expect(newYamlRows[0].rows).toEqual([{ items: [{ component: "a" }] }]);
    expect(newYamlRows[1].tabs).toEqual([{ label: "A", rows: [] }]);
  });

  it("gives a copied item a name of its own, since names key viewer controls", () => {
    const rearrange = generateArrayRearrangeFunction({
      operations: [
        {
          type: "copy",
          insertRow: true,
          source: { row: 0, col: 0 },
          destination: { row: 1, col: 0 },
        },
      ],
    });
    const rows = [
      {
        items: [
          { name: "chart", controls: [{ filter: "publisher" }] },
          { name: "chart_copy" },
        ],
      },
    ];
    const result = rearrange(
      rows,
      () => ({ name: "" }),
      (row) => row,
    );
    const names = result.flatMap((row) => row.items.map((item) => item.name));
    expect(names).toEqual(["chart", "chart_copy", "chart_copy2"]);
    expect(result[1].items[0]).toMatchObject({
      controls: [{ filter: "publisher" }],
    });
  });

  it("names a copied item clear of names elsewhere on the canvas", () => {
    // "chart_copy" is taken by an item in a container, outside the rows being edited.
    const rearrange = generateArrayRearrangeFunction(
      {
        operations: [
          {
            type: "copy",
            insertRow: true,
            source: { row: 0, col: 0 },
            destination: { row: 1, col: 0 },
          },
        ],
      },
      new Set(["chart", "chart_copy"]),
    );
    const rows = [{ items: [{ name: "chart" }] }];
    const apply = () =>
      rearrange(
        rows,
        () => ({ name: "" }),
        (row) => row,
      ).flatMap((row) => row.items.map((item) => item.name));

    expect(apply()).toEqual(["chart", "chart_copy2"]);
    // The transaction is applied to the YAML rows and the spec rows: same name both times.
    expect(apply()).toEqual(["chart", "chart_copy2"]);
  });

  it("keeps a container when a top-level row is added before it", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: header
  - rows:
      - items:
          - component: d1--component-c1-0-0
`);
    const { newYamlRows, newSpecRows } = generateNewAssets({
      transaction: {
        operations: [
          {
            type: "add",
            insertRow: true,
            componentType: "markdown",
            destination: { row: 1, col: 0 },
          },
        ],
      },
      yamlRows: mapGuard(rowsGuard(doc.getIn(["rows"]))),
      specRows: [
        { items: [{ component: "header" }] },
        {
          container: {
            name: "container-1",
            rows: [{ items: [{ component: "d1--component-c1-0-0" }] }],
          },
        },
      ],
      resolvedComponents: {},
      canvasName: "d1",
      defaultMetrics: { metricsViewName: "foo", metricsViewSpec: undefined },
    });

    expect(newYamlRows).toHaveLength(3);
    expect(newYamlRows[2].rows).toEqual([
      { items: [{ component: "d1--component-c1-0-0" }] },
    ]);
    expect(newSpecRows[2].container?.rows?.[0]?.items?.[0]?.component).toBe(
      "d1--component-c1-0-0",
    );
  });
});

describe("generateNewAssets with blocks as the spec JSON has them", () => {
  it("keeps tab groups and containers, which come with empty items, aligned through a row edit", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: a
  - tabs:
      - label: Overview
        rows: []
  - rows:
      - items:
          - component: d1--component-c2-0-0
  - items:
      - component: b
`);
    const { newYamlRows, newSpecRows } = generateNewAssets({
      // Delete the first row's only widget, which removes the row.
      transaction: {
        operations: [{ type: "delete", target: { row: 0, col: 0 } }],
      },
      yamlRows: mapGuard(rowsGuard(doc.getIn(["rows"]))),
      specRows: [
        { items: [{ component: "a", definedInCanvas: false }] },
        {
          items: [],
          tabGroup: { name: "group-1", tabs: [{ name: "overview" }] },
        },
        {
          items: [],
          container: {
            name: "container-2",
            rows: [{ items: [{ component: "d1--component-c2-0-0" }] }],
          },
        },
        { items: [{ component: "b", definedInCanvas: false }] },
      ],
      resolvedComponents: {},
      canvasName: "d1",
      defaultMetrics: { metricsViewName: "foo", metricsViewSpec: undefined },
    });

    expect(newYamlRows.map((row) => Object.keys(row)[0])).toEqual([
      "tabs",
      "rows",
      "items",
    ]);
    expect(newSpecRows.map((row) => !!row.tabGroup)).toEqual([
      true,
      false,
      false,
    ]);
    expect(newSpecRows[1].container?.name).toBe("container-2");
    expect(newSpecRows[2].items?.[0]?.component).toBe("b");
  });
});

describe("rows paths", () => {
  const spec = [
    { items: [{ component: "a" }] },
    {
      tabGroup: {
        name: "g",
        tabs: [
          { name: "t0", rows: [{ items: [{ component: "b" }] }] },
          { name: "t1", rows: [{ items: [{ component: "c" }] }] },
        ],
      },
    },
    { container: { name: "story", rows: [{ items: [{ component: "d" }] }] } },
    { container: { name: "empty" } },
  ];

  it("derives the parser's position key prefix from a rows path", () => {
    expect(positionKeyPrefix(["rows"])).toBe("");
    expect(positionKeyPrefix(["rows", 2, "tabs", 0, "rows"])).toBe("g2-t0-");
    expect(positionKeyPrefix(["rows", 1, "rows"])).toBe("c1-");
  });

  it("finds the spec rows at a YAML rows path", () => {
    expect(specRowsAt(spec, ["rows"])).toBe(spec);
    expect(specRowsAt(spec, ["rows", 1, "tabs", 1, "rows"])).toBe(
      spec[1].tabGroup?.tabs?.[1].rows,
    );
    expect(specRowsAt(spec, ["rows", 2, "rows"])).toBe(spec[2].container?.rows);
    expect(specRowsAt(spec, ["rows", 0, "rows"])).toBeUndefined();
    expect(specRowsAt(spec, ["rows", 5, "tabs", 0, "rows"])).toBeUndefined();
  });

  it("creates the missing rows list of an empty block on request", () => {
    const rows = structuredClone(spec);
    expect(specRowsAt(rows, ["rows", 3, "rows"])).toBeUndefined();
    const created = specRowsAt(rows, ["rows", 3, "rows"], true);
    expect(created).toEqual([]);
    expect(rows[3].container?.rows).toBe(created);
  });

  it("visits every rows list with its YAML path", () => {
    const visited: string[] = [];
    forEachRowsList(spec, (rows, rowsPath) =>
      visited.push(`${rowsPath.join(".")}:${rows.length}`),
    );
    expect(visited).toEqual([
      "rows:4",
      "rows.1.tabs.0.rows:1",
      "rows.1.tabs.1.rows:1",
      "rows.2.rows:1",
      "rows.3.rows:0",
    ]);
  });

  it("finds the edit target a component path lives in", () => {
    expect(
      targetOfComponent(["rows", 0, "items", 1, "markdown"]),
    ).toBeUndefined();
    expect(
      targetOfComponent(["rows", 2, "rows", 0, "items", 1, "markdown"]),
    ).toEqual({ rowsPath: ["rows", 2, "rows"] });
    expect(
      targetOfComponent([
        "rows",
        1,
        "tabs",
        0,
        "rows",
        0,
        "items",
        0,
        "markdown",
      ]),
    ).toEqual({ rowsPath: ["rows", 1, "tabs", 0, "rows"] });
  });
});

describe("generateNewAssets with a tab group present", () => {
  it("preserves the tab group when a top-level row is added before it", () => {
    const doc = parseDocument(CANVAS);
    const yamlRows = mapGuard(rowsGuard(doc.getIn(["rows"])));
    const specRows = [
      { items: [{ component: "header" }] },
      {
        tabGroup: {
          name: "deep_dive",
          tabs: [
            {
              name: "overview",
              displayName: "Overview",
              rows: [{ items: [{ component: "d1--component-g1-t0-0-0" }] }],
            },
          ],
        },
      },
    ];

    const { newYamlRows, newSpecRows } = generateNewAssets({
      transaction: {
        operations: [
          {
            type: "add",
            insertRow: true,
            componentType: "markdown",
            destination: { row: 1, col: 0 },
          },
        ],
      },
      yamlRows,
      specRows,
      resolvedComponents: {},
      canvasName: "d1",
      defaultMetrics: { metricsViewName: "foo", metricsViewSpec: undefined },
    });

    // The new row was inserted and the tab group survived (it was previously stripped
    // to empty items and deleted by the cleanup step).
    const tabRow = newYamlRows.find((r) => r.tabs !== undefined);
    expect(tabRow).toBeDefined();
    expect(newYamlRows).toHaveLength(3);

    const specTabRow = newSpecRows.find((r) => r.tabGroup);
    expect(
      specTabRow?.tabGroup?.tabs?.[0]?.rows?.[0]?.items?.[0]?.component,
    ).toBe("d1--component-g1-t0-0-0");
  });
});

describe("generateNewAssets moving items between rows", () => {
  const doc = parseDocument(`type: canvas
rows:
  - if: '{{ has "finance" .user.groups }}'
    items:
      - component: margin
      - component: revenue
        if: '{{ .user.admin }}'
  - items:
      - component: header
`);
  const yamlRows = mapGuard(rowsGuard(doc.getIn(["rows"])));
  const specRows = [
    { items: [{ component: "margin" }, { component: "revenue" }] },
    { items: [{ component: "header" }] },
  ];
  const move = (col: number, insertRow: boolean) =>
    generateNewAssets({
      transaction: {
        operations: [
          {
            type: "move",
            insertRow,
            source: { row: 0, col },
            destination: { row: insertRow ? 2 : 1, col: 0 },
          },
        ],
      },
      yamlRows,
      specRows,
      resolvedComponents: {},
      canvasName: "d1",
      defaultMetrics: { metricsViewName: "foo", metricsViewSpec: undefined },
    }).newYamlRows;

  it("keeps the source row's condition on an item moved to another row", () => {
    const rows = move(0, false);
    expect(rows[1].items?.[0]).toMatchObject({
      component: "margin",
      if: '{{ has "finance" .user.groups }}',
    });
  });

  it("combines the source row's condition with the item's own condition in a new row", () => {
    const rows = move(1, true);
    expect(rows[2].items?.[0]).toMatchObject({
      component: "revenue",
      if: '({{ has "finance" .user.groups }}) AND ({{ .user.admin }})',
    });
  });
});

describe("getItemPropertiesExcept", () => {
  it("returns the item's properties other than its component definition", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - width: 6
        if: '{{ .user.admin }}'
        bar_chart:
          metrics_view: orders
`);
    expect(
      getItemPropertiesExcept(doc, ["rows", 0, "items", 0], "bar_chart"),
    ).toEqual({ width: 6, if: "{{ .user.admin }}" });
  });
});

describe("isHiddenPath", () => {
  const hiddenPaths = new Set([
    "rows.1",
    "rows.3.tabs.0.rows.0.items.1",
    "rows.4.rows.1",
  ]);

  it("hides the rows of a container that is hidden or whose row is hidden", () => {
    expect(isHiddenPath(hiddenPaths, ["rows", 4, "rows", 1, "items", 0])).toBe(
      true,
    );
    expect(isHiddenPath(hiddenPaths, ["rows", 4, "rows", 0, "items", 0])).toBe(
      false,
    );
  });

  it("hides an element whose own path or an ancestor's path is hidden", () => {
    expect(isHiddenPath(hiddenPaths, ["rows", 1, "items", 0])).toBe(true);
    expect(
      isHiddenPath(hiddenPaths, ["rows", 3, "tabs", 0, "rows", 0, "items", 1]),
    ).toBe(true);
  });

  it("shows other elements", () => {
    expect(isHiddenPath(hiddenPaths, ["rows", 0, "items", 0])).toBe(false);
    expect(isHiddenPath(hiddenPaths, ["rows", 10, "items", 0])).toBe(false);
    expect(
      isHiddenPath(hiddenPaths, ["rows", 3, "tabs", 0, "rows", 0, "items", 0]),
    ).toBe(false);
  });
});

describe("generateNewAssets with component references", () => {
  const REF_CANVAS = `type: canvas
rows:
  - items:
      - component: trend
        params:
          metrics_view: mv1
          measure: total
`;

  const trendResource = {
    meta: { name: { name: "trend", kind: "rill.runtime.v1.Component" } },
    component: {
      state: {
        validSpec: {
          renderer: "custom_chart",
          rendererProperties: {
            metrics_sql:
              "SELECT {{ .params.measure }} FROM {{ .params.metrics_view }}",
            spec: {
              chartType: "Bar Chart",
              encodings: { y: { field: "{{ .params.measure }}" } },
            },
          },
          params: [
            { name: "metrics_view", type: "metrics_view", required: true },
            {
              name: "measure",
              type: "measure",
              required: true,
              metricsViewParam: "metrics_view",
            },
          ],
        },
      },
    },
  };

  const metricsViewSpec: V1MetricsViewSpec = {
    timeDimension: "ts",
    dimensions: [{ name: "country", type: "DIMENSION_TYPE_CATEGORICAL" }],
    measures: [{ name: "total" }],
  };

  it("adds a reference item with default param bindings", () => {
    const doc = parseDocument(REF_CANVAS);
    const yamlRows = mapGuard(rowsGuard(doc.getIn(["rows"])));
    const specRows = [
      {
        items: [
          {
            component: "trend",
            definedInCanvas: false,
            params: { metrics_view: "mv1", measure: "total" },
          },
        ],
      },
    ];

    const { newYamlRows, newSpecRows, newResolvedComponents } =
      generateNewAssets({
        transaction: {
          operations: [
            {
              type: "add",
              insertRow: true,
              componentType: "component_ref",
              componentName: "trend",
              destination: { row: 1, col: 0 },
            },
          ],
        },
        yamlRows,
        specRows,
        resolvedComponents: { trend: trendResource },
        canvasName: "d1",
        defaultMetrics: { metricsViewName: "mv1", metricsViewSpec },
        componentResources: { trend: trendResource },
      });

    // The new YAML item references the component with default bindings.
    expect(newYamlRows).toHaveLength(2);
    expect(newYamlRows[1].items?.[0]).toEqual({
      component: "trend",
      params: { metrics_view: "mv1", measure: "total" },
      // The touched row's widths are normalized across its (single) item.
      width: 12,
    });

    // Both spec items keep the real component name (no positional rename).
    expect(newSpecRows[0].items?.[0]?.component).toBe("trend");
    expect(newSpecRows[1].items?.[0]?.component).toBe("trend");
    expect(newSpecRows[1].items?.[0]?.definedInCanvas).toBe(false);
    expect(newSpecRows[1].items?.[0]?.params).toEqual({
      metrics_view: "mv1",
      measure: "total",
    });

    // The optimistic resolved components map is keyed by the real resource name.
    expect(newResolvedComponents["trend"]?.meta?.name?.name).toBe("trend");
  });

  it("keeps reference names intact when inline items are repositioned", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: trend
        params:
          metrics_view: mv1
      - markdown:
          content: Hi
`);
    const yamlRows = mapGuard(rowsGuard(doc.getIn(["rows"])));
    const specRows = [
      {
        items: [
          {
            component: "trend",
            definedInCanvas: false,
            params: { metrics_view: "mv1" },
          },
          { component: "d1--component-0-1", definedInCanvas: true },
        ],
      },
    ];

    const { newSpecRows } = generateNewAssets({
      transaction: {
        operations: [
          {
            type: "move",
            source: { row: 0, col: 1 },
            destination: { row: 0, col: 0 },
          },
        ],
      },
      yamlRows,
      specRows,
      resolvedComponents: {
        trend: trendResource,
        "d1--component-0-1": {
          meta: { name: { name: "d1--component-0-1" } },
          component: {
            state: {
              validSpec: {
                renderer: "markdown",
                rendererProperties: { content: "Hi" },
              },
            },
          },
        },
      },
      canvasName: "d1",
      defaultMetrics: { metricsViewName: "mv1", metricsViewSpec },
    });

    // The inline item was renamed to its new position; the reference kept its name.
    expect(newSpecRows[0].items?.[0]?.component).toBe("d1--component-0-0");
    expect(newSpecRows[0].items?.[1]?.component).toBe("trend");
    expect(newSpecRows[0].items?.[1]?.params).toEqual({ metrics_view: "mv1" });
  });
});
