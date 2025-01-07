import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
import * as defaults from "./constants";
import type { DropPosition, RowGroup } from "./types";

export class Grid {
  private items: V1CanvasItem[];

  constructor(items: V1CanvasItem[]) {
    this.items = this.preventCollisions([...items]);
  }

  static isValidItem(item: V1CanvasItem): item is V1CanvasItem & {
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

  private preventCollisions(items: V1CanvasItem[]): V1CanvasItem[] {
    const newItems = [...items];

    // Sort all items by y position, then x position
    newItems.sort((a, b) => {
      const yDiff = (a.y ?? 0) - (b.y ?? 0);
      return yDiff !== 0 ? yDiff : (a.x ?? 0) - (b.x ?? 0);
    });

    // Process each row
    const rows = Grid.groupItemsByRow(newItems);
    rows.forEach((row) => {
      let currentX = 0;

      // Sort items within row by x position
      row.items
        .sort((a, b) => (a.x ?? 0) - (b.x ?? 0))
        .forEach((item) => {
          if (!Grid.isValidItem(item)) return;

          // Place item at currentX
          item.x = currentX;
          item.y = row.y;

          // Move currentX past this item
          currentX += item.width;
          if (currentX > defaults.COLUMN_COUNT) {
            currentX = defaults.COLUMN_COUNT;
          }
        });
    });

    return newItems;
  }

  // Row Management
  static groupItemsByRow(items: V1CanvasItem[]): RowGroup[] {
    const rows: RowGroup[] = [];

    items.forEach((item) => {
      const existingRow = rows.find((row) => row.y === item.y);
      if (existingRow) {
        existingRow.items.push(item);
        existingRow.height = Math.max(
          existingRow.height ?? 0,
          item.height ?? 0,
        );
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

  static leftAlignRow(row: RowGroup): void {
    let currentX = 0;
    row.items
      .sort((a, b) => (a.x ?? 0) - (b.x ?? 0))
      .forEach((item) => {
        item.x = currentX;
        currentX += item.width ?? defaults.COMPONENT_WIDTH;
      });
  }

  private validateItemPositions(items: V1CanvasItem[]): void {
    const rows = Grid.groupItemsByRow(items);
    rows.forEach((row) => Grid.leftAlignRow(row));

    items.forEach((item) => {
      if (item.x !== undefined && item.width !== undefined) {
        item.x = Math.min(
          Math.max(0, item.x),
          defaults.COLUMN_COUNT - item.width,
        );
      }
    });
  }

  public getDropPosition(
    mouseX: number,
    mouseY: number,
    targetRect: DOMRect,
  ): DropPosition {
    const zoneSize = 0.2; // 20% of element size for edge detection
    const bottomZone = targetRect.bottom - targetRect.height * zoneSize;
    const topZone = targetRect.top + targetRect.height * zoneSize;
    const leftZone = targetRect.left + targetRect.width * zoneSize;
    const rightZone = targetRect.right - targetRect.width * zoneSize;

    if (mouseY > bottomZone) return "bottom";
    if (mouseY < topZone) return "top";
    if (mouseX < leftZone) return "left";
    if (mouseX > rightZone) return "right";

    const distanceToLeft = mouseX - targetRect.left;
    const distanceToRight = targetRect.right - mouseX;
    return distanceToLeft < distanceToRight ? "left" : "right";
  }

  public moveItem(
    draggedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
    position: DropPosition,
    dragIndex: number,
  ): { items: V1CanvasItem[]; insertIndex: number } {
    if (!Grid.isValidItem(targetItem) || !Grid.isValidItem(draggedItem)) {
      throw new Error("Invalid items provided to moveItem");
    }

    const newItems = [...this.items];
    const [removedItem] = newItems.splice(dragIndex, 1);

    switch (position) {
      case "left":
        this.handleLeftDrop(removedItem, targetItem);
        break;
      case "right":
        this.handleRightDrop(removedItem, targetItem);
        break;
      case "top":
        this.handleTopDrop(removedItem, targetItem);
        break;
      case "bottom":
        this.handleBottomDrop(removedItem, targetItem);
        break;
    }

    // Find insert index based on new position
    const insertIndex = newItems.findIndex(
      (item) =>
        (item.y ?? 0) > (removedItem.y ?? 0) ||
        ((item.y ?? 0) === (removedItem.y ?? 0) &&
          (item.x ?? 0) > (removedItem.x ?? 0)),
    );

    // Insert the item at the correct position
    if (insertIndex === -1) {
      newItems.push(removedItem);
    } else {
      newItems.splice(insertIndex, 0, removedItem);
    }

    // Update matrix and prevent any collisions
    this.items = this.preventCollisions(newItems);

    return {
      items: this.items,
      insertIndex: insertIndex === -1 ? this.items.length - 1 : insertIndex,
    };
  }

  private handleRightDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
  ): void {
    if (!Grid.isValidItem(targetItem) || !Grid.isValidItem(removedItem)) return;

    // Initialize removedItem properties if undefined
    removedItem.x = targetItem.x + targetItem.width;
    removedItem.y = targetItem.y;
    removedItem.width = removedItem.width ?? defaults.COMPONENT_WIDTH;
    // Match target row height
    removedItem.height = targetItem.height;

    // If would overflow column count, adjust width
    if (removedItem.x + removedItem.width > defaults.COLUMN_COUNT) {
      removedItem.width = defaults.COLUMN_COUNT - removedItem.x;
    }
  }

  private handleLeftDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
  ): void {
    if (!Grid.isValidItem(targetItem) || !Grid.isValidItem(removedItem)) return;

    // Initialize removedItem properties if undefined
    removedItem.x = targetItem.x;
    removedItem.y = targetItem.y;
    removedItem.width = removedItem.width ?? defaults.COMPONENT_WIDTH;
    // Match target row height
    removedItem.height = targetItem.height;

    // Shift target item right
    targetItem.x = removedItem.x + removedItem.width;
  }

  private handleTopDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
  ): void {
    if (!Grid.isValidItem(targetItem) || !Grid.isValidItem(removedItem)) return;

    // Initialize removedItem properties if undefined
    removedItem.x = 0;
    removedItem.y = targetItem.y;
    removedItem.width = defaults.COLUMN_COUNT; // Full width for new row
    // Keep original height for new row
    removedItem.height = removedItem.height ?? defaults.COMPONENT_HEIGHT;

    // Move all items in this row and below down
    this.items.forEach((item) => {
      if ((item.y ?? 0) >= targetItem.y) {
        item.y = (item.y ?? 0) + removedItem.height;
      }
    });
  }

  private handleBottomDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
  ): void {
    if (!Grid.isValidItem(targetItem) || !Grid.isValidItem(removedItem)) return;

    // Initialize removedItem properties if undefined
    removedItem.x = 0;
    removedItem.y = targetItem.y + targetItem.height;
    removedItem.width = defaults.COLUMN_COUNT; // Full width for new row
    // Keep original height for new row
    removedItem.height = removedItem.height ?? defaults.COMPONENT_HEIGHT;

    // Move all items below target's row down
    this.items.forEach((item) => {
      if ((item.y ?? 0) > targetItem.y) {
        item.y = (item.y ?? 0) + removedItem.height;
      }
    });
  }
}

export const isValidItem = Grid.isValidItem;
export const groupItemsByRow = Grid.groupItemsByRow;
