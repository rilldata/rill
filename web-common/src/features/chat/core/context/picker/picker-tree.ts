import { type InlineContext } from "@rilldata/web-common/features/chat/core/context/inline-context.ts";

export type PickerItem = {
  id: string;
  context: InlineContext;
  parentId?: string; // undefined for top-level items
  recentlyUsed?: boolean;
  currentlyActive?: boolean;
  childrenLoading?: boolean;
  hasChildren?: boolean;
};

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
  let prevItem: PickerItem | null = null;
  rootNodes.forEach((rootNode) => {
    if (rootNode.item.currentlyActive || rootNode.item.recentlyUsed) {
      boundaryIndices.add(rootNode.item.id);
      return;
    }

    if (prevItem?.context?.type === rootNode.item.context.type) {
      prevItem = rootNode.item;
      return;
    }

    const isFirstBoundary = prevItem === null && boundaryIndices.size === 0;
    if (!isFirstBoundary) boundaryIndices.add(rootNode.item.id);
    prevItem = rootNode.item;
  });

  return {
    rootNodes,
    boundaryIndices,
  };
}
