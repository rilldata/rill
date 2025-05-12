<script lang="ts">
  import { page } from "$app/stores";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizationMemberUsergroups,
    createAdminServiceListOrganizationMemberUsers,
  } from "@rilldata/web-admin/client";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import OrgGroupsTable from "@rilldata/web-admin/features/organizations/users/OrgGroupsTable.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Plus } from "lucide-svelte";
  import CreateUserGroupDialog from "@rilldata/web-admin/features/organizations/users/CreateUserGroupDialog.svelte";
  import { Search } from "@rilldata/web-common/components/search";

  export let data;

  $: ({ organizationPermissions } = data);
  $: manageOrgAdmins = organizationPermissions.manageOrgAdmins;

  const PAGE_SIZE = 20;

  let userGroupName = "";
  let isCreateUserGroupDialogOpen = false;
  let searchText = "";
  let pageToken = "";

  $: organization = $page.params.organization;
  $: listOrganizationMemberUsergroups =
    createAdminServiceListOrganizationMemberUsergroups(organization, {
      pageSize: PAGE_SIZE,
      pageToken,
      includeCounts: true,
    });
  $: listOrganizationMemberUsers =
    createAdminServiceListOrganizationMemberUsers(organization);

  const currentUser = createAdminServiceGetCurrentUser();

  $: filteredGroups =
    $listOrganizationMemberUsergroups.data?.members.filter(
      (group) =>
        !group.groupManaged &&
        group.groupName.toLowerCase().includes(searchText.toLowerCase()),
    ) ?? [];

  $: hasNextPage = Boolean(
    $listOrganizationMemberUsergroups.data?.nextPageToken,
  );
  $: isFetchingNextPage = $listOrganizationMemberUsergroups.isFetching;

  function handleLoadMore() {
    if (hasNextPage) {
      pageToken = $listOrganizationMemberUsergroups.data?.nextPageToken ?? "";
    }
  }
</script>

<div class="flex flex-col w-full">
  {#if $listOrganizationMemberUsergroups.isLoading}
    <DelayedSpinner
      isLoading={$listOrganizationMemberUsergroups.isLoading}
      size="1rem"
    />
  {:else if $listOrganizationMemberUsergroups.isError}
    <div class="text-red-500">
      Error loading organization user groups: {$listOrganizationMemberUsergroups.error}
    </div>
  {:else if $listOrganizationMemberUsergroups.isSuccess && $listOrganizationMemberUsers.isSuccess}
    <div class="flex flex-col">
      <div class="flex flex-row gap-x-4">
        <Search
          placeholder="Search"
          bind:value={searchText}
          large
          autofocus={false}
          showBorderOnFocus={false}
        />
        <Button
          type="primary"
          large
          on:click={() => (isCreateUserGroupDialogOpen = true)}
        >
          <Plus size="16px" />
          <span>Create group</span>
        </Button>
      </div>
      <div class="mt-6">
        <OrgGroupsTable
          data={filteredGroups}
          currentUserEmail={$currentUser.data?.user.email}
          searchUsersList={$listOrganizationMemberUsers.data?.members ?? []}
          {hasNextPage}
          {isFetchingNextPage}
          {manageOrgAdmins}
          onLoadMore={handleLoadMore}
        />
      </div>
      {#if filteredGroups.length > 0}
        <div class="px-2 py-3">
          <span class="font-medium text-sm text-gray-500">
            {filteredGroups.length} total group{filteredGroups.length === 1
              ? ""
              : "s"}
          </span>
        </div>
      {/if}
    </div>
  {/if}
</div>

<CreateUserGroupDialog
  bind:open={isCreateUserGroupDialogOpen}
  groupName={userGroupName}
  searchUsersList={$listOrganizationMemberUsers.data?.members ?? []}
  currentUserEmail={$currentUser.data?.user.email}
/>
