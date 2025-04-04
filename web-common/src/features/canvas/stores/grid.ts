import { get, writable } from "svelte/store";
import { Row } from "./row";
import type { V1CanvasRow } from "@rilldata/web-common/runtime-client";
import { COLUMN_COUNT } from "../layout-util";

export class Grid {
  private _rows = writable<Row[]>([]);

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

  moveItem(fromRow: Row, fromIndex: number, toRow: Row, toIndex: number) {
    if (fromIndex < 1 || toIndex < 1) return;

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
    if (fromIndex < 1) return null;

    const value = fromRow.getValue(fromIndex);
    if (value === null) return null;

    const newRow = this.addRow();
    newRow.add(value, 1);
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

      if (
        canvasRow.height !== undefined &&
        get(row.height) !== canvasRow.height
      ) {
        row.height.set(canvasRow.height);
      }

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
