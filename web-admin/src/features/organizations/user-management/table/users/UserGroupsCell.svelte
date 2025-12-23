<script lang="ts">
  import { getUserGroupsForUsersInOrg } from "@rilldata/web-admin/features/organizations/user-management/selectors.ts";
  import { determineDropdownAlign } from "@rilldata/web-admin/features/organizations/user-management/table/dropdownAlignment.ts";
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { writable } from "svelte/store";
  import { browser } from "$app/environment";
  import { onDestroy, onMount, tick } from "svelte";

  export let organization: string;
  export let userId: string;
  export let groupCount: number;
  export let onEditUserGroup: (groupName: string) => void;

  let isDropdownOpen = false;
  const userGroupsEnabledStore = writable(false);
  $: userGroupsEnabledStore.set(isDropdownOpen);
  let dropdownAlign: "start" | "end" = "start";
  let dropdownTriggerEl: HTMLElement | null = null;
  let dropdownContentEl: HTMLElement | null = null;

  const userGroupsQuery = getUserGroupsForUsersInOrg(
    organization,
    userId,
    userGroupsEnabledStore,
  );
  $: ({ data: userGroups, isPending, error } = $userGroupsQuery);
  $: hasGroups = groupCount > 0;

  async function updateDropdownAlignment() {
    if (!browser || !isDropdownOpen || !dropdownTriggerEl) return;
    await tick();

    const menuWidth =
      dropdownContentEl?.offsetWidth ??
      dropdownTriggerEl?.offsetWidth ??
      200;

    dropdownAlign = determineDropdownAlign({
      triggerRect: dropdownTriggerEl.getBoundingClientRect(),
      menuWidth,
      viewportWidth: window.innerWidth,
    });
  }

  function handleWindowResize() {
    void updateDropdownAlignment();
  }

  onMount(() => {
    if (!browser) return;
    window.addEventListener("resize", handleWindowResize);
  });

  onDestroy(() => {
    if (!browser) return;
    window.removeEventListener("resize", handleWindowResize);
  });

  $: if (isDropdownOpen) {
    void updateDropdownAlignment();
  }

  $: if (isDropdownOpen && dropdownContentEl) {
    void updateDropdownAlignment();
  }

  $: if (isDropdownOpen && userGroups?.length !== undefined) {
    void updateDropdownAlignment();
  }
</script>

{#if hasGroups}
  <Dropdown.Root bind:open={isDropdownOpen}>
    <Dropdown.Trigger
      bind:this={dropdownTriggerEl}
      class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
        ? 'bg-slate-200'
        : 'hover:bg-slate-100'} px-2 py-1"
    >
      <span class="capitalize">
        {groupCount} Group{groupCount !== 1 ? "s" : ""}
      </span>
      {#if isDropdownOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </Dropdown.Trigger>
    <Dropdown.Content bind:this={dropdownContentEl} align={dropdownAlign}>
      {#if isPending}
        Loading...
      {:else if error}
        Error
      {:else}
        {#each userGroups as userGroup (userGroup.id)}
          <Dropdown.Item on:click={() => onEditUserGroup(userGroup.name)}>
            <span class="text-gray-700">{userGroup.name}</span>
            {#if userGroup.count > 0}
              <span class="text-gray-500">
                {userGroup.count} member{userGroup.count > 1 ? "s" : ""}
              </span>
            {/if}
          </Dropdown.Item>
        {/each}
      {/if}
    </Dropdown.Content>
  </Dropdown.Root>
{:else}
  <div class="w-18 rounded-sm px-2 py-1 text-gray-400">No groups</div>
{/if}
