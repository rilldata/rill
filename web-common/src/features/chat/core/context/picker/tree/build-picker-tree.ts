import type { PickerItem } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import { InlineContextType } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

export type PickerTreeNode = {
  item: PickerItem;
  children: PickerTreeNode[];
};
export type PickerTree = {
  rootNodes: PickerTreeNode[];
  // Used to show a border around the boundaries.
  boundaryIndices: Set<string>;
};

export function buildPickerTree(pickerItems: PickerItem[]): PickerTree {
  const nodesById = new Map<string, PickerTreeNode>();
  const rootNodes: PickerTreeNode[] = [];

  pickerItems.forEach((item) => {
    const node = { item, children: [] };
    nodesById.set(item.id, node);

    const parentNode = nodesById.get(item.parentId ?? "");
    if (parentNode) {
      parentNode.children.push(node);
    } else {
      rootNodes.push(node);
    }
  });

  // Calculate boundary indices to show border around.
  const boundaryIndices = new Set<string>();
  let prevType: InlineContextType | null = null;
  pickerItems.forEach((item) => {
    if (item.currentlyActive || item.recentlyUsed) {
      boundaryIndices.add(item.id);
      return;
    }

    if (!prevType || prevType !== item.context.type) return;

    prevType = item.context.type;
    boundaryIndices.add(item.id);
  });

  return {
    rootNodes,
    boundaryIndices,
  };
}
