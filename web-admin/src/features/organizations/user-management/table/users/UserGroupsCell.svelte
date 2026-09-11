<script lang="ts">
  import { getUserGroupsForUsersInOrg } from "@rilldata/web-admin/features/organizations/user-management/selectors.ts";
  import * as Dropdown from "@rilldata/web-common/components/dropdown-menu";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";
  import { writable } from "svelte/store";
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";

  export let organization: string;
  export let userId: string;
  export let groupCount: number;
  // For a pending invitee, the groups stored on the invite (there is no user to query yet)
  export let pendingAcceptance: boolean = false;
  export let usergroups: string[] = [];
  export let onEditUserGroup: (groupName: string) => void;
  export let onManageGroups: (() => void) | undefined = undefined;

  let isDropdownOpen = false;
  const userGroupsEnabledStore = writable(false);
  $: userGroupsEnabledStore.set(isDropdownOpen && !pendingAcceptance);

  const userGroupsQuery = getUserGroupsForUsersInOrg(
    organization,
    userId,
    userGroupsEnabledStore,
  );
  $: ({ data: userGroups, isPending, error } = $userGroupsQuery);
  $: count = pendingAcceptance ? usergroups.length : groupCount;
  $: hasGroups = count > 0;
</script>

{#if hasGroups}
  <Dropdown.Root bind:open={isDropdownOpen}>
    <Dropdown.Trigger
      class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
        ? 'bg-gray-200'
        : 'hover:bg-surface-hover'} px-2 py-1"
    >
      <span class="capitalize">
        {m.users_group_count({ count })}
      </span>
      {#if isDropdownOpen}
        <CaretUpIcon size="12px" />
      {:else}
        <CaretDownIcon size="12px" />
      {/if}
    </Dropdown.Trigger>
    <Dropdown.Content align="start">
      {#if pendingAcceptance}
        {#each usergroups as name (name)}
          <Dropdown.Item onclick={() => onEditUserGroup(name)}>
            <span class="text-fg-primary">{name}</span>
          </Dropdown.Item>
        {/each}
      {:else if isPending}
        {m.users_loading()}
      {:else if error}
        {m.users_error()}
      {:else}
        {#each userGroups as userGroup (userGroup.id)}
          <Dropdown.Item onclick={() => onEditUserGroup(userGroup.name)}>
            <span class="text-fg-primary">{userGroup.name}</span>
            {#if userGroup.count > 0}
              <span class="text-fg-secondary">
                {m.users_member_count({ count: userGroup.count })}
              </span>
            {/if}
          </Dropdown.Item>
        {/each}
      {/if}
      {#if onManageGroups}
        <Dropdown.Separator />
        <Dropdown.Item onclick={onManageGroups}>
          <span class="text-fg-primary">{m.users_manage_groups()}</span>
        </Dropdown.Item>
      {/if}
    </Dropdown.Content>
  </Dropdown.Root>
{:else}
  <div class="w-18 rounded-sm px-2 py-1 text-fg-secondary">
    {m.users_no_groups()}
  </div>
{/if}
