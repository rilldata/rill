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

  static leftAlignRow(row: RowGroup) {
    let currentX = 0;

    row.items
      .sort((a, b) => (a.x ?? 0) - (b.x ?? 0))
      .forEach((item) => {
        item.x = currentX;
        currentX += item.width ?? defaults.COMPONENT_WIDTH;
      });
  }

  private validateItemPositions(items: V1CanvasItem[]): void {
    // First group items by row
    const rows = Grid.groupItemsByRow(items);

    // Process each row
    rows.forEach((row) => {
      Grid.leftAlignRow(row);
    });

    // Validate x positions are within bounds
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
    // Get the resize handles
    const rowHandle = document.querySelector(".row-resize-handle");
    const colHandle = document.querySelector(".col-resize-handle");

    if (rowHandle && this.isOverElement(mouseX, mouseY, rowHandle)) {
      return "row";
    }

    if (colHandle && this.isOverElement(mouseX, mouseY, colHandle)) {
      return "col";
    }

    // Default to col if not over any handle
    return "col";
  }

  private isOverElement(
    mouseX: number,
    mouseY: number,
    element: Element,
  ): boolean {
    const rect = element.getBoundingClientRect();
    return (
      mouseX >= rect.left &&
      mouseX <= rect.right &&
      mouseY >= rect.top &&
      mouseY <= rect.bottom
    );
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

    // Create a deep copy of the dragged item
    const [draggedItemFull] = newItems.splice(dragIndex, 1);
    const removedItem = { ...draggedItemFull };

    switch (position) {
      case "col": {
        console.log("[Grid] Dropping into column");
        this.handleColumnDrop(removedItem, targetItem, rows);
        // Find the correct insert index based on x position
        const targetRow = rows.find((row) => row.y === targetItem.y);
        if (targetRow) {
          const sortedRowItems = targetRow.items.sort(
            (a, b) => (a.x ?? 0) - (b.x ?? 0),
          );
          const insertPosition = sortedRowItems.findIndex(
            (item) => (item.x ?? 0) > (removedItem.x ?? 0),
          );
          insertIndex =
            insertPosition === -1
              ? this.items.length
              : this.items.indexOf(sortedRowItems[insertPosition]);
        }
        break;
      }
      case "row": {
        console.log("[Grid] Dropping into row");
        this.handleRowDrop(removedItem, targetItem, rows);
        insertIndex = this.items.findIndex((item) => item.y === targetItem.y);
        break;
      }
    }

    // Reinsert the item
    newItems.splice(insertIndex, 0, removedItem);

    // Prevent collisions and validate
    newItems = this.preventCollisions(newItems);
    this.validateItemPositions(newItems);

    return { items: newItems, insertIndex };
  }

  private handleColumnDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
    rows: ReturnType<typeof Grid.groupItemsByRow>,
  ) {
    const targetRow = rows.find((row) => row.y === targetItem.y);
    if (!targetRow) return;

    // Sort items in the row by x position
    const sortedRowItems = [...targetRow.items].sort(
      (a, b) => (a.x ?? 0) - (b.x ?? 0),
    );

    // Find the target item's position in the sorted row
    const targetIndex = sortedRowItems.indexOf(targetItem);

    // Calculate the new x position for the dropped item
    const newX = targetItem.x ?? 0;

    // Shift all items from the target position to the right
    for (let i = targetIndex; i < sortedRowItems.length; i++) {
      const item = sortedRowItems[i];
      if (!item) continue;

      // Move each item to the right by the width of the dropped item
      item.x = (item.x ?? 0) + (removedItem.width ?? defaults.COMPONENT_WIDTH);
    }

    // Place the dropped item at the target position
    removedItem.y = targetItem.y ?? 0;
    removedItem.x = newX;
    removedItem.height = removedItem.height ?? defaults.COMPONENT_HEIGHT;
    removedItem.width = removedItem.width ?? defaults.COMPONENT_WIDTH;

    // Update the row's height if necessary
    targetRow.height = Math.max(targetRow.height, removedItem.height);
  }

  private handleRowDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
    rows: ReturnType<typeof Grid.groupItemsByRow>,
  ) {
    const targetRow = rows.find((row) => row.y === targetItem.y);
    if (targetRow) {
      removedItem.x = 0;
      removedItem.y = targetItem.y ?? 0;
      removedItem.height = removedItem.height ?? defaults.COMPONENT_HEIGHT;

      // Insert at beginning of row using splice
      targetRow.items.splice(0, 0, removedItem);
      targetRow.height = Math.max(targetRow.height, removedItem.height);
    }
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
