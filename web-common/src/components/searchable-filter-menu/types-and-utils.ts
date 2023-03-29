export type SelectableItem = {
  label: string;
  selected: boolean;
  visibleInMenu: boolean;
};

// export type SelectableItem = Required<SelectableItemObject>;

export const getNumSelectedNotShown = (items: SelectableItem[]): number =>
  items?.filter((x) => x.selected && !x.visibleInMenu)?.length || 0;

export const setItemsVisibleBySearchString = (
  items: SelectableItem[],
  searchText: string
): SelectableItem[] => {
  return items?.map((x) => ({
    ...x,
    visibleInMenu: x.label.includes(searchText.trim()),
  }));
};
