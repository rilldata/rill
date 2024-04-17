import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getPaddingFromPath } from "@rilldata/web-common/features/file-explorer/nav-tree-spacing";
import { splitFolderAndName } from "@rilldata/web-common/features/sources/extract-file-name";
import { get, writable } from "svelte/store";

export type NavDragData = {
  id: string;
  filePath: string;
  fileName?: string;
  isDir: boolean;
  kind?: ResourceKind | undefined;
};

export class NavEntryDragDropStore {
  private static readonly MIN_DRAG_DISTANCE = 9;
  public readonly dragData = writable<null | NavDragData>(null);
  public initialPosition = { left: 0, top: 0 };
  public readonly position = writable({ left: 0, top: 0 });
  public offset = { x: 0, y: 0 };

  private newDragData: NavDragData | null;

  public onMouseDown(e: MouseEvent, dragData: NavDragData) {
    e.preventDefault();
    e.stopPropagation();

    const offsets = this.getOffsets(e, dragData);
    if (!offsets) return;
    const { left, top, x, y } = offsets;

    this.initialPosition = { left, top };

    this.offset = { x, y };

    const [, fileName] = splitFolderAndName(dragData.filePath);
    this.newDragData = {
      ...dragData,
      fileName,
    };
  }

  public async onMouseUp(
    e: MouseEvent,
    dragData: NavDragData | null,
    dropSuccess: (
      fromDragData: NavDragData,
      toDragData: NavDragData,
    ) => Promise<void>,
  ) {
    e.preventDefault();
    e.stopPropagation();

    const curDragData = get(this.dragData);

    if (curDragData && dragData && curDragData.filePath !== dragData.filePath) {
      await dropSuccess(curDragData, dragData);
    }

    this.newDragData = null;
    this.dragData.set(null);
  }

  public onMouseMove(e: MouseEvent) {
    if (!this.newDragData) return;
    const left = e.clientX - this.offset.x;
    const top = e.clientY - this.offset.y;
    this.position.set({ left, top });

    if (get(this.dragData)) return;
    const dist = Math.sqrt(
      Math.pow(left - this.initialPosition.left, 2) +
        Math.pow(top - this.initialPosition.top, 2),
    );
    if (dist < NavEntryDragDropStore.MIN_DRAG_DISTANCE) return;
    this.dragData.set(this.newDragData);
  }

  private getOffsets(e: MouseEvent, dragData: NavDragData) {
    const dragItem = document.getElementById(dragData.id);
    if (!dragItem) return;

    const { left, top } = dragItem.getBoundingClientRect();
    // 14 is the temporary offset for icon. we should add the icon eventually
    const effectiveLeft = left + getPaddingFromPath(dragData.filePath);

    return {
      left: effectiveLeft,
      top,
      x: e.clientX - effectiveLeft,
      y: e.clientY - top,
    };
  }
}

export const navEntryDragDropStore = new NavEntryDragDropStore();
