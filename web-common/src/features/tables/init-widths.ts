const SAMPLE_SIZE = 30;
import { clamp } from "@rilldata/web-common/lib/clamp";
import type { VirtualizedTableColumns } from "@rilldata/web-common/components/virtualized-table/types";

export function initColumnWidths<K>(params: {
  columns: VirtualizedTableColumns[];
  rows: K[];
  nameAccessor: keyof VirtualizedTableColumns;
  minWidth: number;
  maxWidth: number;
  resizableColumns: boolean;
}) {
  const { columns, rows, nameAccessor, minWidth, maxWidth, resizableColumns } =
    params;
  let sampleIndexes: number[] = [];

  if (rows.length < 30) {
    sampleIndexes = rows.map((_, i) => i);
  } else {
    // Deterministic pseudo-random sample
    for (let i = 0; i < SAMPLE_SIZE; i++) {
      if (i <= 10) {
        sampleIndexes.push(i);
      } else if (i >= SAMPLE_SIZE - 10) {
        sampleIndexes.push(rows.length - i - 1);
      } else {
        sampleIndexes.push(Math.floor(rows.length / 2) - 5 + i);
      }
    }
  }

  return columns.map((column) => {
    if (!resizableColumns) return minWidth;

    const name = String(column[nameAccessor]) ?? "";
    const columnNameLength = name.length ?? 8;

    let totalValueCharacters = 0;

    sampleIndexes.forEach((index) => {
      totalValueCharacters += String(rows[index][name]).length;
    });

    const averageCharacterLength = Math.ceil(
      totalValueCharacters / SAMPLE_SIZE,
    );

    const lengthBasis = Math.max(columnNameLength, averageCharacterLength);

    const pixelLength = Math.ceil(lengthBasis * 7);

    if (columnNameLength > averageCharacterLength) {
      return clamp(minWidth, pixelLength, maxWidth) + 80;
    } else {
      return clamp(minWidth, pixelLength, maxWidth) + 80;
    }
  });
}
