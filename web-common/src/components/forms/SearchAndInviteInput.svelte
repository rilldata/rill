<script lang="ts">
  import { Button } from "@rilldata/web-common/components/button";
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-management/UserRoleSelect.svelte";
  import Close from "../icons/Close.svelte";
  import { cn } from "@rilldata/web-common/lib/shadcn";
  import Check from "@rilldata/web-common/components/icons/Check.svelte";
  import Avatar from "@rilldata/web-common/components/avatar/Avatar.svelte";
  import { Chip } from "@rilldata/web-common/components/chip";
  import { getRandomBgColor } from "@rilldata/web-common/features/themes/color-config";
  import { tick, onMount, onDestroy } from "svelte";

  export let placeholder: string = "Search or invite by email";
  export let validators: ((value: string) => boolean | string)[] = [];
  export let roleSelect: boolean = false;
  export let initialRole: string = "viewer";
  export let searchList: any[] | undefined = undefined;
  export let searchKeys: string[] = [];
  export let loop: boolean = false;
  export let multiSelect: boolean = false;
  export let autoFocusInput: -1 | 0 | 1 = 0; // -1: no auto focus, 0: auto focus on mount, 1: auto focus on blur
  export let onSearch: (query: string) => Promise<any[]>;
  export let onInvite: (emails: string[], role?: string) => Promise<void>;

  let input = "";
  let searchResults: any[] = [];
  let selected: string[] = [];
  let loading = false;
  let showDropdown = false;
  let error: string = "";
  let role = initialRole;
  let highlightedIndex = -1;
  let dropdownList: HTMLElement;
  let inputElement: HTMLInputElement;
  let dropdownTop = 0;
  let dropdownLeft = 0;
  let dropdownWidth = 0;

  function updateDropdownPosition() {
    if (inputElement) {
      const rect = inputElement.getBoundingClientRect();
      const inputContainer = inputElement.closest(".input-with-role");
      const containerRect = inputContainer?.getBoundingClientRect();

      dropdownLeft = containerRect?.left || rect.left;
      dropdownTop = (containerRect?.bottom || rect.bottom) + 2;
      dropdownWidth = containerRect?.width || rect.width;
    }
  }

  function scrollToHighlighted() {
    if (highlightedIndex >= 0 && dropdownList) {
      const items = dropdownList.querySelectorAll(".dropdown-item");
      if (items[highlightedIndex]) {
        items[highlightedIndex].scrollIntoView({ block: "nearest" });
      }
    }
  }

  // Only scroll when dropdown is visible and highlighted index changes
  $: if (highlightedIndex >= 0 && showDropdown && dropdownList) {
    scrollToHighlighted();
  }

  // Update dropdown position when selected items change (for multi-row chip wrapping)
  $: if (selected && showDropdown) {
    requestAnimationFrame(() => {
      updateDropdownPosition();
    });
  }

  function processCommaSeparatedInput(raw: string) {
    // Split by comma, trim, filter out empty, and deduplicate
    const parts = raw
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    const newEntries = parts.filter((entry) => !selectedSet.has(entry));
    // Validate each entry
    for (const entry of newEntries) {
      const valid = validate(entry);
      if (valid === true) {
        selected = [...selected, entry];
      } else {
        error = valid as string;
        // Optionally: skip adding invalid, or add anyway and show error
      }
    }
  }

  async function handleInput(e: Event) {
    input = (e.target as HTMLInputElement).value;
    error = "";
    // If input contains a comma, process it
    if (input.includes(",")) {
      processCommaSeparatedInput(input);
      // Only keep the last (possibly incomplete) part in the input
      const lastPart = input.split(",").pop() ?? "";
      input = lastPart.trim();
    }
    loading = true;
    try {
      if (searchList) {
        const lower = input.toLowerCase();
        // Keep selected items in the search results but mark them as selected
        searchResults = searchList.filter((item) =>
          searchKeys.some(
            (key) => item[key] && item[key].toLowerCase().includes(lower),
          ),
        );
      } else {
        const results = await onSearch(input);
        searchResults = results;
      }
      showDropdown = searchResults.length > 0;
      if (showDropdown) {
        updateDropdownPosition();
      }
    } catch {
      searchResults = [];
      showDropdown = false;
    } finally {
      loading = false;
    }
  }

  function validate(value: string) {
    for (const v of validators) {
      const res = v(value);
      if (res !== true) return res;
    }
    return true;
  }

  function handleSelect(result: any) {
    if (multiSelect) {
      // Multi-select mode: toggle selection
      if (selectedSet.has(result.identifier)) {
        selected = selected.filter((id) => id !== result.identifier);
      } else {
        selected = [...selected, result.identifier];
      }
      // Clear input after selection
      input = "";
      // Keep dropdown open and input focused in multi-select mode
      showDropdown = true;
      inputElement?.focus();
    } else {
      // Single-select mode: replace selection
      selected = [result.identifier];
      // Clear input after selection
      input = "";
      showDropdown = false;
      highlightedIndex = -1; // Only reset highlightedIndex in single-select mode
    }
  }

  function handleInvite() {
    const invalids = selected.map(validate).filter((v) => v !== true);
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
    if (!showDropdown || searchResults.length === 0) {
      // Handle Tab key to separate values
      if (e.key === "Tab" && input.trim()) {
        e.preventDefault();
        if (validate(input) === true) {
          if (multiSelect) {
            if (!selectedSet.has(input)) {
              selected = [...selected, input];
            }
          } else {
            selected = [input];
          }
          input = "";
        }
        return;
      }
      // If input contains a comma, process it on Enter
      if (e.key === "Enter" && input.includes(",")) {
        if (multiSelect) {
          processCommaSeparatedInput(input);
        } else {
          // In single-select mode, only take the first valid input
          const firstValid = input
            .split(",")
            .map((s) => s.trim())
            .find((entry) => validate(entry) === true);
          if (firstValid) {
            selected = [firstValid];
          }
        }
        input = "";
        e.preventDefault();
        return;
      }
      // If input is empty and there are selected items, Enter should submit (invite)
      if (e.key === "Enter" && input.trim() === "" && selected.length > 0) {
        handleInvite();
        e.preventDefault();
        return;
      }
    }
    if (e.key === "ArrowDown") {
      if (highlightedIndex === categorizedResults.allResults.length - 1) {
        if (loop) {
          highlightedIndex = 0;
        } else {
          e.preventDefault();
          return;
        }
      } else {
        highlightedIndex = highlightedIndex + 1;
      }
      e.preventDefault();
      showDropdown = true;
      updateDropdownPosition();
    } else if (e.key === "ArrowUp") {
      if (highlightedIndex === 0) {
        if (loop) {
          highlightedIndex = categorizedResults.allResults.length - 1;
        } else {
          e.preventDefault();
          return;
        }
      } else {
        highlightedIndex = highlightedIndex - 1;
      }
      e.preventDefault();
      showDropdown = true;
      updateDropdownPosition();
    } else if (e.key === "Enter") {
      if (
        highlightedIndex >= 0 &&
        highlightedIndex < categorizedResults.allResults.length
      ) {
        handleSelect(categorizedResults.allResults[highlightedIndex]);
        e.preventDefault();
        // In multi-select mode, keep dropdown open and input focused
        if (multiSelect) {
          showDropdown = true;
          inputElement?.focus();
        }
      } else if (input && validate(input) === true) {
        // Allow inviting a new email
        if (multiSelect) {
          if (!selectedSet.has(input)) {
            selected = [...selected, input];
          }
        } else {
          selected = [input];
        }
        input = "";
        showDropdown = true;
        highlightedIndex = -1;
        e.preventDefault();
      }
    } else if (e.key === "Space" && highlightedIndex >= 0) {
      // Add space key support for multi-select
      if (multiSelect) {
        handleSelect(categorizedResults.allResults[highlightedIndex]);
        e.preventDefault();
        showDropdown = true;
        inputElement?.focus();
      }
    } else if (e.key === "Backspace" && input === "" && selected.length > 0) {
      // Remove the last selected chip when backspace is pressed and input is empty
      selected = selected.slice(0, -1);
      e.preventDefault();
    }
  }

  function handleFocus() {
    if (searchList) {
      const lower = input.toLowerCase();
      searchResults = searchList.filter((item) =>
        searchKeys.some(
          (key) => item[key] && item[key].toLowerCase().includes(lower),
        ),
      );
      showDropdown = searchResults.length > 0;
      if (showDropdown) {
        updateDropdownPosition();
      }
    }
  }

  function handleBlur(e: FocusEvent) {
    // In multi-select mode, only close dropdown if focus moves completely outside the component
    if (multiSelect) {
      // Check if the new focus target is within our component
      const relatedTarget = e.relatedTarget as Element;
      if (relatedTarget && dropdownList?.contains(relatedTarget)) {
        return; // Don't close if focus is moving to dropdown
      }
    }
    // Close dropdown in single-select mode or when focus moves outside component
    showDropdown = false;
  }

  function removeSelected(identifier: string) {
    selected = selected.filter((e) => e !== identifier);
  }

  $: categorizedResults = (() => {
    if (!searchResults.length) {
      return {
        groups: [],
        members: [],
        guests: [],
        allResults: [],
        resultIndexMap: new Map(),
      };
    }

    const groups = searchResults.filter((result) => result.type === "group");
    const members = searchResults.filter(
      (result) => result.type === "user" && result.orgRoleName !== "guest",
    );
    const guests = searchResults.filter(
      (result) => result.type === "user" && result.orgRoleName === "guest",
    );
    const allResults = [...groups, ...members, ...guests];

    // Create index map for O(1) lookups instead of O(n) indexOf calls
    const resultIndexMap = new Map();
    allResults.forEach((result, index) => {
      resultIndexMap.set(result, index);
    });

    return { groups, members, guests, allResults, resultIndexMap };
  })();

  // Create a Set for O(1) selected lookups instead of O(n) includes() calls
  $: selectedSet = new Set(selected);

  function getResultIndex(result: any): number {
    return categorizedResults.resultIndexMap.get(result) ?? -1;
  }

  function getInitials(name: string) {
    return name.charAt(0).toUpperCase();
  }
</script>

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
        class="chips-and-input flex flex-wrap gap-1 w-full min-h-[20px] px-1"
      >
        {#each selected as identifier (identifier)}
          <span class="chip text-sm w-fit h-5 overflow-hidden text-ellipsis"
            >{identifier}
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
          on:paste={(e) => {
            const pasted = e.clipboardData?.getData("text") ?? "";
            if (pasted.includes(",")) {
              processCommaSeparatedInput(pasted);
              input = "";
              e.preventDefault();
            }
          }}
          class:error={!!error}
          autocomplete="off"
          tabindex={autoFocusInput}
          class="px-1"
        />
      </div>
      {#if roleSelect}
        <div class="role-select-container">
          <UserRoleSelect bind:value={role} />
        </div>
      {/if}
    </div>
    <Button
      type="primary"
      on:click={handleInvite}
      disabled={selected.length === 0}
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
      style="width: {dropdownWidth}px; top: {dropdownTop}px; left: {dropdownLeft}px;"
    >
      <!-- TODO: hoist item -->
      {#if categorizedResults.groups.length > 0}
        <div class="section-header">GROUPS</div>
        {#each categorizedResults.groups as result}
          {@const resultIndex = getResultIndex(result)}
          {@const isSelected = selectedSet.has(result.identifier)}
          <button
            type="button"
            class:highlighted={resultIndex === highlightedIndex}
            class:selected={isSelected}
            class="dropdown-item"
            on:click={(e) => {
              e.preventDefault();
              handleSelect(result);
            }}
            on:keydown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                handleSelect(result);
              }
            }}
            on:pointerdown={(e) => {
              e.preventDefault();
            }}
            on:pointerenter={() => {
              highlightedIndex = resultIndex;
            }}
            on:pointerleave={() => {
              highlightedIndex = -1;
            }}
          >
            <div class="flex items-center gap-2">
              <div
                class={cn(
                  "h-7 w-7 rounded-sm flex items-center justify-center",
                  getRandomBgColor(result.identifier),
                )}
              >
                <span class="text-sm text-white font-semibold"
                  >{getInitials(result.identifier)}</span
                >
              </div>
              <div class="flex flex-col text-left">
                <span class="text-sm font-medium text-gray-900"
                  >{result.identifier}</span
                >
                {#if result.groupCount !== undefined}
                  <span class="text-xs text-gray-500">
                    {result.groupCount} user{result.groupCount > 1 ? "s" : ""}
                  </span>
                {/if}
              </div>
            </div>
            {#if isSelected}
              <Check size="16px" className="ui-copy-icon" />
            {/if}
          </button>
        {/each}

        {#if categorizedResults.members.length > 0}
          <div class="section-divider"></div>
        {/if}
      {/if}

      <!-- TODO: hoist item -->
      {#if categorizedResults.members.length > 0}
        <div class="section-header">MEMBERS</div>
        {#each categorizedResults.members as result}
          {@const resultIndex = getResultIndex(result)}
          {@const isSelected = selectedSet.has(result.identifier)}
          <button
            type="button"
            class:highlighted={resultIndex === highlightedIndex}
            class:selected={isSelected}
            class="dropdown-item"
            on:click={(e) => {
              e.preventDefault();
              handleSelect(result);
            }}
            on:keydown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                handleSelect(result);
              }
            }}
            on:pointerdown={(e) => {
              e.preventDefault();
            }}
            on:pointerenter={() => {
              highlightedIndex = resultIndex;
            }}
            on:pointerleave={() => {
              highlightedIndex = -1;
            }}
          >
            <div class="flex items-center gap-2">
              <Avatar
                avatarSize="h-7 w-7"
                fontSize="text-xs"
                src={result.photoUrl}
                alt={result.invitedBy ? undefined : result.name}
                bgColor={getRandomBgColor(result.identifier)}
              />
              <div class="flex flex-col text-left">
                <span class="text-sm font-medium text-gray-900">
                  {result.identifier}
                </span>
                <span class="text-xs text-gray-500"
                  >{result.invitedBy ? "Pending invitation" : result.name}</span
                >
              </div>
            </div>
            {#if isSelected}
              <Check size="16px" className="ui-copy-icon" />
            {/if}
          </button>
        {/each}

        {#if categorizedResults.guests.length > 0}
          <div class="section-divider"></div>
        {/if}
      {/if}

      <!-- TODO: hoist item -->
      {#if categorizedResults.guests.length > 0}
        <div class="section-header">GUESTS</div>
        {#each categorizedResults.guests as result}
          {@const resultIndex = getResultIndex(result)}
          {@const isSelected = selectedSet.has(result.identifier)}
          <button
            type="button"
            class:highlighted={resultIndex === highlightedIndex}
            class:selected={isSelected}
            class="dropdown-item"
            on:click={(e) => {
              e.preventDefault();
              handleSelect(result);
            }}
            on:keydown={(e) => {
              if (e.key === "Enter" || e.key === " ") {
                e.preventDefault();
                handleSelect(result);
              }
            }}
            on:pointerdown={(e) => {
              e.preventDefault();
            }}
            on:pointerenter={() => {
              highlightedIndex = resultIndex;
            }}
            on:pointerleave={() => {
              highlightedIndex = -1;
            }}
          >
            <div class="flex items-center gap-2">
              <Avatar
                avatarSize="h-7 w-7"
                fontSize="text-xs"
                src={result.photoUrl}
                alt={result.invitedBy ? undefined : result.name}
                bgColor={getRandomBgColor(result.identifier)}
              />
              <div class="flex flex-col text-left">
                <span
                  class="text-sm font-medium text-gray-900 flex flex-row items-center gap-x-1"
                >
                  {result.identifier}
                  <Chip type="amber" label="Guest" compact readOnly>
                    <svelte:fragment slot="body">Guest</svelte:fragment>
                  </Chip>
                </span>
                <span class="text-xs text-gray-500"
                  >{result.invitedBy ? "Pending invitation" : result.name}</span
                >
              </div>
            </div>
            {#if isSelected}
              <Check size="16px" className="ui-copy-icon" />
            {/if}
          </button>
        {/each}
      {/if}
    </div>
  {:else if loading}
    <div
      class="dropdown loading"
      style="width: {dropdownWidth}px; top: {dropdownTop}px; left: {dropdownLeft}px;"
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
    min-height: 40px;
    gap: 8px;
    flex: 1;
    transition:
      outline 150ms ease-in-out,
      border-color 150ms ease-in-out;
  }

  .role-select-container {
    min-width: 71px;
    max-width: 71px;
    flex: 0 0 71px;
    display: flex;
    align-items: center;
    justify-content: flex-end;
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
  .dropdown button {
    padding: 8px 12px;
    cursor: pointer;
    scroll-margin: 8px;
    display: flex;
    align-items: center;
    justify-content: space-between;
    transition: background-color 150ms ease-in-out;
  }
  .dropdown button:hover {
    @apply bg-slate-100;
  }
  .dropdown button.highlighted {
    @apply bg-slate-200;
    scroll-snap-align: start;
  }
  .dropdown button.selected {
    @apply bg-slate-100;
  }
  .dropdown button.selected:hover {
    @apply bg-slate-200;
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

  .dropdown-item {
    @apply flex items-center justify-between px-3 py-2 cursor-pointer w-full text-left border-none bg-transparent;
    scroll-margin: 8px;
    transition: background-color 150ms ease-in-out;
  }

  .dropdown-item:hover {
    @apply bg-slate-100;
  }

  .dropdown-item.highlighted {
    @apply bg-slate-200;
    scroll-snap-align: start;
  }

  .dropdown-item.selected {
    @apply bg-slate-100;
  }

  .dropdown-item.selected:hover {
    @apply bg-slate-200;
  }
</style>
