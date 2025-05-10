import { get, writable } from "svelte/store";
import { Row } from "./row";
import type { V1CanvasRow } from "@rilldata/web-common/runtime-client";
import { COLUMN_COUNT } from "../layout-util";
import type { CanvasEntity } from "./canvas-entity";

export class Grid {
  private _rows = writable<Row[]>([]);

  constructor(public canvas: CanvasEntity) {}

  subscribe = this._rows.subscribe;

  addRow(): Row {
    const row = new Row(this);
    this._rows.update((r) => [...r, row]);
    return row;
  }

  addRowAt(index: number): Row {
    const row = new Row(this);
    this._rows.update((r) => {
      const copy = [...r];
      copy.splice(index, 0, row);
      return copy;
    });
    return row;
  }

  removeRow(row: Row) {
    this._rows.update((r) => r.filter((rItem) => rItem !== row));
  }

  snapshot(): string[][] {
    return get(this._rows).map((row) => row.snapshot());
  }

  removeItem(fromRow: Row, fromIndex: number) {
    if (fromIndex < 0) return;

    fromRow.remove(fromIndex);
  }

  moveItem(fromRow: Row, fromIndex: number, toRow: Row, toIndex: number) {
    if (fromIndex < 0 || toIndex < 0) return;

    if (fromRow === toRow) {
      fromRow.moveWithin(fromIndex, toIndex);
    } else {
      const value = fromRow.getValue(fromIndex);
      if (value === null) return;

      fromRow.remove(fromIndex);
      toRow.add(value, toIndex);
    }
  }

  copyItemToNewRow(fromRow: Row, fromIndex: number): Row | null {
    console.warn("Do not use without stable IDs");
    if (fromIndex < 0) return null;

    const id = fromRow.getValue(fromIndex);
    if (id === null) return null;

    const newId = this.canvas.duplicateItem(id);

    if (!newId) return null;

    const existingRowIndex = get(this._rows).indexOf(fromRow);
    const newRow = this.addRowAt(existingRowIndex + 1);

    newRow.add(newId, 0);
    return newRow;
  }

  slice(start: number, end: number) {
    this._rows.update((r) => r.slice(start, end));
  }

  updateFromCanvasRows(canvasRows: V1CanvasRow[]) {
    const currentRows = get(this._rows);

    let updatedRowCount = false;
    if (canvasRows.length < currentRows.length) {
      updatedRowCount = true;
      this.slice(0, canvasRows.length);
    }
    canvasRows.forEach((canvasRow, i) => {
      const row = currentRows[i] ?? this.addRow();

      row.height.set(canvasRow.height ?? 0);

      if (Array.isArray(canvasRow.items)) {
        const existingItemIds = get(row.items);
        const itemIds = canvasRow.items.map((item) => item.component ?? "");
        row.items.set(canvasRow.items.map((item) => item.component ?? ""));
        row.widths.set(canvasRow.items.map((item) => item.width ?? 25));

        if (
          existingItemIds.length !== itemIds.length ||
          itemIds.some((itemId, index) => itemId !== existingItemIds[index])
        ) {
          row.items.set(itemIds);
        }
        row.widths.set(
          canvasRow.items.map((item) => {
            return item.width ?? COLUMN_COUNT / (canvasRow.items?.length ?? 1);
          }),
        );
      }
    });

    return updatedRowCount;
  }

  refresh() {
    this._rows.update((r) => r);
  }
}
