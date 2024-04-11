import { splitFolderAndName } from "@rilldata/web-common/features/entity-management/file-selectors";
import { ResourceKind } from "@rilldata/web-common/features/entity-management/resource-selectors";
import { getPaddingFromPath } from "@rilldata/web-common/features/file-explorer/nav-tree-spacing";
import { createDebouncer } from "@rilldata/web-common/lib/create-debouncer";
import { get, writable } from "svelte/store";

export type NavDragData = {
  id: string;
  filePath: string;
  fileName?: string;
  isDir: boolean;
  kind?: ResourceKind | undefined;
};

export class NavEntryDragDropStore {
  public readonly navDragging = writable<null | NavDragData>(null);
  public readonly offset = writable({ x: 0, y: 0 });
  public readonly dragStart = writable({ left: 0, top: 0 });

  private newDragData: NavDragData | null;
  private readonly setDraggingDebouncer = createDebouncer();

  public onMouseDown(e: MouseEvent, dragData: NavDragData) {
    e.preventDefault();
    e.stopPropagation();

    const dragItem = document.getElementById(dragData.id);
    if (!dragItem) return;

    const { left, top } = dragItem.getBoundingClientRect();

    this.dragStart.set({
      left: left + getPaddingFromPath(dragData.filePath),
      top,
    });

    this.offset.set({
      x: e.clientX - left - getPaddingFromPath(dragData.filePath),
      y: e.clientY - top,
    });

    const [, fileName] = splitFolderAndName(dragData.filePath);
    this.newDragData = {
      ...dragData,
      fileName,
    };
  }

  public onMouseUp(
    e: MouseEvent,
    dragData: NavDragData | null,
    dropSuccess: (fromDragData: NavDragData, toDragData: NavDragData) => void,
  ) {
    e.preventDefault();
    e.stopPropagation();

    const curDragData = get(this.navDragging);

    if (curDragData && dragData && curDragData.filePath !== dragData.filePath) {
      dropSuccess(curDragData, dragData);
    }

    this.newDragData = null;
    this.navDragging.set(null);
  }

  public onMouseMove() {
    if (!this.newDragData) return;
    this.navDragging.set(this.newDragData);
    this.setDraggingDebouncer(() => {
      if (this.newDragData) this.navDragging.set(this.newDragData);
    }, 200);
  }
}

export const navEntryDragDropStore = new NavEntryDragDropStore();
