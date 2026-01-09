import { ContextPickerUIState } from "@rilldata/web-common/features/chat/core/context/picker/ui-state.ts";
import { derived, type Readable } from "svelte/store";
import type { PickerItem } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import { getPickerOptions } from "@rilldata/web-common/features/chat/core/context/picker/data";

/**
 * Creates a store that contains a list of options that match the search text.
 * 1. Directly calls {@link getPickerOptions} to get the initial list of options.
 * 2. Filters the list based on the search text.
 * 3. If any child options are present, retains the parent option as well.
 */
export function getFilteredPickerItems(
  uiState: ContextPickerUIState,
  searchTextStore: Readable<string>,
) {
  return derived(
    [getPickerOptions(uiState), searchTextStore],
    ([options, searchText]) => {
      const filterFunction = (label: string, value: string) =>
        searchText.length === 0 ||
        label.toLowerCase().includes(searchText.toLowerCase()) ||
        value.toLowerCase().includes(searchText.toLowerCase());

      const parentsToKeep = new Set<string>();
      const filteredOptions = new Array<PickerItem | null>(options.length).fill(
        null,
      );

      for (let i = options.length - 1; i >= 0; i--) {
        const option = options[i];
        const matches = filterFunction(
          option.context.label ?? "",
          option.context.value,
        );
        if (!matches && !parentsToKeep.has(option.id)) continue;

        if (option.parentId) parentsToKeep.add(option.parentId);
        filteredOptions[i] = option;
      }

      return filteredOptions.filter(Boolean) as PickerItem[];
    },
  );
}
