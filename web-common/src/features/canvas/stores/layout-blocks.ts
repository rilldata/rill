import type { V1CanvasRow } from "@rilldata/web-common/runtime-client";
import {
  canvasItemInstanceId,
  positionKeyPrefix,
  type YAMLPath,
} from "../layout-util";
import type { CanvasEntity } from "./canvas-entity";
import { Grid } from "./grid";
import type { TabGroup } from "./tab-group";

/**
 * A named list of rows with its own grid: one tab of a tab group, or a container. It knows
 * the YAML path of its rows, so the component-path and transaction machinery can target its
 * rows the same way it targets the top-level rows.
 */
export class RowsBlock {
  /** Stable identifier: a tab's URL key, or a container's name. */
  name: string;
  /** User-facing label of a tab; empty for a container, which has none. */
  displayName: string;
  /** The rows. */
  grid: Grid;
  /** YAML path of the rows, e.g. ["rows", 2, "tabs", 0, "rows"] for a tab or ["rows", 1, "rows"] for a container. */
  rowsPath: YAMLPath;

  constructor(
    canvas: CanvasEntity,
    name: string,
    displayName: string,
    rowsPath: YAMLPath,
  ) {
    this.name = name;
    this.displayName = displayName;
    this.rowsPath = rowsPath;
    this.grid = new Grid(canvas);
  }

  /** Sync the grid from the block's spec rows, which live at rowsPath in the YAML. */
  updateRows(rows: V1CanvasRow[], rowsPath: YAMLPath) {
    this.rowsPath = rowsPath;
    const namePrefix = positionKeyPrefix(rowsPath);
    this.grid.updateFromCanvasRows(rows, (item, row, col) =>
      canvasItemInstanceId(item, row, col, namePrefix),
    );
  }
}

/**
 * A top-level layout block. The canvas body is an ordered list of these, mirroring the
 * heterogeneous `rows` array in the spec: a plain row, a tab group or a container.
 */
export type LayoutBlock =
  | { kind: "row"; rowIndex: number; freeRowIndex: number }
  | { kind: "tab-group"; rowIndex: number; group: TabGroup }
  | { kind: "container"; rowIndex: number; container: RowsBlock };

/** The name of a tab group or container block, or undefined for a plain row. */
export function layoutBlockName(block: LayoutBlock): string | undefined {
  switch (block.kind) {
    case "tab-group":
      return block.group.name;
    case "container":
      return block.container.name;
    default:
      return undefined;
  }
}

/** A key that identifies a block across layout updates, for keyed each blocks. */
export function layoutBlockKey(block: LayoutBlock): string {
  switch (block.kind) {
    case "tab-group":
      return `g-${block.group.name}`;
    case "container":
      return `c-${block.container.name}`;
    default:
      return `r-${block.rowIndex}`;
  }
}
