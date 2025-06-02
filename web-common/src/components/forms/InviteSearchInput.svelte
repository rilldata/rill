<script lang="ts">
  import { createEventDispatcher, onMount } from "svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-management/UserRoleSelect.svelte";

  export let onSearch: (query: string) => Promise<any[]>;
  export let onInvite: (emails: string[], role?: string) => Promise<void>;
  export let placeholder: string = "Search or invite by email";
  export let validators: ((value: string) => boolean | string)[] = [];
  export let roleSelect: boolean = false;
  export let initialRole: string = "viewer";
  export let searchList: any[] | undefined = undefined;
  export let searchKeys: string[] = [];
  export let autoFocusInput: -1 | 0 | 1 = 0;

  const dispatch = createEventDispatcher();

  let input = "";
  let searchResults: any[] = [];
  let selected: string[] = [];
  let loading = false;
  let showDropdown = false;
  let error: string = "";
  let role = initialRole;
  let highlightedIndex = -1;
  let dropdownList: HTMLElement;

  const DROPDOWN_WIDTH = 406.62;

  function scrollToHighlighted() {
    if (highlightedIndex >= 0 && dropdownList) {
      const items = dropdownList.getElementsByTagName("li");
      if (items[highlightedIndex]) {
        items[highlightedIndex].scrollIntoView({ block: "nearest" });
      }
    }
  }

  $: if (highlightedIndex >= 0) {
    scrollToHighlighted();
  }

  function processCommaSeparatedInput(raw: string) {
    // Split by comma, trim, filter out empty, and deduplicate
    const parts = raw
      .split(",")
      .map((s) => s.trim())
      .filter(Boolean);
    const newEntries = parts.filter((entry) => !selected.includes(entry));
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
        // TODO: Modify search logic to:
        // 1. Keep selected items in the search results but mark them as selected
        // 2. Add visual indicator (e.g. checkbox or different styling) for selected items
        // 3. Allow toggling selection state of items in the dropdown
        searchResults = searchList
          .filter((item) =>
            searchKeys.some(
              (key) => item[key] && item[key].toLowerCase().includes(lower),
            ),
          )
          .filter((item) => !selected.includes(item.identifier));
      } else {
        const results = await onSearch(input);
        searchResults = results.filter(
          (item) => !selected.includes(item.identifier),
        );
      }
      showDropdown = searchResults.length > 0;
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
    // TODO: Modify selection logic to:
    // 1. Toggle selection state instead of just adding
    // 2. Keep item in dropdown but mark as selected
    // 3. Update visual state of item in dropdown
    if (!selected.includes(result.identifier)) {
      selected = [...selected, result.identifier];
      input = "";
      showDropdown = false;
      highlightedIndex = -1;
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
        dispatch("inviteSuccess", { emails: selected });
        selected = [];
        input = "";
        error = "";
      })
      .catch((err) => {
        error = err.message || "Failed to invite.";
        dispatch("inviteError", { error });
      });
  }

  function handleInputKeydown(e: KeyboardEvent) {
    if (!showDropdown || searchResults.length === 0) {
      // Handle Tab key to separate values
      if (e.key === "Tab" && input.trim()) {
        e.preventDefault();
        if (validate(input) === true && !selected.includes(input)) {
          selected = [...selected, input];
          input = "";
        }
        return;
      }
      // If input contains a comma, process it on Enter
      if (e.key === "Enter" && input.includes(",")) {
        processCommaSeparatedInput(input);
        input = "";
        e.preventDefault();
        return;
      }
    }
    if (e.key === "ArrowDown") {
      highlightedIndex = (highlightedIndex + 1) % searchResults.length;
      e.preventDefault();
    } else if (e.key === "ArrowUp") {
      highlightedIndex =
        (highlightedIndex - 1 + searchResults.length) % searchResults.length;
      e.preventDefault();
    } else if (e.key === "Enter") {
      if (highlightedIndex >= 0 && highlightedIndex < searchResults.length) {
        handleSelect(searchResults[highlightedIndex]);
        e.preventDefault();
      } else if (input && validate(input) === true) {
        // Allow inviting a new email
        if (!selected.includes(input)) {
          selected = [...selected, input];
          input = "";
          showDropdown = false;
          highlightedIndex = -1;
        }
        e.preventDefault();
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
      searchResults = searchList
        .filter((item) =>
          searchKeys.some(
            (key) => item[key] && item[key].toLowerCase().includes(lower),
          ),
        )
        .filter((item) => !selected.includes(item.identifier));
      showDropdown = searchResults.length > 0;
    }
  }

  function handleBlur() {
    showDropdown = false;
  }

  function removeSelected(identifier: string) {
    selected = selected.filter((e) => e !== identifier);
  }
</script>

<div class="invite-search-input">
  <div class="input-row">
    <div class="input-with-role p-1">
      <div
        class="chips-and-input flex flex-wrap gap-1 w-full min-h-[24px] px-1"
      >
        {#each selected as identifier (identifier)}
          <span class="chip"
            >{identifier}<button
              type="button"
              on:click={() => removeSelected(identifier)}>&times;</button
            ></span
          >
        {/each}
        <input
          type="text"
          bind:value={input}
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
          class="outline outline-1 outline-primary-500 px-1"
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
    <ul
      class="dropdown"
      bind:this={dropdownList}
      style="width: {DROPDOWN_WIDTH}px; left: 0;"
    >
      {#each searchResults as result, i}
        <li
          class:highlighted={i === highlightedIndex}
          class="hover:bg-slate-100"
        >
          <!-- TODO: Add checkbox or other visual indicator for selection state -->
          <button
            type="button"
            class="w-full text-left"
            on:pointerdown={() => handleSelect(result)}
          >
            {result.identifier}
          </button>
        </li>
      {/each}
    </ul>
  {:else if loading}
    <div class="dropdown loading" style="width: {DROPDOWN_WIDTH}px; left: 0;">
      <div class="loading-spinner"></div>
      <span>Searching...</span>
    </div>
  {/if}
</div>

<style>
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
    border: 1px solid #d1d5db;
    border-radius: 6px;
    min-height: 40px;
    gap: 8px;
    flex: 1;
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
    border: none;
    outline: none;
    flex: 1 0 120px;
    min-width: 120px;
    padding: 0;
    background: transparent;
    color: #222;
    margin: 0;
  }
  .chip {
    background: #f3f4f6;
    color: #222;
    border-radius: 12px;
    padding: 2px 10px;
    font-size: 0.95em;
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
    position: absolute;
    background: #fff;
    border: 1px solid #d1d5db;
    border-radius: 6px;
    margin-top: 2px;
    z-index: 10;
    max-height: 180px;
    overflow-y: auto;
    list-style: none;
    padding: 0;
    color: #222;
  }
  .dropdown li {
    padding: 8px 12px;
    cursor: pointer;
    scroll-margin: 8px;
  }
  .dropdown li.highlighted {
    @apply bg-slate-100;
    scroll-snap-align: start;
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
</style>
