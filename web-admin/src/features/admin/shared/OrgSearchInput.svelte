<!-- Org name input with search dropdown powered by project name search -->
<script lang="ts">
  import { createEventDispatcher } from "svelte";
  import { searchOrgNames } from "@rilldata/web-admin/features/admin/organizations/selectors";

  export let value = "";
  export let placeholder = "Organization name";

  const dispatch = createEventDispatcher<{ select: string }>();

  let showDropdown = false;
  let justSelected = false;

  $: orgSearchQuery = searchOrgNames(value);
  $: orgNames = extractUniqueOrgs($orgSearchQuery.data?.names ?? []);

  function extractUniqueOrgs(projectPaths: string[]): string[] {
    const orgs = new Set<string>();
    for (const path of projectPaths) {
      const org = path.split("/")[0];
      if (org) orgs.add(org);
    }
    return [...orgs].sort();
  }

  function selectOrg(org: string) {
    value = org;
    showDropdown = false;
    justSelected = true;
    dispatch("select", org);
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
    // Delay to allow mousedown on dropdown items to fire first
    setTimeout(() => {
      showDropdown = false;
    }, 150);
  }

  // When new results arrive, show dropdown only if user is actively typing
  $: if (orgNames.length > 0 && value.length >= 3 && !justSelected) {
    showDropdown = true;
  }
</script>

<div class="relative">
  <input
    type="text"
    class="w-full px-3 py-2 text-sm rounded-md border border-slate-300 dark:border-slate-600 bg-white dark:bg-slate-800 text-slate-900 dark:text-slate-100 placeholder:text-slate-400 dark:placeholder:text-slate-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
    {placeholder}
    bind:value
    on:keydown={handleKeydown}
    on:input={handleInput}
    on:blur={handleBlur}
  />
  {#if $orgSearchQuery.isFetching && value.length >= 3}
    <div
      class="absolute right-3 top-1/2 -translate-y-1/2 w-4 h-4 border-2 border-slate-300 border-t-blue-600 rounded-full animate-spin"
    />
  {/if}
  {#if showDropdown && orgNames.length > 0}
    <div
      class="absolute z-10 w-full mt-1 bg-white dark:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-md shadow-lg max-h-48 overflow-y-auto"
    >
      {#each orgNames as org}
        <button
          class="w-full text-left px-3 py-2 text-sm text-slate-700 dark:text-slate-300 hover:bg-slate-100 dark:hover:bg-slate-700 cursor-pointer"
          on:mousedown|preventDefault={() => selectOrg(org)}
        >
          {org}
        </button>
      {/each}
    </div>
  {/if}
</div>
