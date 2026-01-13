<script lang="ts">
  import { getUserGroupsForUsersInOrg } from "@rilldata/web-admin/features/organizations/user-management/selectors.ts";
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { writable } from "svelte/store";

  export let organization: string;
  export let userId: string;
  export let groupCount: number;
  export let onEditUserGroup: (groupName: string) => void;

  let isDropdownOpen = false;
  const userGroupsEnabledStore = writable(false);
  $: userGroupsEnabledStore.set(isDropdownOpen);

  const userGroupsQuery = getUserGroupsForUsersInOrg(
    organization,
    userId,
    userGroupsEnabledStore,
  );
  $: ({ data: userGroups, isPending, error } = $userGroupsQuery);
  $: hasGroups = groupCount > 0;
</script>

{#if hasGroups}
  <Dropdown.Root bind:open={isDropdownOpen}>
    <Dropdown.Trigger
      class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
        ? 'bg-gray-200'
        : 'hover:bg-gray-100'} px-2 py-1"
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
    <Dropdown.Content>
      {#if isPending}
        Loading...
      {:else if error}
        Error
      {:else}
        {#each userGroups as userGroup (userGroup.id)}
          <Dropdown.Item on:click={() => onEditUserGroup(userGroup.name)}>
            <span class="text-surface-foreground">{userGroup.name}</span>
            {#if userGroup.count > 0}
              <span class="text-muted-foreground">
                {userGroup.count} member{userGroup.count > 1 ? "s" : ""}
              </span>
            {/if}
          </Dropdown.Item>
        {/each}
      {/if}
    </Dropdown.Content>
  </Dropdown.Root>
{:else}
  <div class="w-18 rounded-sm px-2 py-1 text-muted-foreground">No groups</div>
{/if}
