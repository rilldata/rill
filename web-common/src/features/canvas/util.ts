import * as defaults from "./constants";
import type { PositionedItem, Vector } from "./types";
import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";

interface RowGroup {
  y: number;
  height: number;
  items: V1CanvasItem[];
}

interface GridItem {
  position: [number, number]; // [x, y]
  size: [number, number]; // [width, height]
  node: any; // YAML node
}

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

// Allowed widths for components
const ALLOWED_WIDTHS = [3, 4, 6, 8, 9, 12];

// Snap to the closest valid width
function getValidWidth(newWidth: number): number {
  return ALLOWED_WIDTHS.reduce((closest, width) =>
    Math.abs(width - newWidth) < Math.abs(closest - newWidth) ? width : closest,
  );
}

// Check if a position is free of collisions
function isPositionFree(
  existingItems: PositionedItem[],
  x: number,
  y: number,
  width: number,
  height: number,
): boolean {
  return !existingItems.some((item) => {
    const overlapsInX = x < item.x + item.width && x + width > item.x;
    const overlapsInY = y < item.y + item.height && y + height > item.y;
    return overlapsInX && overlapsInY;
  });
}

// Row-based grouping with sequential placement with collision checks
export function findNextAvailablePosition(
  existingItems: PositionedItem[],
  newWidth: number,
  newHeight: number,
): [number, number] {
  const validWidth = getValidWidth(newWidth);

  if (!existingItems?.length) {
    return [0, 0];
  }

  // Group items by row (y coordinate)
  const rowGroups = new Map<number, PositionedItem[]>();
  existingItems.forEach((item) => {
    const items = rowGroups.get(item.y) || [];
    items.push(item);
    rowGroups.set(item.y, items);
  });

  // Sort rows top-to-bottom
  const rows = Array.from(rowGroups.entries()).sort(([y1], [y2]) => y1 - y2);

  // First pass: find space at the end of rows
  for (const [y, items] of rows) {
    const rightmostX = Math.max(...items.map((item) => item.x + item.width), 0);
    if (rightmostX + validWidth <= defaults.COLUMN_COUNT) {
      if (isPositionFree(existingItems, rightmostX, y, validWidth, newHeight)) {
        return [rightmostX, y];
      }
    }
  }

  // Second pass: find gaps within rows
  for (const [y, items] of rows) {
    const sortedItems = items.sort((a, b) => a.x - b.x);

    let x = 0;
    for (const item of sortedItems) {
      if (
        x + validWidth <= item.x &&
        isPositionFree(existingItems, x, y, validWidth, newHeight)
      ) {
        return [x, y];
      }
      x = item.x + item.width;
    }

    // Check after the last item in the row
    if (
      x + validWidth <= defaults.COLUMN_COUNT &&
      isPositionFree(existingItems, x, y, validWidth, newHeight)
    ) {
      return [x, y];
    }
  }

  // Final pass: add a new row
  const lastRowY = Math.max(
    ...existingItems.map((item) => item.y + item.height),
    0,
  );
  const newY = lastRowY; // Place the new row below the tallest existing item
  return [0, newY];
}

export function isValidItem(item: V1CanvasItem): item is V1CanvasItem & {
  x: number;
  y: number;
  width: number;
  height: number;
} {
  return (
    item?.x !== undefined &&
    item?.y !== undefined &&
    item?.width !== undefined &&
    item?.height !== undefined
  );
}

export function validateItemPositions(items: V1CanvasItem[]): void {
  items.forEach((item) => {
    if (item.x !== undefined && item.width !== undefined) {
      item.x = Math.min(
        Math.max(0, item.x),
        defaults.COLUMN_COUNT - item.width,
      );
    }
  });
}

export function groupItemsByRow(items: V1CanvasItem[]): RowGroup[] {
  const rows: RowGroup[] = [];

  items.forEach((item) => {
    const existingRow = rows.find((row) => row.y === item.y);
    if (existingRow) {
      existingRow.items.push(item);
      existingRow.height = Math.max(existingRow.height ?? 0, item.height ?? 0);
    } else {
      rows.push({
        y: item.y ?? 0,
        height: item.height ?? 0,
        items: [item],
      });
    }
  });

  return rows.sort((a, b) => a.y - b.y);
}

export function flattenRowGroups(rows: RowGroup[]): V1CanvasItem[] {
  return rows.flatMap((row) => row.items);
}

export function convertToGridItems(yamlItems: any[]): GridItem[] {
  return yamlItems.map((item) => ({
    position: [item.get("x"), item.get("y")],
    size: [item.get("width"), item.get("height")],
    node: item,
  }));
}

export function sortItemsByPosition(items: GridItem[]): GridItem[] {
  return items.sort((a, b) => {
    // Sort by Y first, then X for items in the same row
    if (a.position[1] === b.position[1]) {
      return a.position[0] - b.position[0];
    }
    return a.position[1] - b.position[1];
  });
}

export function compactGrid(items: GridItem[]) {
  let currentY = 0;
  let lastRowHeight = 0;
  let lastY = -1;

  items.forEach(({ position, size, node }) => {
    if (position[1] !== lastY) {
      // Starting a new row
      currentY += lastRowHeight;
      lastRowHeight = size[1];
      lastY = position[1];
    } else {
      // Same row - update max height if needed
      lastRowHeight = Math.max(lastRowHeight, size[1]);
    }

    // Update item's Y position
    node.set("y", currentY);
  });
}
