import { ContextPickerUIState } from "@rilldata/web-common/features/chat/core/context/picker/ui-state.ts";
import { derived, type Readable } from "svelte/store";
import type { InlineContextPickerParentOption } from "@rilldata/web-common/features/chat/core/context/picker/types.ts";
import { getPickerOptions } from "@rilldata/web-common/features/chat/core/context/picker/data";

/**
 * Creates a store that contains a list of options that match the search text.
 * 1. Directly calls {@link getPickerOptions} to get the initial list of options.
 * 2. Bubbles up the recently used and active top level options to the top of the list.
 * 3. Removes any top level options that don't match the search text including any 2nd level options within it.
 */
export function getFilteredPickerOptions(
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

      let recentlyUsed: InlineContextPickerParentOption | null = null;
      let currentlyActive: InlineContextPickerParentOption | null = null;

      const filteredOptions = options
        .map((option) => {
          const children =
            option.children?.filter((c) => filterFunction(c.label!, c.value)) ??
            [];

          const parentMatches = filterFunction(
            option.context.label ?? "",
            option.context.value,
          );

          if (!parentMatches && children.length === 0) return null;

          const filteredOption = {
            ...option,
            children,
          } satisfies InlineContextPickerParentOption;

          if (!recentlyUsed && option.recentlyUsed) {
            recentlyUsed = filteredOption;
          }
          if (!currentlyActive && option.currentlyActive) {
            currentlyActive = filteredOption;
          }
          if (option.recentlyUsed || option.currentlyActive) return null; // these are added explicitly

          return filteredOption;
        })
        .filter(Boolean) as InlineContextPickerParentOption[];

      if (recentlyUsed === currentlyActive) currentlyActive = null;

      const allOptions = [
        recentlyUsed,
        currentlyActive,
        ...filteredOptions,
      ].filter(Boolean) as InlineContextPickerParentOption[];

      return allOptions;
    },
  );
}
