import type {
  ArrayRuneStore,
  RuneStore,
} from "@rilldata/web-common/lib/store-utils/types.svelte.ts";

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
      defaultValue: string;
      multiSelect?: false;
    }
  | {
      selectedStore: ArrayRuneStore<string>;
      defaultValue: string[];
      multiSelect: true;
    }
);

export type SortOption = {
  value: string;
  label: string;
};
