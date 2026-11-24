import { describe, expect, it } from "vitest";
import { parseDocument } from "yaml";
import {
  addBlockAt,
  blockHasContent,
  blockIndexOf,
  blockKindAt,
  combineConditions,
  defaultBlockName,
  deleteBlock,
  duplicateBlock,
  isBlockRow,
  isGeneratedBlockName,
  moveBlock,
  moveItemAcrossContainers,
  nameBlocks,
  renameBlock,
  setKeyBeforeContent,
  unwrapRows,
  wrapRow,
  type BlockKind,
} from "./block-edit";

const BASE = `type: canvas
rows:
  - items:
      - component: c1
`;

// The YAML of a block of each kind holding one row of the given component.
const BLOCK_YAML: Record<BlockKind, (component: string) => string> = {
  "tab-group": (component) => `tabs:
      - label: Group
        rows:
          - items:
              - component: ${component}`,
  container: (component) => `rows:
      - items:
          - component: ${component}`,
};

const KINDS: BlockKind[] = ["tab-group", "container"];

// The name the editor gives the first new block of each kind.
const FIRST_NAME: Record<BlockKind, string> = {
  "tab-group": "group-1",
  container: "container-1",
};

describe.each(KINDS)("block-edit YAML transforms for a %s", (kind) => {
  const block = BLOCK_YAML[kind];

  it("addBlockAt inserts an empty, named block at the given index", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: a
  - items:
      - component: b
`);
    expect(addBlockAt(doc, 1, kind)).toBe(1);

    const json = doc.toJSON();
    expect(json.rows).toHaveLength(3);
    expect(blockKindAt(doc, 1)).toBe(kind);
    expect(blockHasContent(doc, 1)).toBe(false);
    expect(json.rows[1]).toEqual(
      kind === "tab-group"
        ? { name: "group-1", tabs: [{ label: "Tab 1", rows: [] }] }
        : { name: "container-1", rows: [] },
    );
    expect(json.rows[0]).toEqual({ items: [{ component: "a" }] });
    expect(json.rows[2]).toEqual({ items: [{ component: "b" }] });
  });

  it("addBlockAt appends past the end and creates the rows sequence when absent", () => {
    const doc = parseDocument(BASE);
    expect(addBlockAt(doc, 5, kind)).toBe(1);
    expect(isBlockRow(doc, 1, kind)).toBe(true);

    const empty = parseDocument(`type: canvas\n`);
    expect(addBlockAt(empty, 0, kind)).toBe(0);
    expect(empty.toJSON().rows).toHaveLength(1);
    expect(isBlockRow(empty, 0, kind)).toBe(true);
  });

  it("renameBlock sets the block name and ignores an empty one", () => {
    const doc = parseDocument(`type: canvas
rows:
  - name: story
    ${block("a")}
`);
    expect(renameBlock(doc, 0, " Sales overview ")).toBe(true);
    expect(doc.toJSON().rows[0].name).toBe("Sales overview");

    // Every block keeps a name.
    expect(renameBlock(doc, 0, "  ")).toBe(false);
    expect(doc.toJSON().rows[0].name).toBe("Sales overview");

    // Plain rows have no name.
    const plain = parseDocument(BASE);
    expect(renameBlock(plain, 0, "x")).toBe(false);
    expect(plain.toString()).toBe(BASE);
  });

  it("deleteBlock removes the whole block entry", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: header
  - ${block("a")}
`);
    expect(deleteBlock(doc, 1)).toBe(true);

    const json = doc.toJSON();
    expect(json.rows).toEqual([{ items: [{ component: "header" }] }]);
    // Deleting a plain row is a no-op.
    expect(deleteBlock(doc, 0)).toBe(false);
  });

  it("duplicateBlock inserts a full copy of the block after the original", () => {
    const doc = parseDocument(`type: canvas
rows:
  - name: story
    if: '{{ .user.admin }}'
    ${block("a")}
  - items:
      - component: b
`);
    expect(duplicateBlock(doc, 0)).toBe(FIRST_NAME[kind]);

    const json = doc.toJSON();
    expect(json.rows).toHaveLength(3);
    // The copy carries the content and the condition under a new name.
    expect(json.rows[1].name).toBe(FIRST_NAME[kind]);
    expect(json.rows[1].if).toBe("{{ .user.admin }}");
    const content = kind === "tab-group" ? "tabs" : "rows";
    expect(json.rows[1][content]).toEqual(json.rows[0][content]);
    // The original and the row that followed it are untouched.
    expect(json.rows[0].name).toBe("story");
    expect(json.rows[2]).toEqual({ items: [{ component: "b" }] });
  });

  it("duplicateBlock is a noop on a plain row", () => {
    const doc = parseDocument(BASE);
    expect(duplicateBlock(doc, 0)).toBeUndefined();
    expect(doc.toString()).toBe(BASE);
  });

  it("moveBlock moves a block to a drop slot above or below other blocks", () => {
    const yaml = `type: canvas
rows:
  - items:
      - component: a
  - ${block("g")}
  - items:
      - component: b
  - items:
      - component: c
`;

    // Drop slot 0 is above the first block.
    let doc = parseDocument(yaml);
    expect(moveBlock(doc, 1, 0)).toBe(0);
    expect(blockKindAt(doc, 0)).toBe(kind);
    expect(doc.toJSON().rows[1]).toEqual({ items: [{ component: "a" }] });

    // Drop slot 4 is below the last block; the block ends up at index 3.
    doc = parseDocument(yaml);
    expect(moveBlock(doc, 1, 4)).toBe(3);
    expect(blockKindAt(doc, 3)).toBe(kind);

    // Drop slot 3 (between b and c) moves the block one down.
    doc = parseDocument(yaml);
    expect(moveBlock(doc, 1, 3)).toBe(2);
    expect(doc.toJSON().rows[1]).toEqual({ items: [{ component: "b" }] });
    expect(blockKindAt(doc, 2)).toBe(kind);
  });

  it("moveBlock is a noop for its own adjacent slots, plain rows and bad indices", () => {
    const yaml = `type: canvas
rows:
  - items:
      - component: a
  - ${block("g")}
`;
    const doc = parseDocument(yaml);
    // The slots directly above and below the block leave the order unchanged.
    expect(moveBlock(doc, 1, 1)).toBe(-1);
    expect(moveBlock(doc, 1, 2)).toBe(-1);
    // Only blocks move as a unit.
    expect(moveBlock(doc, 0, 2)).toBe(-1);
    expect(moveBlock(doc, 1, 3)).toBe(-1);
    expect(moveBlock(doc, 1, -1)).toBe(-1);
    expect(doc.toString()).toBe(yaml);
  });

  it("wrapRow turns a plain row into the block's only content", () => {
    const doc = parseDocument(`type: canvas
rows:
  - if: '{{ .user.admin }}'
    items:
      - component: a
      - component: b
`);
    expect(wrapRow(doc, 0, kind)).toBe(true);
    expect(blockKindAt(doc, 0)).toBe(kind);

    const row = {
      if: "{{ .user.admin }}",
      items: [{ component: "a" }, { component: "b" }],
    };
    const json = doc.toJSON();
    if (kind === "tab-group") {
      expect(json.rows[0].tabs).toEqual([{ label: "Tab 1", rows: [row] }]);
    } else {
      expect(json.rows[0].rows).toEqual([row]);
    }
    // The block gets a name but no condition: the row keeps its own.
    expect(json.rows[0].name).toBe(FIRST_NAME[kind]);
    expect(json.rows[0].if).toBeUndefined();
  });

  it("wrapRow is a noop on an existing block", () => {
    for (const other of KINDS) {
      const doc = parseDocument(BASE);
      addBlockAt(doc, 1, other);
      const before = doc.toString();
      expect(wrapRow(doc, 1, kind)).toBe(false);
      expect(doc.toString()).toBe(before);
    }
  });

  it("blockHasContent reflects whether the block has any rows", () => {
    const doc = parseDocument(`type: canvas
rows:
  - ${block("a")}
`);
    expect(blockHasContent(doc, 0)).toBe(true);
    addBlockAt(doc, 1, kind);
    expect(blockHasContent(doc, 1)).toBe(false);
    addBlockAt(doc, 2, "tab-group");
    expect(blockHasContent(doc, 2)).toBe(false);
  });
});

describe("block kinds", () => {
  it("tells blocks apart from plain rows by their content key", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items: []
  - tabs: []
  - rows: []
  - height: 40px
`);
    expect(blockKindAt(doc, 0)).toBeUndefined();
    expect(blockKindAt(doc, 1)).toBe("tab-group");
    expect(blockKindAt(doc, 2)).toBe("container");
    expect(blockKindAt(doc, 3)).toBeUndefined();
    expect(blockKindAt(doc, 4)).toBeUndefined();
    expect(isBlockRow(doc, 2)).toBe(true);
    expect(isBlockRow(doc, 2, "tab-group")).toBe(false);
  });

  it("recognizes a block by its content list, whatever `items` holds", () => {
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
  - rows:
    items:
      - component: b
`);
    expect(blockKindAt(doc, 0)).toBe("container");
    expect(blockKindAt(doc, 1)).toBe("tab-group");
    // A `rows` key without a list is a plain row, as the parser reads it.
    expect(blockKindAt(doc, 2)).toBeUndefined();
  });

  it("names unnamed blocks like the parser does", () => {
    expect(defaultBlockName("tab-group", 3)).toBe("group-3");
    expect(defaultBlockName("container", 0)).toBe("container-0");
  });
});

describe("block names", () => {
  it("renameBlock rejects a name another block has", () => {
    const doc = parseDocument(`type: canvas
rows:
  - name: story
    rows: []
  - name: details
    tabs: []
  - rows: []
`);
    expect(renameBlock(doc, 1, "story")).toBe(false);
    // The parser names the unnamed block "container-2".
    expect(renameBlock(doc, 0, "container-2")).toBe(false);
    expect(doc.toJSON().rows.map((r: { name?: string }) => r.name)).toEqual([
      "story",
      "details",
      undefined,
    ]);
    // A block can keep its own name.
    expect(renameBlock(doc, 0, " story ")).toBe(true);
  });

  it("nameBlocks writes the default name of each unnamed block before its content", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: a
  - tabs:
      - label: A
        rows: []
  - name: story
    rows: []
  - if: "{{ .user.admin }}"
    rows: []
`);
    nameBlocks(doc);
    expect(doc.toString()).toBe(`type: canvas
rows:
  - items:
      - component: a
  - name: group-1
    tabs:
      - label: A
        rows: []
  - name: story
    rows: []
  - if: "{{ .user.admin }}"
    name: container-3
    rows: []
`);
  });

  it("keeps each block's name when a named document's blocks move", () => {
    const doc = nameBlocks(
      parseDocument(`type: canvas
rows:
  - rows: []
  - rows: []
`),
    );
    moveBlock(doc, 1, 0);
    expect(doc.toJSON().rows.map((row: { name: string }) => row.name)).toEqual([
      "container-1",
      "container-0",
    ]);
  });

  it("gives a new block the smallest number no block has", () => {
    // container-1 is taken by name, container-2 by the default of the unnamed container.
    const yaml = `type: canvas
rows:
  - name: container-1
    rows: []
  - items:
      - component: a
  - rows: []
`;
    let doc = parseDocument(yaml);
    addBlockAt(doc, 3, "container");
    expect(doc.toJSON().rows[3].name).toBe("container-3");

    doc = parseDocument(yaml);
    wrapRow(doc, 1, "container");
    expect(doc.toJSON().rows[1].name).toBe("container-3");

    doc = parseDocument(yaml);
    expect(duplicateBlock(doc, 0)).toBe("container-3");

    // Numbers are per kind.
    doc = parseDocument(yaml);
    addBlockAt(doc, 0, "tab-group");
    expect(doc.toJSON().rows[0].name).toBe("group-1");
  });

  it("follows YAML aliases like the parser, and never names a shared map", () => {
    // Two containers share their rows through an alias; the parser reads both as containers.
    const shared = parseDocument(`type: canvas
rows:
  - rows: &story []
  - rows: *story
`);
    nameBlocks(shared);
    expect(
      shared.toJSON().rows.map((row: { name: string }) => row.name),
    ).toEqual(["container-0", "container-1"]);
    addBlockAt(shared, 2, "container");
    expect(shared.toJSON().rows[2].name).toBe("container-2");

    // A block that is itself an alias stays unnamed, or both copies would get one name.
    // A new block still avoids the default names its copies get.
    const copied = parseDocument(`type: canvas
rows:
  - &story
    rows: []
  - *story
`);
    nameBlocks(copied);
    expect(copied.toJSON().rows[0].name).toBeUndefined();
    addBlockAt(copied, 2, "container");
    expect(copied.toJSON().rows[2].name).toBe("container-2");
  });

  it("blockIndexOf finds a block by its name, or by the default name of an unnamed one", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: a
  - name: story
    rows: []
  - tabs:
      - label: A
        rows: []
`);
    expect(blockIndexOf(doc, "story")).toBe(1);
    expect(blockIndexOf(doc, "group-2")).toBe(2);
    // A plain row has no name, and a default name only matches the block at its index.
    expect(blockIndexOf(doc, "container-0")).toBe(-1);
    expect(blockIndexOf(doc, "container-1")).toBe(-1);
  });

  it("isGeneratedBlockName tells generated names from chosen ones", () => {
    expect(isGeneratedBlockName("container", "container-0")).toBe(true);
    expect(isGeneratedBlockName("container", "container-12")).toBe(true);
    expect(isGeneratedBlockName("tab-group", "group-2")).toBe(true);
    expect(isGeneratedBlockName("container", "group-2")).toBe(false);
    expect(isGeneratedBlockName("container", "container-")).toBe(false);
    expect(isGeneratedBlockName("container", "container-2b")).toBe(false);
    expect(isGeneratedBlockName("container", "revenue_story")).toBe(false);
  });
});

describe("unwrapRows", () => {
  it("replaces a container with its rows, which keep the container's condition", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: header
  - name: story
    if: '{{ has "finance" .user.groups }}'
    rows:
      - items:
          - component: a
      - if: '{{ .user.admin }}'
        items:
          - component: b
  - items:
      - component: footer
`);
    expect(unwrapRows(doc, 1, ["rows", 1, "rows"])).toBe(2);

    const json = doc.toJSON();
    expect(json.rows).toEqual([
      { items: [{ component: "header" }] },
      { if: '{{ has "finance" .user.groups }}', items: [{ component: "a" }] },
      {
        if: '({{ has "finance" .user.groups }}) AND ({{ .user.admin }})',
        items: [{ component: "b" }],
      },
      { items: [{ component: "footer" }] },
    ]);
  });

  it("removes an empty container and leaves plain rows alone", () => {
    const doc = parseDocument(`type: canvas
rows:
  - rows: []
  - items:
      - component: a
`);
    expect(unwrapRows(doc, 1, ["rows", 1, "rows"])).toBe(-1);
    expect(unwrapRows(doc, 0, ["rows", 0, "rows"])).toBe(0);
    expect(doc.toJSON().rows).toEqual([{ items: [{ component: "a" }] }]);
  });

  // An admin-only container holds an anchored row; a finance-only container holds an alias
  // of it.
  const SHARED_ROW = `type: canvas
rows:
  - name: admin
    if: "{{ .user.admin }}"
    rows:
      - &private_row
        items:
          - component: secret
  - name: finance
    if: "{{ .user.finance }}"
    rows:
      - *private_row
`;
  const PRIVATE_ITEMS = [{ component: "secret" }];

  it("gives an aliased row the condition of the container it leaves", () => {
    const doc = parseDocument(SHARED_ROW);
    expect(unwrapRows(doc, 1, ["rows", 1, "rows"])).toBe(1);

    // The alias becomes a copy with the finance condition; the anchor keeps its own.
    expect(doc.toJSON().rows).toEqual([
      {
        name: "admin",
        if: "{{ .user.admin }}",
        rows: [{ items: PRIVATE_ITEMS }],
      },
      { if: "{{ .user.finance }}", items: PRIVATE_ITEMS },
    ]);
    expect(parseDocument(doc.toString()).errors).toEqual([]);
  });

  it("leaves the aliases of a row alone when its container is ungrouped", () => {
    const doc = parseDocument(SHARED_ROW);
    expect(unwrapRows(doc, 0, ["rows", 0, "rows"])).toBe(1);

    // The finance container keeps an unconditioned copy of the row.
    expect(doc.toJSON().rows).toEqual([
      { if: "{{ .user.admin }}", items: PRIVATE_ITEMS },
      {
        name: "finance",
        if: "{{ .user.finance }}",
        rows: [{ items: PRIVATE_ITEMS }],
      },
    ]);
    expect(parseDocument(doc.toString()).errors).toEqual([]);
  });
});

describe("moveItemAcrossContainers", () => {
  it("drops a widget to the left of a widget inside a tab", () => {
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

  it("appends a new row when not dropping into a column", () => {
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

  it("keeps the conditions a widget leaves when it moves out of a container", () => {
    const doc = parseDocument(`type: canvas
rows:
  - if: "{{ .user.admin }}"
    rows:
      - if: "{{ .user.domain }} = 'acme.com'"
        items:
          - width: 12
            markdown:
              content: secret
  - items:
      - component: free
`);
    expect(
      moveItemAcrossContainers(
        doc,
        { rowsPath: ["rows", 0, "rows"], row: 0, col: 0 },
        { rowsPath: ["rows"], row: 2, col: null },
      ),
    ).toBe(true);
    // The container's and the row's conditions now sit on the widget itself.
    expect(doc.toJSON().rows[2].items[0]).toEqual({
      width: 12,
      if: "({{ .user.admin }}) AND ({{ .user.domain }} = 'acme.com')",
      markdown: { content: "secret" },
    });
  });

  it("only adds the conditions the destination doesn't already impose", () => {
    const doc = parseDocument(`type: canvas
rows:
  - if: "{{ .user.admin }}"
    tabs:
      - label: A
        if: "{{ .user.a }}"
        rows:
          - items:
              - if: "{{ .user.own }}"
                component: x
      - label: B
        rows:
          - if: "{{ .user.b }}"
            items:
              - component: y
`);
    // Into tab B's row, which has a condition of its own: the group's condition is shared.
    moveItemAcrossContainers(
      doc,
      { rowsPath: ["rows", 0, "tabs", 0, "rows"], row: 0, col: 0 },
      { rowsPath: ["rows", 0, "tabs", 1, "rows"], row: 0, col: 0 },
    );
    expect(doc.toJSON().rows[0].tabs[1].rows[0].items[0]).toEqual({
      if: "({{ .user.a }}) AND ({{ .user.own }})",
      component: "x",
    });

    // Back into tab A as a new row: only tab B's row condition is left behind.
    moveItemAcrossContainers(
      doc,
      { rowsPath: ["rows", 0, "tabs", 1, "rows"], row: 0, col: 0 },
      { rowsPath: ["rows", 0, "tabs", 0, "rows"], col: null },
    );
    expect(doc.toJSON().rows[0].tabs[0].rows[0].items[0]).toEqual({
      if: "({{ .user.b }}) AND (({{ .user.a }}) AND ({{ .user.own }}))",
      component: "x",
    });
  });

  it("moves a widget into an empty container and back out", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: a
      - component: b
  - name: story
    rows: []
`);
    expect(
      moveItemAcrossContainers(
        doc,
        { rowsPath: ["rows"], row: 0, col: 1 },
        { rowsPath: ["rows", 1, "rows"], row: 0, col: null },
      ),
    ).toBe(true);
    expect(doc.toJSON().rows).toEqual([
      { items: [{ component: "a" }] },
      { name: "story", rows: [{ items: [{ component: "b" }] }] },
    ]);

    // Moving the container's only widget out leaves the container empty, not removed.
    expect(
      moveItemAcrossContainers(
        doc,
        { rowsPath: ["rows", 1, "rows"], row: 0, col: 0 },
        { rowsPath: ["rows"], row: 0, col: 1 },
      ),
    ).toBe(true);
    expect(doc.toJSON().rows).toEqual([
      { items: [{ component: "a" }, { component: "b" }] },
      { name: "story", rows: [] },
    ]);
  });

  it("gives an aliased widget the conditions it leaves, without changing its anchor", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - &chart
        component: c
  - name: g
    if: "{{ .user.admin }}"
    tabs:
      - label: A
        rows:
          - items:
              - *chart
`);
    moveItemAcrossContainers(
      doc,
      { rowsPath: ["rows", 1, "tabs", 0, "rows"], row: 0, col: 0 },
      { rowsPath: ["rows"], row: 0, col: 1 },
    );
    expect(doc.toJSON().rows[0].items).toEqual([
      { component: "c" },
      { if: "{{ .user.admin }}", component: "c" },
    ]);
    expect(parseDocument(doc.toString()).errors).toEqual([]);
  });
});

describe("readable YAML", () => {
  it("puts a new name and condition before the block's content", () => {
    const doc = parseDocument(`type: canvas
rows:
  - rows:
      - items:
          - component: a
`);
    renameBlock(doc, 0, "story");
    setKeyBeforeContent(doc, ["rows", 0], "if", "{{ .user.admin }}");
    expect(doc.toString()).toBe(`type: canvas
rows:
  - name: story
    if: "{{ .user.admin }}"
    rows:
      - items:
          - component: a
`);
    // An existing key keeps its place.
    renameBlock(doc, 0, "revenue");
    expect(doc.toString()).toContain("  - name: revenue\n    if:");
  });

  it("puts a new condition before an item's component", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - width: 6
        markdown:
          content: a
`);
    setKeyBeforeContent(doc, ["rows", 0, "items", 0], "if", "true");
    expect(doc.toString()).toContain(`      - width: 6
        if: "true"
        markdown:`);
  });

  it("writes the rows dropped into an empty block in block style", () => {
    const doc = parseDocument(`type: canvas
rows:
  - items:
      - component: a
      - component: b
  - rows: []
`);
    moveItemAcrossContainers(
      doc,
      { rowsPath: ["rows"], row: 0, col: 1 },
      { rowsPath: ["rows", 1, "rows"], row: 0, col: null },
    );
    expect(doc.toString()).toBe(`type: canvas
rows:
  - items:
      - component: a
  - rows:
      - items:
          - component: b
`);
  });

  it("puts the condition an ungrouped row inherits before its items", () => {
    const doc = parseDocument(`type: canvas
rows:
  - if: "{{ .user.admin }}"
    rows:
      - items:
          - component: a
        height: 40px
`);
    unwrapRows(doc, 0, ["rows", 0, "rows"]);
    expect(doc.toString()).toBe(`type: canvas
rows:
  - if: "{{ .user.admin }}"
    items:
      - component: a
    height: 40px
`);
  });
});

describe("combineConditions", () => {
  it("skips empty conditions", () => {
    expect(combineConditions(undefined, "", null)).toBeUndefined();
    expect(combineConditions(undefined, "a")).toBe("a");
    expect(combineConditions(false, "a")).toBe("(false) AND (a)");
  });
});
