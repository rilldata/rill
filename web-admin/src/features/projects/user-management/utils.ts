import { OrgUserRoles } from "@rilldata/web-common/features/users/roles.ts";

export interface SearchResult {
  identifier: string;
  type: "user" | "group";
  orgRoleName?: string;
  [key: string]: any;
}

export interface CategorizedResults {
  groups: SearchResult[];
  members: SearchResult[];
  guests: SearchResult[];
  allResults: SearchResult[];
  resultIndexMap: Map<SearchResult, number>;
}

export interface DropdownPosition {
  top: number;
  left: number;
  width: number;
}

/**
 * Categorizes search results into groups, members, and guests
 * Optimized for performance with O(n) complexity
 */
export function categorizeResults(
  searchResults: SearchResult[],
): CategorizedResults {
  if (!searchResults.length) {
    return {
      groups: [],
      members: [],
      guests: [],
      allResults: [],
      resultIndexMap: new Map(),
    };
  }

  const groups: SearchResult[] = [];
  const members: SearchResult[] = [];
  const guests: SearchResult[] = [];

  // Single pass categorization for better performance
  for (const result of searchResults) {
    if (result.type === "group") {
      groups.push(result);
    } else if (result.type === "user") {
      if (result.orgRoleName === OrgUserRoles.Guest) {
        guests.push(result);
      } else {
        members.push(result);
      }
    }
  }

  const allResults = [...groups, ...members, ...guests];

  // Create index map for O(1) lookups
  const resultIndexMap = new Map();
  allResults.forEach((result, index) => {
    resultIndexMap.set(result, index);
  });

  return { groups, members, guests, allResults, resultIndexMap };
}

/**
 * Filters search results based on query and search keys
 * Optimized for case-insensitive search
 */
export function filterSearchResults(
  searchList: SearchResult[],
  searchKeys: string[],
  query: string,
): SearchResult[] {
  if (!query.trim()) return searchList;

  const lowerQuery = query.toLowerCase();
  return searchList.filter((item) =>
    searchKeys.some(
      (key) =>
        item[key] && String(item[key]).toLowerCase().includes(lowerQuery),
    ),
  );
}

/**
 * Validates a value against multiple validators
 * Returns true if valid, error message if invalid
 */
export function validate(
  value: string,
  validators: ((value: string) => boolean | string)[],
): boolean | string {
  for (const validator of validators) {
    const result = validator(value);
    if (result !== true) return result;
  }
  return true;
}

/**
 * Processes comma-separated input and returns new entries and error
 * Handles deduplication and validation
 */
export function processCommaSeparatedInput(
  raw: string,
  selectedSet: Set<string>,
  validators: ((value: string) => boolean | string)[],
): { newEntries: string[]; error: string } {
  const parts = raw
    .split(",")
    .map((s) => s.trim())
    .filter(Boolean);

  const newEntries: string[] = [];
  let error = "";

  for (const entry of parts) {
    if (selectedSet.has(entry)) continue; // Skip duplicates

    const valid = validate(entry, validators);
    if (valid === true) {
      newEntries.push(entry);
    } else {
      error = valid as string;
      break; // Stop on first error
    }
  }

  return { newEntries, error };
}

/**
 * Calculates dropdown position based on input element
 * Handles container positioning for multi-row chips
 */
export function getDropdownPosition(
  inputElement: HTMLInputElement,
): DropdownPosition {
  const rect = inputElement.getBoundingClientRect();
  const inputContainer = inputElement.closest(".input-with-role");
  const containerRect = inputContainer?.getBoundingClientRect();

  return {
    left: containerRect?.left || rect.left,
    top: (containerRect?.bottom || rect.bottom) + 2,
    width: containerRect?.width || rect.width,
  };
}

/**
 * Scrolls to highlighted item in dropdown
 * Uses nearest scroll behavior for smooth UX
 */
export function scrollToHighlighted(
  highlightedIndex: number,
  dropdownList: HTMLElement,
): void {
  if (highlightedIndex >= 0 && dropdownList) {
    const items = dropdownList.querySelectorAll(".dropdown-item");
    if (items[highlightedIndex]) {
      items[highlightedIndex].scrollIntoView({ block: "nearest" });
    }
  }
}

/**
 * Calculates next highlight index for keyboard navigation
 * Supports loop behavior and boundary checking
 */
export function getNextHighlightIndex(
  currentIndex: number,
  totalItems: number,
  direction: "up" | "down",
  loop: boolean,
): number {
  if (totalItems === 0) return -1;

  if (direction === "down") {
    if (currentIndex === totalItems - 1) {
      return loop ? 0 : currentIndex;
    }
    return currentIndex + 1;
  } else {
    if (currentIndex === 0) {
      return loop ? totalItems - 1 : currentIndex;
    }
    return currentIndex - 1;
  }
}

/**
 * Extracts the last incomplete part from comma-separated input
 * Useful for keeping partial input in the field
 */
export function getLastIncompletePart(input: string): string {
  const parts = input.split(",");
  return parts[parts.length - 1]?.trim() || "";
}

/**
 * Gets the index of a result in the categorized results
 * Uses the resultIndexMap for O(1) lookup performance
 */
export function getResultIndex(
  result: SearchResult,
  categorizedResults: CategorizedResults,
): number {
  return categorizedResults.resultIndexMap.get(result) ?? -1;
}

/**
 * Checks if focus should be maintained in multi-select mode
 * Prevents dropdown from closing when clicking inside it
 */
export function shouldMaintainFocus(
  relatedTarget: Element | null,
  dropdownList: HTMLElement | null,
  multiSelect: boolean,
): boolean {
  if (!multiSelect) return false;
  return (relatedTarget && dropdownList?.contains(relatedTarget)) || false;
}
