import type { V1CanvasItem } from "@rilldata/web-common/runtime-client";
import * as defaults from "./constants";
import type { DropPosition, Vector, RowGroup, GridItem } from "./types";

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
    // Define zones - make them smaller for more precise targeting
    const bottomZone = targetRect.bottom - targetRect.height * 0.2;
    const topZone = targetRect.top + targetRect.height * 0.2;
    const leftZone = targetRect.left + targetRect.width * 0.2;
    const rightZone = targetRect.right - targetRect.width * 0.2;

    // Check vertical zones first
    if (mouseY > bottomZone) {
      return "bottom";
    } else if (mouseY < topZone) {
      // If near the top edge, determine if it should be "top" or "row"
      if (mouseX < leftZone) {
        return "row"; // Start of row when near top-left
      }
      return "top";
    }

    // If in the middle zone, determine left/right
    if (mouseX < leftZone) {
      return "left";
    } else if (mouseX > rightZone) {
      return "right";
    }

    // Default to closest edge if in center
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

    // Create a deep copy of the dragged item
    const [draggedItemFull] = newItems.splice(dragIndex, 1);
    const removedItem = { ...draggedItemFull };

    switch (position) {
      case "top": {
        console.log("[Grid] Dropping top");
        this.handleTopDrop(removedItem, targetItem, rows);
        insertIndex = this.items.indexOf(targetItem);
        break;
      }
      case "bottom": {
        console.log("[Grid] Dropping bottom");
        this.handleBottomDrop(removedItem, targetItem, rows);
        insertIndex = this.items.indexOf(targetItem) + 1;
        break;
      }
      case "right": {
        console.log("[Grid] Dropping right");
        this.handleRightDrop(removedItem, targetItem, rows);
        insertIndex = this.items.indexOf(targetItem) + 1;
        break;
      }
      case "left": {
        console.log("[Grid] Dropping left");
        this.handleLeftDrop(removedItem, targetItem, rows);
        insertIndex = this.items.indexOf(targetItem);
        break;
      }
      case "row": {
        console.log("[Grid] Dropping row");
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

  private handleBottomDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
    rows: ReturnType<typeof Grid.groupItemsByRow>,
  ) {
    const targetY = targetItem.y ?? 0;
    const targetHeight = targetItem.height ?? defaults.COMPONENT_HEIGHT;
    const newY = targetY + targetHeight;

    removedItem.y = newY;
    removedItem.x = 0;
    removedItem.width = defaults.COLUMN_COUNT;
    removedItem.height = removedItem.height ?? defaults.COMPONENT_HEIGHT;

    // Create new row and add item at index 0
    const newRow = {
      y: newY,
      height: removedItem.height,
      items: [],
    };
    newRow.items.splice(0, 0, removedItem);
    rows.push(newRow);
  }

  private handleRightDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
    rows: ReturnType<typeof Grid.groupItemsByRow>,
  ) {
    const targetRow = rows.find((row) => row.y === targetItem.y);
    if (targetRow) {
      // Ensure removedItem has defined dimensions
      removedItem.width = removedItem.width ?? defaults.COMPONENT_WIDTH;
      removedItem.height = removedItem.height ?? defaults.COMPONENT_HEIGHT;

      // Align y position and set x to immediately after targetItem
      removedItem.y = targetItem.y ?? 0;
      removedItem.x =
        (targetItem.x ?? 0) + (targetItem.width ?? defaults.COMPONENT_WIDTH);

      // Insert removedItem after targetItem
      const targetIndex = targetRow.items.indexOf(targetItem);
      targetRow.items.splice(targetIndex + 1, 0, removedItem);

      // Update row height if necessary
      targetRow.height = Math.max(targetRow.height, removedItem.height);
    }
  }

  private handleLeftDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
    rows: ReturnType<typeof Grid.groupItemsByRow>,
  ) {
    const targetRow = rows.find((row) => row.y === targetItem.y);
    if (targetRow) {
      // Set the y position and height of the removed item
      removedItem.y = targetItem.y ?? 0;
      removedItem.height = removedItem.height ?? defaults.COMPONENT_HEIGHT;

      // Insert the removed item before the target item
      const targetIndex = targetRow.items.indexOf(targetItem);
      targetRow.items.splice(targetIndex, 0, removedItem);

      // Adjust x positions to prevent overlaps
      let currentX = 0;
      targetRow.items.forEach((item) => {
        item.x = currentX;
        currentX += item.width ?? defaults.COMPONENT_WIDTH;
      });

      // Update the row's height if necessary
      targetRow.height = Math.max(targetRow.height, removedItem.height);
    }
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

  private handleTopDrop(
    removedItem: V1CanvasItem,
    targetItem: V1CanvasItem,
    rows: ReturnType<typeof Grid.groupItemsByRow>,
  ) {
    // Move all items in the target row and below down
    const targetY = targetItem.y ?? 0;
    const itemHeight = removedItem.height ?? defaults.COMPONENT_HEIGHT;

    // Shift down items at or below the target
    this.items.forEach((item) => {
      if ((item.y ?? 0) >= targetY) {
        item.y = (item.y ?? 0) + itemHeight;
      }
    });

    // Position the removed item
    removedItem.y = targetY;
    removedItem.x = targetItem.x ?? 0;
    removedItem.width = removedItem.width ?? defaults.COMPONENT_WIDTH;
    removedItem.height = itemHeight;
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

  // static convertToGridItems(yamlItems: any[]): GridItem[] {
  //   return yamlItems.map((item) => ({
  //     position: [item.get("x"), item.get("y")],
  //     size: [item.get("width"), item.get("height")],
  //     node: item,
  //   }));
  // }

  // static sortItemsByPosition(items: GridItem[]): GridItem[] {
  //   return items.sort((a, b) => {
  //     // Sort by Y first, then X for items in the same row
  //     if (a.position[1] === b.position[1]) {
  //       return a.position[0] - b.position[0];
  //     }
  //     return a.position[1] - b.position[1];
  //   });
  // }

  // static compactGrid(items: GridItem[]) {
  //   let currentY = 0;
  //   let lastRowHeight = 0;
  //   let lastY = -1;
  //   let currentRowItems: GridItem[] = [];

  //   // Process all items
  //   items.forEach((item) => {
  //     if (item.position[1] !== lastY) {
  //       // When we hit a new row, process the previous row
  //       if (currentRowItems.length > 0) {
  //         // Sort by X position and compact
  //         currentRowItems.sort((a, b) => a.position[0] - b.position[0]);
  //         let currentX = 0;
  //         currentRowItems.forEach((rowItem) => {
  //           rowItem.node.set("x", currentX);
  //           currentX += rowItem.size[0];
  //         });
  //       }

  //       // Start new row
  //       currentY += lastRowHeight;
  //       lastRowHeight = item.size[1];
  //       lastY = item.position[1];
  //       currentRowItems = [item];
  //     } else {
  //       // Same row - update max height if needed
  //       lastRowHeight = Math.max(lastRowHeight, item.size[1]);
  //       currentRowItems.push(item);
  //     }

  //     // Update Y position
  //     item.node.set("y", currentY);
  //   });

  //   // Process the last row
  //   if (currentRowItems.length > 0) {
  //     currentRowItems.sort((a, b) => a.position[0] - b.position[0]);
  //     let currentX = 0;
  //     currentRowItems.forEach((rowItem) => {
  //       rowItem.node.set("x", currentX);
  //       currentX += rowItem.size[0];
  //     });
  //   }
  // }
}

export const isValidItem = Grid.isValidItem;
export const groupItemsByRow = Grid.groupItemsByRow;
