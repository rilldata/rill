import * as defaults from "./constants";

export function isString(value: unknown): value is string {
  return typeof value === "string";
}

type GridRow = {
  y: number;
  items: PositionedItem[];
};

interface PositionedItem {
  x: number;
  y: number;
  width: number;
  height: number;
}

/**
 * Finds the next available position for a new component in the grid
 */
export function findNextAvailablePosition(
  existingItems: PositionedItem[],
  newWidth: number,
  newHeight: number,
): [number, number] {
  if (!existingItems?.length) return [0, 0];

  const rows = existingItems.reduce((acc, item) => {
    if (!acc.has(item.y)) {
      acc.set(item.y, { y: item.y, items: [] });
    }
    acc.get(item.y)!.items.push(item);
    return acc;
  }, new Map<number, GridRow>());

  const sortedRows = Array.from(rows.values()).sort((a, b) => a.y - b.y);

  for (const row of sortedRows) {
    const sortedItems = row.items.sort((a, b) => a.x - b.x);

    let x = 0;
    for (const item of sortedItems) {
      if (item.x - x >= newWidth) {
        const hasOverlap = existingItems.some(
          (other) =>
            other.y !== row.y && // Different row
            other.y < row.y + newHeight && // Overlaps vertically
            other.y + other.height > row.y &&
            other.x < x + newWidth && // Overlaps horizontally
            other.x + other.width > x,
        );

        if (!hasOverlap) return [x, row.y];
      }
      x = item.x + item.width;
    }

    // Check if there's space at the end of the row
    if (x + newWidth <= defaults.DEFAULT_COLUMN_COUNT) {
      const hasOverlap = existingItems.some(
        (other) =>
          other.y !== row.y &&
          other.y < row.y + newHeight &&
          other.y + other.height > row.y &&
          other.x < x + newWidth &&
          other.x + other.width > x,
      );

      if (!hasOverlap) return [x, row.y];
    }
  }

  // If no gaps found, add to new row
  const lastRowY = Math.max(
    ...existingItems.map((item) => item.y + item.height),
    0,
  );

  return [0, lastRowY];
}
