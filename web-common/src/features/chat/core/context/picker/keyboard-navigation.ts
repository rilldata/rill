import type { PickerItem } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import { get, writable } from "svelte/store";
import type { ContextPickerUIState } from "@rilldata/web-common/features/chat/core/context/picker/ui-state.ts";
import type { InlineContext } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

export class KeyboardNavigationManager {
  public readonly focusedItemStore = writable<PickerItem | null>(null);

  private pickerItems: PickerItem[] = [];
  private itemIdToIndex = new Map<string, number>();
  private focusedIndex = -1;

  public constructor(
    private readonly uiState: ContextPickerUIState,
    private onSelect: (ctx: InlineContext) => void,
  ) {}

  public setPickerItems(
    pickerItems: PickerItem[],
    expandedParents: Record<string, boolean>,
    selectedItemId: string | null,
  ) {
    const expandedItems = pickerItems.filter(
      (item) => !item.parentId || expandedParents[item.parentId],
    );
    this.focusedIndex = this.getUpdatedIndex(expandedItems, selectedItemId);

    this.pickerItems = expandedItems;
    this.itemIdToIndex = new Map(
      pickerItems.map((item, index) => [item.id, index]),
    );

    this.updateFocusedItem();
  }

  public focusNext() {
    this.focusedIndex = (this.focusedIndex + 1) % this.pickerItems.length;
    this.updateFocusedItem();
  }

  public focusPrevious() {
    this.focusedIndex =
      (this.focusedIndex - 1 + this.pickerItems.length) %
      this.pickerItems.length;
    this.updateFocusedItem();
  }

  public collapseToClosestParent() {
    const focusedItem = get(this.focusedItemStore);
    if (!focusedItem) return;

    // Collapse the focused item if it is expanded.
    if (this.uiState.isExpanded(focusedItem.id)) {
      this.uiState.collapse(focusedItem.id);
    }

    // Else try to find the parent and collapse it.
    const parentIndex = this.itemIdToIndex.get(focusedItem.parentId ?? "");
    if (parentIndex == undefined || parentIndex === -1) return;

    this.uiState.collapse(focusedItem.parentId ?? "");
    // Focuse the parent after collapsing.
    this.focusedIndex = parentIndex;
    this.updateFocusedItem();
  }

  public openCurrentParentOption() {
    const focusedItem = get(this.focusedItemStore);
    // Ensure the focused item has children before expanding it.
    // While it doesn't hurt to mark leaf nodes as expanded this is good for consistency.
    if (!focusedItem?.hasChildren) return;
    this.uiState.expand(focusedItem.id);
  }

  public handleKeyDown(event: KeyboardEvent) {
    switch (event.key) {
      case "ArrowUp":
        this.focusPrevious();
        break;
      case "ArrowDown":
        this.focusNext();
        break;
      case "ArrowLeft":
        this.collapseToClosestParent();
        break;
      case "ArrowRight":
        this.openCurrentParentOption();
        break;
      case "Enter": {
        const focusedItem = get(this.focusedItemStore);
        if (focusedItem) this.onSelect(focusedItem.context);
        event.preventDefault();
        event.stopPropagation();
        event.stopImmediatePropagation();
        break;
      }
    }
  }

  public ensureIsFocused = (node: Node, item: PickerItem) => {
    const onMouseEnter = () => this.focusItem(item);
    node.addEventListener("mousemove", onMouseEnter);
    return {
      destroy() {
        node.removeEventListener("mousemove", onMouseEnter);
      },
    };
  };

  private focusItem(item: PickerItem) {
    if (!this.itemIdToIndex.has(item.id)) return;
    this.focusedIndex = this.itemIdToIndex.get(item.id)!;
    this.updateFocusedItem();
  }

  private updateFocusedItem() {
    this.focusedItemStore.set(this.pickerItems[this.focusedIndex] ?? null);
  }

  private getUpdatedIndex(
    newItems: PickerItem[],
    selectedItemId: string | null,
  ) {
    const focusedItem = get(this.focusedItemStore);
    if (focusedItem) {
      // Prefer the item itself if it is available.
      const focusedIndex = newItems.findIndex((i) => i.id === focusedItem.id);
      if (focusedIndex !== -1) return focusedIndex;

      // Otherwise prefer the parent if it is available.
      const parentIndex = newItems.findIndex(
        (i) => i.id === focusedItem.parentId,
      );
      if (parentIndex !== -1) return parentIndex;
    }

    // Then prefer the selected item if available.
    if (selectedItemId) {
      const selectedIndex = newItems.findIndex((i) => i.id === selectedItemId);
      if (selectedIndex !== -1) return selectedIndex;
    }

    // Finally, prefer to select the child of the first parent item.
    const firstItemIsParent =
      this.pickerItems.length > 0 && this.pickerItems[0].hasChildren;
    if (!firstItemIsParent) return 0;

    const secondItemIsChild =
      this.pickerItems.length > 1 && !this.pickerItems[1].hasChildren;
    return secondItemIsChild ? 1 : 0;
  }
}
