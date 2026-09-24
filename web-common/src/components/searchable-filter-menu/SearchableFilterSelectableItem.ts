export interface SearchableFilterSelectableGroup {
  name: string;
  label?: string;
  items: SearchableFilterSelectableItem[];
}

export interface SearchableFilterSelectableItem {
  name: string;
  label: string;
  // Shown as a native tooltip on the menu row.
  description?: string;
  // Marks ephemeral measures with an fx icon.
  ephemeral?: boolean;
}
