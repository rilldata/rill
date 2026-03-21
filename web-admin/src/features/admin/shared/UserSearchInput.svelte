<!-- User email input with search dropdown powered by user search -->
<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { searchUsers } from "@rilldata/web-admin/features/admin/users/selectors";

  export let value = "";
  export let placeholder = "Search by email...";

  const dispatch = createEventDispatcher<{ select: string }>();

  let showDropdown = false;
  let justSelected = false;

  $: usersQuery = searchUsers(value);
  $: emails = ($usersQuery.data?.users ?? []).map((u) => u.email ?? "").filter(Boolean);

  function selectUser(email: string) {
    value = email;
    showDropdown = false;
    justSelected = true;
    dispatch("select", email);
  }

  function handleKeydown(e: KeyboardEvent) {
    if (e.key === "Enter") {
      showDropdown = false;
      justSelected = true;
      dispatch("select", value);
    }
  }

  function handleInput() {
    justSelected = false;
    if (value.length >= 3) {
      showDropdown = true;
    } else {
      showDropdown = false;
    }
  }

  function handleBlur() {
    setTimeout(() => {
      showDropdown = false;
    }, 150);
  }

  $: if (emails.length > 0 && value.length >= 3 && !justSelected) {
    showDropdown = true;
  }
</script>

<div class="search-container">
  <input
    type="text"
    class="input"
    {placeholder}
    bind:value
    on:keydown={handleKeydown}
    on:input={handleInput}
    on:blur={handleBlur}
  />
  {#if $usersQuery.isFetching && value.length >= 3}
    <div class="search-spinner" />
  {/if}
  {#if showDropdown && emails.length > 0}
    <div class="dropdown">
      {#each emails as email}
        <button
          class="dropdown-item"
          on:mousedown|preventDefault={() => selectUser(email)}
        >
          {email}
        </button>
      {/each}
    </div>
  {/if}
</div>

<style lang="postcss">
  .search-container {
    @apply relative;
  }

  .input {
    @apply w-full px-3 py-2 text-sm rounded-md border border-slate-300
      dark:border-slate-600 bg-white dark:bg-slate-800
      text-slate-900 dark:text-slate-100
      placeholder:text-slate-400 dark:placeholder:text-slate-500
      focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent;
  }

  .search-spinner {
    @apply absolute right-3 top-1/2 -translate-y-1/2
      w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin;
  }

  .dropdown {
    @apply absolute z-10 w-full mt-1 bg-white dark:bg-slate-800
      border border-slate-200 dark:border-slate-700
      rounded-md shadow-lg max-h-48 overflow-y-auto;
  }

  .dropdown-item {
    @apply w-full text-left px-3 py-2 text-sm text-slate-700 dark:text-slate-300
      hover:bg-slate-100 dark:hover:bg-slate-700 cursor-pointer;
  }
</style>
