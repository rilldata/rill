export type SelectableItem = {
  label: string;
  selected: boolean;
};

export const getNumSelectedNotShown = (
  items: SelectableItem[],
  visibleInSearch: boolean[]
): number =>
  items?.filter((x, i) => x.selected && !visibleInSearch[i])?.length || 0;

export const setItemsVisibleBySearchString = (
  items: string[],
  searchText: string
): boolean[] => {
  return items?.map((x) => x.includes(searchText.trim()));
};
