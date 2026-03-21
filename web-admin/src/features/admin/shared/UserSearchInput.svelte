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

<div class="relative">
  <input
    type="text"
    class="w-full px-3 py-2 text-sm rounded-md border border-slate-300 bg-slate-50 text-slate-900 placeholder:text-slate-400 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
    {placeholder}
    bind:value
    on:keydown={handleKeydown}
    on:input={handleInput}
    on:blur={handleBlur}
  />
  {#if $usersQuery.isFetching && value.length >= 3}
    <div
      class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin"
    />
  {/if}
  {#if showDropdown && emails.length > 0}
    <div
      class="absolute z-10 w-full mt-1 bg-slate-50 border border-slate-200 rounded-md shadow-lg max-h-48 overflow-y-auto"
    >
      {#each emails as email}
        <button
          class="w-full text-left px-3 py-2 text-sm text-slate-700 hover:bg-slate-100 cursor-pointer"
          on:mousedown|preventDefault={() => selectUser(email)}
        >
          {email}
        </button>
      {/each}
    </div>
  {/if}
</div>
