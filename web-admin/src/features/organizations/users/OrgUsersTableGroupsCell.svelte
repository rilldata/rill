<script lang="ts">
  import { getUserGroupsForUsersInOrg } from "@rilldata/web-admin/features/organizations/users/selectors.ts";
  import * as Popover from "@rilldata/web-common/components/popover";
  import CaretDownIcon from "@rilldata/web-common/components/icons/CaretDownIcon.svelte";
  import CaretUpIcon from "@rilldata/web-common/components/icons/CaretUpIcon.svelte";

  export let organization: string;
  export let userId: string;

  let isDropdownOpen = false;

  const userGroupsQuery = getUserGroupsForUsersInOrg(organization, userId);
  $: ({ data: userGroups, isPending, error } = $userGroupsQuery);
</script>

<Popover.Root bind:open={isDropdownOpen}>
  <Popover.Trigger
    class="w-18 flex flex-row gap-1 items-center rounded-sm {isDropdownOpen
      ? 'bg-slate-200'
      : 'hover:bg-slate-100'} px-2 py-1"
  >
    <span class="capitalize">
      {userGroups.length} Group{userGroups.length > 1 ? "s" : ""}
    </span>
    {#if isDropdownOpen}
      <CaretUpIcon size="12px" />
    {:else}
      <CaretDownIcon size="12px" />
    {/if}
  </Popover.Trigger>
  <Popover.Content>
    {#if isPending}
      Loading...
    {:else if error}
      Error
    {:else}
      <div class="flex flex-col gap-y-1">
        {#each userGroups as userGroup (userGroup.id)}
          <div class="flex flex-row items-center text-xs">
            <span class="text-gray-700">{userGroup.name}</span>
            {#if userGroup.count > 0}
              <span class="text-gray-500">
                {userGroup.count} member{userGroup.count > 1 ? "s" : ""}
              </span>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </Popover.Content>
</Popover.Root>
