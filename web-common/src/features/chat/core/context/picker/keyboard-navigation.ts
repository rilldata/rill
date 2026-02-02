import { get, writable } from "svelte/store";
import type { ContextPickerUIState } from "@rilldata/web-common/features/chat/core/context/picker/ui-state.ts";
import type { InlineContext } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";
import type { PickerItem } from "@rilldata/web-common/features/chat/core/context/picker/picker-tree.ts";
import { EventEmitter } from "@rilldata/web-common/lib/event-emitter.ts";

type KeyboardNavigationManagerEvents = {
  select: InlineContext;
  "item-focused": PickerItem | null;
};

export class KeyboardNavigationManager {
  public readonly focusedItemStore = writable<PickerItem | null>(null);

  private pickerItems: PickerItem[] = [];
  private itemIdToIndex = new Map<string, number>();
  private focusedIndex = -1;

  private events = new EventEmitter<KeyboardNavigationManagerEvents>();
  public readonly on = this.events.on.bind(
    this.events,
  ) as typeof this.events.on;

  public constructor(private readonly uiState: ContextPickerUIState) {}

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
    // Focus the parent after collapsing.
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
        if (focusedItem) this.events.emit("select", focusedItem.context);
        event.preventDefault();
        event.stopPropagation();
        event.stopImmediatePropagation();
        break;
      }
    }
  }

  /**
   * Enhances a picker node with keyboard navigation support. Focus can be changed using keyboard shortcuts, up and down arrows.
   * 1. Hovering should update the focused item to support using keyboard after hovering.
   * 2. When a focused item changes, scroll it into view.
   *
   * Note that we are using mouse move instead of enter/exit.
   * This prevents focus changing when the layout changes, but the mouse doesn't move.
   *
   * For focusing items, using "focused" independently can have a feedback loop when focused using the mouse.
   */
  public enhancePickerNode = (node: HTMLElement, item: PickerItem) => {
    const onMouseEnter = () => this.focusItem(item);
    node.addEventListener("mousemove", onMouseEnter);
    const itemFocusedUnsub = this.events.on("item-focused", (focusedItem) => {
      if (focusedItem?.id === item.id) {
        node.scrollIntoView({ block: "nearest" });
      }
    });

    return {
      destroy() {
        node.removeEventListener("mousemove", onMouseEnter);
        itemFocusedUnsub();
      },
    };
  };

  private focusItem(item: PickerItem) {
    if (!this.itemIdToIndex.has(item.id)) return;
    this.focusedIndex = this.itemIdToIndex.get(item.id)!;
    this.updateFocusedItem();
  }

  private updateFocusedItem() {
    const newFocusedItem = this.pickerItems[this.focusedIndex] ?? null;
    this.focusedItemStore.set(newFocusedItem);
    this.events.emit("item-focused", newFocusedItem);
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
