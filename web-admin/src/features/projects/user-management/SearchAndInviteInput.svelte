<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-management/UserRoleSelect.svelte";
  import { ProjectUserRoles } from "@rilldata/web-common/features/users/roles.ts";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import SearchAndInviteListItem from "./SearchAndInviteListItem.svelte";
  import Close from "@rilldata/web-common/components/icons/Close.svelte";
  import {
    type SearchResult,
    type DropdownPosition,
    categorizeResults,
    filterSearchResults,
    validate,
    processCommaSeparatedInput,
    getDropdownPosition,
    scrollToHighlighted,
    getNextHighlightIndex,
    getLastIncompletePart,
    shouldMaintainFocus,
    getResultIndex,
  } from "./utils";
  import { debounce } from "@rilldata/web-common/lib/create-debouncer";

  export let placeholder: string = "Search or invite by email";
  export let validators: ((value: string) => boolean | string)[] = [];
  export let roleSelect: boolean = false;
  export let initialRole: string = ProjectUserRoles.Viewer;
  export let searchList: SearchResult[] | undefined = undefined;
  export let searchKeys: string[] = [];
  export let loop: boolean = false;
  export let multiSelect: boolean = false;
  export let autoFocusInput: -1 | 0 | 1 = 0;
  export let onSearch: (query: string) => Promise<SearchResult[]>;
  export let onInvite: (
    emailsAndGroups: string[],
    role?: string,
  ) => Promise<void>;

  let input = "";
  let searchResults: SearchResult[] = [];
  let selected: string[] = [];
  let loading = false;
  let showDropdown = false;
  let error: string = "";
  let role = initialRole;
  let highlightedIndex = -1;
  let dropdownList: HTMLElement;
  let inputElement: HTMLInputElement;
  let dropdownPosition: DropdownPosition = { top: 0, left: 0, width: 0 };
  let keyboardNavigationActive = false;

  $: categorizedResults = categorizeResults(searchResults);
  $: selectedSet = new Set(selected);

  // Debounced search for better performance
  const debouncedSearch = debounce(async (query: string) => {
    let remoteResults = [] as SearchResult[];
    let localResults = [] as SearchResult[];

    try {
      if (onSearch) {
        remoteResults = await onSearch(query);
      }
    } catch {
      remoteResults = [];
    }

    if (searchList) {
      localResults = filterSearchResults(searchList, searchKeys, query);
    }

    // Merge and de-duplicate by identifier
    const seen = new Set<string>();
    const merged: SearchResult[] = [];
    for (const r of [...localResults, ...remoteResults]) {
      if (!seen.has(r.identifier)) {
        merged.push(r);
        seen.add(r.identifier);
      }
    }

    searchResults = merged;
    showDropdown = searchResults.length > 0;
    if (showDropdown) {
      updateDropdownPosition();
    }
  }, 150);

  $: if (highlightedIndex >= 0 && showDropdown && dropdownList) {
    scrollToHighlighted(highlightedIndex, dropdownList);
  }

  $: if (selected && showDropdown) {
    requestAnimationFrame(updateDropdownPosition);
  }

  function updateDropdownPosition() {
    if (inputElement) {
      dropdownPosition = getDropdownPosition(inputElement);
    }
  }

  async function handleInput(e: Event) {
    const target = e.target as HTMLInputElement;
    input = target.value;
    error = "";

    // Handle comma-separated input
    if (input.includes(",")) {
      const { newEntries, error: processError } = processCommaSeparatedInput(
        input,
        selectedSet,
        validators,
      );

      if (processError) {
        error = processError;
      } else {
        selected = [...selected, ...newEntries];
      }

      // Keep only the last incomplete part
      input = getLastIncompletePart(target.value);
    }

    loading = true;
    await debouncedSearch(input);
    loading = false;
  }

  function handleSelect(result: SearchResult) {
    if (multiSelect) {
      // Toggle selection in multi-select mode
      selected = selectedSet.has(result.identifier)
        ? selected.filter((id) => id !== result.identifier)
        : [...selected, result.identifier];

      input = "";
      refreshSearchResults();
      showDropdown = true;
      inputElement?.focus();
    } else {
      // Replace selection in single-select mode
      selected = [result.identifier];
      input = "";
      showDropdown = false;
      highlightedIndex = -1;
    }
  }

  async function refreshSearchResults() {
    try {
      const remote = onSearch ? await onSearch("") : [];
      const local = searchList ?? [];
      const seen = new Set<string>();
      const merged: SearchResult[] = [];
      for (const r of [...local, ...remote]) {
        if (!seen.has(r.identifier)) {
          merged.push(r);
          seen.add(r.identifier);
        }
      }
      searchResults = merged;
      showDropdown = searchResults.length > 0;
      if (showDropdown) updateDropdownPosition();
    } catch {
      searchResults = searchList ?? [];
      showDropdown = searchResults.length > 0;
    }
  }

  function handleInvite() {
    // Validate current input if it exists
    if (input.trim()) {
      const inputValid = validate(input, validators);
      if (inputValid === true) {
        selected = [...selected, input];
        input = "";
      } else {
        error = inputValid as string;
        return;
      }
    }

    // Validate all selected items
    const invalids = selected
      .map((item) => validate(item, validators))
      .filter((v) => v !== true);

    if (invalids.length > 0) {
      error = invalids[0] as string;
      return;
    }

    onInvite(selected, role)
      .then(() => {
        selected = [];
        input = "";
        error = "";
      })
      .catch((err) => {
        error = err.message || "Failed to invite.";
      });
  }

  function handleInputKeydown(e: KeyboardEvent) {
    // Handle Tab key
    if (e.key === "Tab" && input.trim()) {
      e.preventDefault();
      const inputValid = validate(input, validators);
      if (inputValid === true) {
        if (multiSelect && !selectedSet.has(input)) {
          selected = [...selected, input];
        } else if (!multiSelect) {
          selected = [input];
        }
        input = "";
      }
      return;
    }

    // Handle Enter key
    if (e.key === "Enter") {
      if (input.includes(",")) {
        const { newEntries, error: processError } = processCommaSeparatedInput(
          input,
          selectedSet,
          validators,
        );

        if (!processError) {
          selected = multiSelect
            ? [...selected, ...newEntries]
            : newEntries.length > 0
              ? [newEntries[0]]
              : selected;
        }
        input = "";
        e.preventDefault();
        return;
      }

      if (input.trim() === "" && selected.length > 0) {
        handleInvite();
        e.preventDefault();
        return;
      }

      if (
        showDropdown &&
        highlightedIndex >= 0 &&
        highlightedIndex < categorizedResults.allResults.length
      ) {
        handleSelect(categorizedResults.allResults[highlightedIndex]);
        e.preventDefault();
        return;
      }

      if (input && validate(input, validators) === true) {
        if (multiSelect && !selectedSet.has(input)) {
          selected = [...selected, input];
        } else if (!multiSelect) {
          selected = [input];
        }
        input = "";
        showDropdown = true;
        highlightedIndex = -1;
        e.preventDefault();
      }
    }

    // Handle arrow keys
    if (e.key === "ArrowDown" || e.key === "ArrowUp") {
      keyboardNavigationActive = true;
      const direction = e.key === "ArrowDown" ? "down" : "up";
      const newIndex = getNextHighlightIndex(
        highlightedIndex,
        categorizedResults.allResults.length,
        direction,
        loop,
      );

      if (newIndex !== highlightedIndex) {
        highlightedIndex = newIndex;
        showDropdown = true;
        updateDropdownPosition();
      }
      e.preventDefault();
    }

    // Handle Space key for multi-select
    if (e.key === "Space" && highlightedIndex >= 0 && multiSelect) {
      handleSelect(categorizedResults.allResults[highlightedIndex]);
      e.preventDefault();
      showDropdown = true;
      inputElement?.focus();
    }

    // Handle Backspace
    if (e.key === "Backspace" && input === "" && selected.length > 0) {
      selected = selected.slice(0, -1);
      e.preventDefault();
    }
  }

  async function handleFocus() {
    // On focus, load top 5 users from server (via onSearch("") -> no searchPattern)
    await refreshSearchResults();
  }

  function handleBlur(e: FocusEvent) {
    const relatedTarget = e.relatedTarget as Element;
    if (shouldMaintainFocus(relatedTarget, dropdownList, multiSelect)) {
      return;
    }
    showDropdown = false;
  }

  function handlePaste(e: ClipboardEvent) {
    const pasted = e.clipboardData?.getData("text") ?? "";
    if (pasted.includes(",")) {
      const { newEntries, error: processError } = processCommaSeparatedInput(
        pasted,
        selectedSet,
        validators,
      );

      if (!processError) {
        selected = [...selected, ...newEntries];
      }
      input = "";
      e.preventDefault();
    }
  }

  function removeSelected(identifier: string) {
    selected = selected.filter((id) => id !== identifier);
  }
</script>

<!-- Template remains largely the same but cleaner -->
<div class="invite-search-input">
  <div class="input-row">
    <div
      class={cn(
        "input-with-role p-1 border border-gray-200 outline-transparent",
        showDropdown
          ? "border-transparent outline outline-1 outline-primary-500"
          : "",
      )}
    >
      <div
        class="chips-and-input flex flex-wrap gap-1 w-full min-h-[24px] px-1 max-h-[120px] overflow-y-auto pr-1"
      >
        {#each selected as identifier (identifier)}
          <span
            class="chip text-sm w-fit h-[24px] overflow-hidden text-ellipsis"
          >
            {identifier}
            <button
              on:click={() => removeSelected(identifier)}
              class="ml-1 rounded hover:bg-gray-100 transition-colors"
            >
              <Close size="12px" />
            </button>
          </span>
        {/each}

        <input
          type="text"
          bind:value={input}
          bind:this={inputElement}
          placeholder={selected.length === 0 ? placeholder : ""}
          on:input={handleInput}
          on:keydown={handleInputKeydown}
          on:focus={handleFocus}
          on:blur={handleBlur}
          on:paste={handlePaste}
          class:error={!!error}
          autocomplete="off"
          tabindex={autoFocusInput}
          class="px-1"
        />
      </div>

      {#if roleSelect && (selected.length > 0 || input.trim())}
        <div class="shrink-0 ml-2">
          <UserRoleSelect bind:value={role} />
        </div>
      {/if}
    </div>

    <Button
      type="primary"
      onClick={handleInvite}
      disabled={selected.length === 0 && !input.trim()}
      forcedStyle="height: 32px !important; padding-left: 20px; padding-right: 20px;"
    >
      Invite
    </Button>
  </div>

  {#if error}
    <div class="error">{error}</div>
  {/if}

  {#if showDropdown && searchResults.length > 0}
    <div
      class="dropdown"
      bind:this={dropdownList}
      style="width: {dropdownPosition.width}px; top: {dropdownPosition.top}px; left: {dropdownPosition.left}px;"
      on:pointermove={() => {
        keyboardNavigationActive = false;
      }}
    >
      {#if categorizedResults.groups.length > 0}
        <div class="section-header">GROUPS</div>
        {#each categorizedResults.groups as result}
          {@const resultIndex = getResultIndex(result, categorizedResults)}
          {@const isSelected = selectedSet.has(result.identifier)}
          <SearchAndInviteListItem
            {result}
            {resultIndex}
            {isSelected}
            {highlightedIndex}
            {keyboardNavigationActive}
            onSelect={handleSelect}
            onHighlight={(index) => {
              highlightedIndex = index;
            }}
            onClearHighlight={() => {
              highlightedIndex = -1;
            }}
          />
        {/each}

        {#if categorizedResults.members.length > 0}
          <div class="section-divider"></div>
        {/if}
      {/if}

      {#if categorizedResults.members.length > 0}
        <div class="section-header">MEMBERS</div>
        {#each categorizedResults.members as result}
          {@const resultIndex = getResultIndex(result, categorizedResults)}
          {@const isSelected = selectedSet.has(result.identifier)}
          <SearchAndInviteListItem
            {result}
            {resultIndex}
            {isSelected}
            {highlightedIndex}
            {keyboardNavigationActive}
            onSelect={handleSelect}
            onHighlight={(index) => {
              highlightedIndex = index;
            }}
            onClearHighlight={() => {
              highlightedIndex = -1;
            }}
          />
        {/each}

        {#if categorizedResults.guests.length > 0}
          <div class="section-divider"></div>
        {/if}
      {/if}

      {#if categorizedResults.guests.length > 0}
        <div class="section-header">GUESTS</div>
        {#each categorizedResults.guests as result}
          {@const resultIndex = getResultIndex(result, categorizedResults)}
          {@const isSelected = selectedSet.has(result.identifier)}
          <SearchAndInviteListItem
            {result}
            {resultIndex}
            {isSelected}
            {highlightedIndex}
            {keyboardNavigationActive}
            onSelect={handleSelect}
            onHighlight={(index) => {
              highlightedIndex = index;
            }}
            onClearHighlight={() => {
              highlightedIndex = -1;
            }}
          />
        {/each}
      {/if}
    </div>
  {:else if loading}
    <div
      class="dropdown loading"
      style="width: {dropdownPosition.width}px; top: {dropdownPosition.top}px; left: {dropdownPosition.left}px;"
    >
      <div class="loading-spinner"></div>
      <span>Searching...</span>
    </div>
  {/if}
</div>

<style lang="postcss">
  .invite-search-input {
    width: 100%;
    position: relative;
  }

  .input-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }

  .input-with-role {
    display: flex;
    align-items: center;
    flex-wrap: nowrap;
    background: #fff;
    border-radius: 6px;
    min-height: 32px;
    gap: 8px;
    flex: 1;
    transition:
      outline 150ms ease-in-out,
      border-color 150ms ease-in-out;
  }

  .input-with-role input[type="text"] {
    @apply text-sm;
    border: none;
    outline: none;
    flex: 1 0 120px;
    min-width: 120px;
    padding: 0;
    background: transparent;
    margin: 0;
  }

  .chip {
    background: #f3f4f6;
    color: #222;
    border-radius: 12px;
    padding: 2px 10px;
    display: flex;
    align-items: center;
    border: 1px solid #e5e7eb;
  }

  .chip button {
    background: none;
    border: none;
    color: #888;
    margin-left: 4px;
    cursor: pointer;
  }

  .input-with-role :global(.dropdown-menu-trigger) {
    border: none;
    background: transparent;
    margin-left: 4px;
    min-width: 90px;
  }

  .dropdown {
    position: fixed;
    background: #fff;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    z-index: 50;
    min-height: 60px;
    max-height: 320px;
    overflow-y: auto;
    list-style: none;
    padding: 0;
    color: #222;
  }

  .dropdown.loading {
    display: flex;
    align-items: center;
    justify-content: center;
    gap: 8px;
    padding: 16px;
  }

  .loading-spinner {
    width: 16px;
    height: 16px;
    border: 2px solid #f3f4f6;
    border-top: 2px solid #3b82f6;
    border-radius: 50%;
    animation: spin 1s linear infinite;
  }

  @keyframes spin {
    0% {
      transform: rotate(0deg);
    }
    100% {
      transform: rotate(360deg);
    }
  }

  .section-header {
    @apply text-xs font-semibold text-gray-500 uppercase tracking-wide px-3 py-2;
    border-top: 1px solid #f3f4f6;
  }

  .section-header:first-child {
    border-top: none;
  }

  .section-divider {
    @apply border-t border-gray-200;
  }
</style>
