import { describe, it, expect } from "vitest";
import { parseDocument } from "yaml";
import {
  addTab,
  addTabGroup,
  addTabGroupAt,
  convertRowToTabGroup,
  deleteTab,
  deleteTabGroup,
  duplicateTab,
  isTabGroupRow,
  moveItemAcrossContainers,
  moveTab,
  renameTab,
  reorderTab,
  tabCount,
  tabHasContent,
} from "./tab-edit";

const BASE = `type: canvas
rows:
  - items:
      - component: c1
`;

describe("tab-edit YAML transforms", () => {
  it("duplicateTab inserts a copy after the original with a (copy) label", () => {
    const doc = parseDocument(`type: canvas
rows:
  - tabs:
      - label: Overview
        rows:
          - items:
              - component: a
      - label: Detail
        rows: []
`);
    const newIndex = duplicateTab(doc, 0, 0);
    expect(newIndex).toBe(1);

    const tabs = doc.toJSON().rows[0].tabs;
    expect(tabs).toHaveLength(3);
    expect(tabs[1].label).toBe("Overview (copy)");
    // The copy carries the original's rows/content.
    expect(tabs[1].rows).toEqual([{ items: [{ component: "a" }] }]);
    // The original and the tab that followed it are preserved and in order.
    expect(tabs[0].label).toBe("Overview");
    expect(tabs[2].label).toBe("Detail");
  });

  it("deleteTabGroup removes the whole group entry", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: header
  - tabs:
      - label: A
        rows:
          - items:
              - component: a
`);
    expect(deleteTabGroup(doc, 1)).toBe(true);

    const json = doc.toJSON();
    expect(json.rows).toHaveLength(1);
    expect(json.rows[0]).toEqual({ items: [{ component: "header" }] });
    // Deleting a non-tab-group entry is a no-op.
    expect(deleteTabGroup(doc, 0)).toBe(false);
  });

  it("addTabGroup appends a group with one empty tab", () => {
    const doc = parseDocument(BASE);
    const index = addTabGroup(doc);

    expect(index).toBe(1);
    expect(isTabGroupRow(doc, 1)).toBe(true);
    expect(tabCount(doc, 1)).toBe(1);

    const json = doc.toJSON();
    expect(json.rows[1]).toEqual({
      tabs: [{ label: "Tab 1", rows: [] }],
    });
    // The pre-existing free row is untouched.
    expect(json.rows[0]).toEqual({ items: [{ component: "c1" }] });
  });

  it("addTabGroup creates the rows sequence when absent", () => {
    const doc = parseDocument(`type: canvas\n`);
    const index = addTabGroup(doc);
    expect(index).toBe(0);
    expect(doc.toJSON().rows).toHaveLength(1);
  });

  it("addTab appends a labeled empty tab", () => {
    const doc = parseDocument(BASE);
    addTabGroup(doc);
    const tabIndex = addTab(doc, 1);

    expect(tabIndex).toBe(1);
    expect(tabCount(doc, 1)).toBe(2);
    expect(doc.toJSON().rows[1].tabs[1]).toEqual({ label: "Tab 2", rows: [] });
  });

  it("addTab is a noop on a plain row", () => {
    const doc = parseDocument(BASE);
    expect(addTab(doc, 0)).toBe(-1);
  });

  it("renameTab updates the label", () => {
    const doc = parseDocument(BASE);
    addTabGroup(doc);
    renameTab(doc, 1, 0, "Overview");
    expect(doc.toJSON().rows[1].tabs[0].label).toBe("Overview");
  });

  it("deleteTab removes a tab when more than one remains", () => {
    const doc = parseDocument(BASE);
    addTabGroup(doc);
    addTab(doc, 1);
    expect(tabCount(doc, 1)).toBe(2);

    const result = deleteTab(doc, 1, 0);
    expect(result).toBe("removed-tab");
    expect(tabCount(doc, 1)).toBe(1);
    // The surviving tab is the one that was at index 1.
    expect(doc.toJSON().rows[1].tabs[0].label).toBe("Tab 2");
  });

  it("moveTab swaps a tab with its neighbor", () => {
    const doc = parseDocument(BASE);
    addTabGroup(doc);
    addTab(doc, 1);
    renameTab(doc, 1, 0, "First");
    renameTab(doc, 1, 1, "Second");

    moveTab(doc, 1, 0, 1);
    const labels = doc
      .toJSON()
      .rows[1].tabs.map((t: { label: string }) => t.label);
    expect(labels).toEqual(["Second", "First"]);
  });

  it("moveTab is a noop at the boundary", () => {
    const doc = parseDocument(BASE);
    addTabGroup(doc);
    addTab(doc, 1);
    moveTab(doc, 1, 0, -1); // already leftmost
    const labels = doc
      .toJSON()
      .rows[1].tabs.map((t: { label: string }) => t.label);
    expect(labels).toEqual(["Tab 1", "Tab 2"]);
  });

  it("convertRowToTabGroup wraps a plain row as Tab 1", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: a
      - component: b
`);
    const ok = convertRowToTabGroup(doc, 0);
    expect(ok).toBe(true);
    expect(isTabGroupRow(doc, 0)).toBe(true);

    const json = doc.toJSON();
    expect(json.rows[0].tabs).toHaveLength(1);
    expect(json.rows[0].tabs[0].label).toBe("Tab 1");
    expect(json.rows[0].tabs[0].rows).toEqual([
      { items: [{ component: "a" }, { component: "b" }] },
    ]);
  });

  it("convertRowToTabGroup is a noop on an existing tab group", () => {
    const doc = parseDocument(BASE);
    addTabGroup(doc);
    expect(convertRowToTabGroup(doc, 1)).toBe(false);
  });

  it("deleteTab unwraps the group into free rows when deleting the last tab", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: header
  - tabs:
      - label: Only
        rows:
          - items:
              - component: a
          - items:
              - component: b
`);
    expect(isTabGroupRow(doc, 1)).toBe(true);

    const result = deleteTab(doc, 1, 0);
    expect(result).toBe("unwrapped-group");

    const json = doc.toJSON();
    // The group block is replaced by its tab's two rows, after the header row.
    expect(json.rows).toHaveLength(3);
    expect(json.rows[0]).toEqual({ items: [{ component: "header" }] });
    expect(json.rows[1]).toEqual({ items: [{ component: "a" }] });
    expect(json.rows[2]).toEqual({ items: [{ component: "b" }] });
    expect(isTabGroupRow(doc, 1)).toBe(false);
  });

  it("addTabGroupAt inserts a group at the given index", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: a
  - items:
      - component: b
`);
    const index = addTabGroupAt(doc, 1);
    expect(index).toBe(1);

    const json = doc.toJSON();
    expect(json.rows).toHaveLength(3);
    expect(isTabGroupRow(doc, 1)).toBe(true);
    expect(json.rows[0]).toEqual({ items: [{ component: "a" }] });
    expect(json.rows[2]).toEqual({ items: [{ component: "b" }] });
  });

  it("reorderTab moves a tab from one position to another", () => {
    const doc = parseDocument(`type: canvas
rows:
  - tabs:
      - label: A
        rows: []
      - label: B
        rows: []
      - label: C
        rows: []
`);
    reorderTab(doc, 0, 0, 2);

    const labels = doc
      .toJSON()
      .rows[0].tabs.map((t: { label: string }) => t.label);
    expect(labels).toEqual(["B", "C", "A"]);
  });

  it("moveItemAcrossContainers drops a widget to the left of a widget inside a tab", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: outside_a
  - name: deep_dive
    tabs:
      - label: Overview
        rows:
          - items:
              - component: inside_b
`);
    const ok = moveItemAcrossContainers(
      doc,
      { rowsPath: ["rows"], row: 0, col: 0 },
      { rowsPath: ["rows", 1, "tabs", 0, "rows"], row: 0, col: 0 },
    );
    expect(ok).toBe(true);

    const json = doc.toJSON();
    // The source free row was removed; the tab group remains a tab group (no items key).
    expect(json.rows).toHaveLength(1);
    expect(json.rows[0].items).toBeUndefined();
    expect(json.rows[0].tabs).toBeDefined();
    // The widget joined the existing tab row to the LEFT of inside_b, losing nothing.
    expect(json.rows[0].tabs[0].rows[0].items).toEqual([
      { component: "outside_a" },
      { component: "inside_b" },
    ]);
  });

  it("moveItemAcrossContainers appends a new row when not dropping into a column", () => {
    const doc = parseDocument(`type: canvas
rows:
  - name: g
    tabs:
      - label: A
        rows:
          - items:
              - component: a
      - label: B
        rows:
          - items:
              - component: b
`);
    // Move from tab A (row 0) to tab B as a new row (col null).
    const ok = moveItemAcrossContainers(
      doc,
      { rowsPath: ["rows", 0, "tabs", 0, "rows"], row: 0, col: 0 },
      { rowsPath: ["rows", 0, "tabs", 1, "rows"], col: null },
    );
    expect(ok).toBe(true);

    const tabs = doc.toJSON().rows[0].tabs;
    // Tab A is now empty; tab B has its original row plus the moved one.
    expect(tabs[0].rows).toEqual([]);
    expect(tabs[1].rows).toEqual([
      { items: [{ component: "b" }] },
      { items: [{ component: "a" }] },
    ]);
  });

  it("tabHasContent reflects whether a tab has rows", () => {
    const doc = parseDocument(`type: canvas
rows:
  - tabs:
      - label: Empty
        rows: []
      - label: Full
        rows:
          - items:
              - component: a
`);
    expect(tabHasContent(doc, 0, 0)).toBe(false);
    expect(tabHasContent(doc, 0, 1)).toBe(true);
  });
});
