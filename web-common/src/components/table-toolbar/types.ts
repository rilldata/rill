import type {
  ArrayRuneStore,
  RuneStore,
} from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

export type SortDirection = "newest" | "oldest";

export type ViewMode = "list" | "grid";

export interface FilterOption {
  value: string;
  label: string;
}

export type FilterGroup = {
  label: string;
  key: string;
  options: FilterOption[];
} & (
  | {
      selectedStore: RuneStore<string>;
      selected: string;
      defaultValue: string;
      multiSelect?: false;
    }
  | {
      selectedStore: ArrayRuneStore<string>;
      selected: string[];
      defaultValue: string[];
      multiSelect: true;
    }
);
