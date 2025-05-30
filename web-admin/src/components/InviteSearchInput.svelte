<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { Button } from "@rilldata/web-common/components/button";
  import UserRoleSelect from "@rilldata/web-admin/features/projects/user-management/UserRoleSelect.svelte";

  export let onSearch: (query: string) => Promise<any[]>;
  export let onInvite: (emails: string[], role?: string) => Promise<void>;
  export let placeholder: string = "Search or invite by email";
  export let validators: ((value: string) => boolean | string)[] = [];
  export let roleSelect: boolean = false;
  export let initialRole: string = "viewer";
  export let searchList: any[] | undefined = undefined;

  const dispatch = createEventDispatcher();

  let input = "";
  let searchResults: any[] = [];
  let selected: string[] = [];
  let loading = false;
  let showDropdown = false;
  let error: string = "";
  let role = initialRole;
  let highlightedIndex = -1;

  async function handleInput(e: Event) {
    input = (e.target as HTMLInputElement).value;
    error = "";
    if (!input) {
      searchResults = [];
      showDropdown = false;
      return;
    }
    loading = true;
    try {
      if (searchList) {
        console.log("SearchList:", searchList);
        const lower = input.toLowerCase();
        searchResults = searchList.filter(
          (u) =>
            (u.email && u.email.toLowerCase().includes(lower)) ||
            (u.userEmail && u.userEmail.toLowerCase().includes(lower)) ||
            (u.name && u.name.toLowerCase().includes(lower)) ||
            (u.userName && u.userName.toLowerCase().includes(lower)),
        );
        console.log("Filtered searchList results:", searchResults);
        showDropdown = true;
      } else {
        searchResults = await onSearch(input);
        showDropdown = true;
      }
    } catch (err) {
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
    if (!selected.includes(result.email)) {
      selected = [...selected, result.email];
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
    if (!showDropdown || searchResults.length === 0) return;
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
    }
  }

  function removeSelected(email: string) {
    selected = selected.filter((e) => e !== email);
  }
</script>

<div class="invite-search-input">
  <div class="selected-list">
    {#each selected as email (email)}
      <span class="chip"
        >{email}
        <button type="button" on:click={() => removeSelected(email)}
          >&times;</button
        ></span
      >
    {/each}
  </div>
  <div class="input-row">
    <input
      type="text"
      bind:value={input}
      {placeholder}
      on:input={handleInput}
      on:keydown={handleInputKeydown}
      on:focus={() => (showDropdown = searchResults.length > 0)}
      class:error={!!error}
      autocomplete="off"
    />
    {#if roleSelect}
      <UserRoleSelect bind:value={role} />
    {/if}
    <Button on:click={handleInvite} disabled={selected.length === 0}>
      Invite
    </Button>
  </div>
  {#if error}
    <div class="error">{error}</div>
  {/if}
  {#if showDropdown && searchResults.length > 0}
    <ul class="dropdown">
      {#each searchResults as result, i}
        <li
          class:highlighted={i === highlightedIndex}
          on:mousedown={() => handleSelect(result)}
        >
          {result.email}
        </li>
      {/each}
    </ul>
  {/if}
</div>

<style>
  .invite-search-input {
    width: 100%;
    position: relative;
  }
  .selected-list {
    display: flex;
    flex-wrap: wrap;
    gap: 6px;
    margin-bottom: 4px;
  }
  .chip {
    background: #f3f4f6; /* light gray */
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
  .input-row {
    display: flex;
    align-items: center;
    gap: 8px;
  }
  input[type="text"] {
    flex: 1;
    padding: 8px;
    border-radius: 6px;
    border: 1px solid #d1d5db;
    background: #fff;
    color: #222;
  }
  input[type="text"].error {
    border-color: #e74c3c;
  }
  .error {
    color: #e74c3c;
    margin-top: 4px;
    font-size: 0.95em;
  }
  .dropdown {
    position: absolute;
    left: 0;
    right: 0;
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
  }
  .dropdown li.highlighted {
    background: #f3f4f6;
  }
  .dropdown li.invite-new {
    color: #4caf50;
  }
</style>
