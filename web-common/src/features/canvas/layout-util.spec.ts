import { describe, it, expect } from "vitest";
import { parseDocument } from "yaml";
import { generateNewAssets, mapGuard, rowsGuard } from "./layout-util";

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
