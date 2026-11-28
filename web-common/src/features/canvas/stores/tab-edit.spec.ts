import { describe, it, expect } from "vitest";
import { parseDocument } from "yaml";
import { addBlockAt } from "./block-edit";
import {
  addTab,
  deleteTab,
  duplicateTab,
  isTabGroupRow,
  moveTab,
  renameTab,
  reorderTab,
  setTabName,
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
      - name: ov
        label: Overview
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
    // The copy carries the original's rows/content but not its URL name.
    expect(tabs[1].rows).toEqual([{ items: [{ component: "a" }] }]);
    expect(tabs[1].name).toBeUndefined();
    // The original and the tab that followed it are preserved and in order.
    expect(tabs[0].label).toBe("Overview");
    expect(tabs[2].label).toBe("Detail");
  });

  it("duplicateTab gives named items in the copy names of their own", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - name: chart_copy
        component: a
  - tabs:
      - label: Overview
        controls:
          - filter: publisher
        rows:
          - items:
              - name: chart
                controls:
                  - measure: total_bids
                    options: [total_bids, spend]
                component: b
              - component: c
`);
    duplicateTab(doc, 1, 0);

    const [original, copy] = doc.toJSON().rows[1].tabs;
    expect(original.rows[0].items[0].name).toBe("chart");
    // "chart_copy" is taken by an item outside the group.
    expect(copy.rows[0].items[0].name).toBe("chart_copy2");
    expect(copy.rows[0].items[0].controls).toEqual(
      original.rows[0].items[0].controls,
    );
    expect(copy.rows[0].items[1].name).toBeUndefined();
    // The tab's own controls are carried over.
    expect(copy.controls).toEqual([{ filter: "publisher" }]);
  });

  it("addTab appends a labeled empty tab", () => {
    const doc = parseDocument(BASE);
    addBlockAt(doc, 1, "tab-group");
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
    addBlockAt(doc, 1, "tab-group");
    renameTab(doc, 1, 0, "Overview");
    expect(doc.toJSON().rows[1].tabs[0].label).toBe("Overview");
  });

  it("setTabName sets and clears the tab URL name", () => {
    const doc = parseDocument(BASE);
    addBlockAt(doc, 1, "tab-group");
    setTabName(doc, 1, 0, "overview");
    expect(doc.toJSON().rows[1].tabs[0].name).toBe("overview");
    setTabName(doc, 1, 0, "  ");
    expect("name" in doc.toJSON().rows[1].tabs[0]).toBe(false);
  });

  it("deleteTab removes a tab when more than one remains", () => {
    const doc = parseDocument(BASE);
    addBlockAt(doc, 1, "tab-group");
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
    addBlockAt(doc, 1, "tab-group");
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
    addBlockAt(doc, 1, "tab-group");
    addTab(doc, 1);
    moveTab(doc, 1, 0, -1); // already leftmost
    const labels = doc
      .toJSON()
      .rows[1].tabs.map((t: { label: string }) => t.label);
    expect(labels).toEqual(["Tab 1", "Tab 2"]);
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

  it("deleteTab keeps the visibility of the unwrapped rows", () => {
    const doc = parseDocument(`type: canvas
rows:
  - if: '{{ has "finance" .user.groups }}'
    tabs:
      - label: Only
        if: '{{ .user.admin }}'
        rows:
          - items:
              - component: a
          - if: "'{{ .user.domain }}' = 'acme.com'"
            items:
              - component: b
`);

    expect(deleteTab(doc, 0, 0)).toBe("unwrapped-group");

    const json = doc.toJSON();
    expect(json.rows).toHaveLength(2);
    expect(json.rows[0].if).toBe(
      `({{ has "finance" .user.groups }}) AND ({{ .user.admin }})`,
    );
    expect(json.rows[1].if).toBe(
      `({{ has "finance" .user.groups }}) AND ({{ .user.admin }}) AND ('{{ .user.domain }}' = 'acme.com')`,
    );
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
