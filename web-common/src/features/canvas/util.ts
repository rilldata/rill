import * as defaults from "./constants";
import type { Vector } from "./types";

export const vector = {
  add: (add: Vector, initial: Vector): Vector => {
    return [add[0] + initial[0], add[1] + initial[1]];
  },
  multiply: (vector: Vector, multiplier: Vector): Vector => {
    return [vector[0] * multiplier[0], vector[1] * multiplier[1]];
  },
  subtract: (minuend: Vector, subtrahend: Vector): Vector => {
    return [minuend[0] - subtrahend[0], minuend[1] - subtrahend[1]];
  },
  absolute: (vector: Vector): Vector => {
    return [Math.abs(vector[0]), Math.abs(vector[1])];
  },
  divide: (vector: Vector, divisor: Vector): Vector => {
    return [vector[0] / divisor[0], vector[1] / divisor[1]];
  },
};

export function isString(value: unknown): value is string {
  return typeof value === "string";
}

interface PositionedItem {
  x: number;
  y: number;
  width: number;
  height: number;
}

/**
 * Finds the next available position for a new component in the canvas.
 * Height is not considered in placement strategy because:
 * 1. Components in the same row can have different heights (flexible row height)
 * 2. Taller components don't block shorter components from being placed next to them
 * 3. The canvas layout system automatically handles vertical spacing
 *
 * The placement strategy prioritizes:
 * 1. Filling existing rows (from top to bottom) by:
 *    a. First checking for space at the end of each row
 *    b. Then checking for gaps within each row
 * 2. Only creating a new row if no space is found in existing rows
 */
export function findNextAvailablePosition(
  existingItems: PositionedItem[],
  newWidth: number,
): [number, number] {
  if (!existingItems?.length) {
    return [0, 0];
  }

  // Group items by row (y coordinate) for efficient row-based placement
  const rowGroups = new Map<number, PositionedItem[]>();
  existingItems.forEach((item) => {
    const items = rowGroups.get(item.y) || [];
    items.push(item);
    rowGroups.set(item.y, items);
  });

  // Process rows from top to bottom
  const rows = Array.from(rowGroups.entries()).sort(([y1], [y2]) => y1 - y2);

  // First pass: look for space at the end of existing rows (simplest placement)
  for (const [y, items] of rows) {
    const rightmostX = Math.max(...items.map((item) => item.x + item.width));
    if (rightmostX + newWidth <= defaults.COLUMN_COUNT) {
      return [rightmostX, y];
    }
  }

  // Second pass: look for gaps within existing rows
  for (const [y, items] of rows) {
    const sortedItems = items.sort((a, b) => a.x - b.x);

    // Check each possible position from left to right
    let x = 0;
    for (const item of sortedItems) {
      // If there's enough space before the current item, use that position
      if (x + newWidth <= item.x) {
        return [x, y];
      }
      x = item.x + item.width;
    }
  }

  // Last resort: create a new row below all existing content
  const lastRow = rows[rows.length - 1];
  // Use the height of the last row to determine the y-coordinate for the new row
  const newY = lastRow ? lastRow[0] + lastRow[1][0].height : 0;
  return [0, newY];
}
