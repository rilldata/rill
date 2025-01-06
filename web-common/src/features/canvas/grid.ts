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

    const rows = Grid.groupItemsByRow([...this.items]);
    let newItems = [...this.items];
    let insertIndex = this.items.indexOf(targetItem);

    const [draggedItemFull] = newItems.splice(dragIndex, 1);
    const removedItem = { ...draggedItemFull };

    switch (position) {
      case "left":
        this.handleLeftDrop(removedItem, targetItem, rows);
        insertIndex = this.items.indexOf(targetItem);
        break;
      case "right":
        this.handleRightDrop(removedItem, targetItem, rows);
        insertIndex = this.items.indexOf(targetItem) + 1;
        break;
      case "bottom":
        this.handleBottomDrop(removedItem, targetItem, rows);
        insertIndex = this.items.indexOf(targetItem) + 1;
        break;
      case "top":
        this.handleTopDrop(removedItem, targetItem, rows);
        insertIndex = this.items.findIndex((item) => item.y === targetItem.y);
        break;
    }

    newItems.splice(insertIndex, 0, removedItem);
    newItems = this.preventCollisions(newItems);
    this.validateItemPositions(newItems);

    return { items: newItems, insertIndex };
  }

  private handleLeftDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
    rows: RowGroup[],
  ): void {
    const targetRow = rows.find((row) => row.y === targetItem.y);
    if (!targetRow) return;

    // Fit item to grid
    removedItem.width = Math.min(
      removedItem.width ?? defaults.COMPONENT_WIDTH,
      defaults.COLUMN_COUNT - (targetItem.x ?? 0),
    );
    removedItem.height = removedItem.height ?? defaults.COMPONENT_HEIGHT;
    removedItem.y = targetItem.y ?? 0;

    const targetIndex = targetRow.items.indexOf(targetItem);
    targetRow.items.splice(targetIndex, 0, removedItem);

    let currentX = 0;
    targetRow.items.forEach((item) => {
      item.x = currentX;
      currentX += item.width ?? defaults.COMPONENT_WIDTH;
    });

    targetRow.height = Math.max(targetRow.height, removedItem.height);
  }

  private handleRightDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
    rows: RowGroup[],
  ): void {
    const targetRow = rows.find((row) => row.y === targetItem.y);
    if (!targetRow) return;

    // Fit item to grid
    const remainingSpace =
      defaults.COLUMN_COUNT -
      ((targetItem.x ?? 0) + (targetItem.width ?? defaults.COMPONENT_WIDTH));
    removedItem.width = Math.min(
      removedItem.width ?? defaults.COMPONENT_WIDTH,
      remainingSpace,
    );
    removedItem.height = removedItem.height ?? defaults.COMPONENT_HEIGHT;
    removedItem.y = targetItem.y ?? 0;
    removedItem.x =
      (targetItem.x ?? 0) + (targetItem.width ?? defaults.COMPONENT_WIDTH);

    const targetIndex = targetRow.items.indexOf(targetItem);
    targetRow.items.splice(targetIndex + 1, 0, removedItem);
    targetRow.height = Math.max(targetRow.height, removedItem.height);
  }

  private handleBottomDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
    rows: RowGroup[],
  ): void {
    const targetY = targetItem.y ?? 0;
    const targetHeight = targetItem.height ?? defaults.COMPONENT_HEIGHT;
    const newY = targetY + targetHeight;

    // For bottom drops, use full width
    removedItem.y = newY;
    removedItem.x = 0;
    removedItem.width = defaults.COLUMN_COUNT;
    removedItem.height = removedItem.height ?? defaults.COMPONENT_HEIGHT;

    rows.push({
      y: newY,
      height: removedItem.height,
      items: [removedItem],
    });
  }

  private handleTopDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
    rows: RowGroup[],
  ): void {
    const targetRow = rows.find((row) => row.y === targetItem.y);
    if (!targetRow) return;

    // For top drops, use full width
    removedItem.x = 0;
    removedItem.y = targetItem.y ?? 0;
    removedItem.width = defaults.COLUMN_COUNT;
    removedItem.height = removedItem.height ?? defaults.COMPONENT_HEIGHT;

    targetRow.items.splice(0, 0, removedItem);
    targetRow.height = Math.max(targetRow.height, removedItem.height);
  }

  private preventCollisions(items: V1CanvasItem[]): V1CanvasItem[] {
    const rowGroups = Grid.groupItemsByRow(items);

    rowGroups.forEach((row) => {
      let currentX = 0;
      row.items.sort((a, b) => (a.x ?? 0) - (b.x ?? 0));

      row.items.forEach((item) => {
        item.x = currentX;
        currentX += item.width ?? defaults.COMPONENT_WIDTH;
        if (currentX > defaults.COLUMN_COUNT) {
          currentX = defaults.COLUMN_COUNT;
        }
      });
    });

    return items;
  }
}

export const isValidItem = Grid.isValidItem;
export const groupItemsByRow = Grid.groupItemsByRow;
