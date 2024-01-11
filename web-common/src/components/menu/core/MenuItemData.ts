export interface MenuItemGroupData {
  name: string;
  showDivider: boolean;
  items: MenuItemData[];
}

export interface MenuItemData {
  name: string;
  label: string;
  selected: boolean;
  index: number;
}
