import type {
  MenuItemData,
  MenuItemGroupData,
} from "@rilldata/web-common/components/menu/core/MenuItemData";
import { matchSorter } from "match-sorter";

export interface SearchableFilterSelectableGroup {
  name: string;
  items: SearchableFilterSelectableItem[];
}

export interface SearchableFilterSelectableItem {
  name: string;
  label: string;
}

export function getMenuGroups(
  groups: SearchableFilterSelectableGroup[],
  selected: boolean[][],
  searchText: string
) {
  return groups.map(
    (g, i) =>
      <MenuItemGroupData>{
        name: g.name,
        showDivider: i > 0,
        items: getMenuItems(g.items, selected[i] ?? [], searchText),
      }
  );
}

function getMenuItems(
  items: SearchableFilterSelectableItem[],
  selected: boolean[],
  searchText: string
) {
  const menuItems = items.map(
    (item, i) =>
      <MenuItemData>{
        name: item.name,
        label: item.label,
        selected: selected[i],
        index: i,
      }
  );
  if (!searchText) return menuItems;
  return matchSorter(menuItems, searchText, { keys: ["label"] });
}
