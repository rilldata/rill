import { DimensionFilterMode } from "./constants";

export function getSearchPlaceholder(mode: DimensionFilterMode): string {
  switch (mode) {
    case DimensionFilterMode.Select:
      return "Enter search term or paste list of values";
    case DimensionFilterMode.InList:
      return "Paste a list separated by commas or \\n";
    case DimensionFilterMode.Contains:
      return "Enter a search term";
    default:
      return "Enter search term";
  }
}

export function getEffectiveSelectedValues(
  mode: DimensionFilterMode,
  selectedValuesProxy: string[],
  correctedSearchResults: string[],
  selectedValues: string[],
): string[] {
  switch (mode) {
    case DimensionFilterMode.Select:
      return selectedValuesProxy;
    case DimensionFilterMode.InList:
      return correctedSearchResults ?? [];
    case DimensionFilterMode.Contains:
    default:
      return selectedValues;
  }
}

export function shouldDisableApplyButton(
  mode: DimensionFilterMode,
  enableSearchCountQuery: boolean,
  inListTooLong: boolean,
): boolean {
  switch (mode) {
    case DimensionFilterMode.Select:
      return false; // Never disable Apply for Select mode
    case DimensionFilterMode.InList:
    case DimensionFilterMode.Contains:
    default:
      return !enableSearchCountQuery || inListTooLong;
  }
}

export function getItemLists(
  mode: DimensionFilterMode,
  correctedSearchResults: string[],
  selectedValues: string[],
  curSearchText: string,
): { checkedItems: string[]; uncheckedItems: string[] } {
  if (mode === DimensionFilterMode.Select && correctedSearchResults) {
    // While searching in Select mode, include selected items in the unified list
    // so that matching selections are visible. When not searching, keep the
    // split view: checked items first, then unchecked.
    const isSearching = Boolean(curSearchText);
    return {
      checkedItems: isSearching
        ? []
        : correctedSearchResults.filter((item) =>
            selectedValues.includes(item),
          ),
      uncheckedItems: correctedSearchResults.filter((item) =>
        isSearching ? true : !selectedValues.includes(item),
      ),
    };
  }

  return {
    checkedItems: [],
    uncheckedItems: correctedSearchResults ?? [],
  };
}
