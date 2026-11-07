<script lang="ts">
  import { m } from "@rilldata/web-common/lib/i18n/gen/messages";
  import { page } from "$app/stores";
  import { createAdminServiceGetCurrentUser } from "@rilldata/web-admin/client";
  import CreateUserGroupDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/CreateUserGroupDialog.svelte";
  import EditUserGroupDialog from "@rilldata/web-admin/features/organizations/user-management/dialogs/EditUserGroupDialog.svelte";
  import OrgGroupsTable from "@rilldata/web-admin/features/organizations/user-management/table/groups/OrgGroupsTable.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { Plus } from "lucide-svelte";
  import { getOrgUsergroupsInfinite } from "@rilldata/web-admin/features/organizations/user-management/selectors.ts";

  let userGroupName = "";
  let isCreateUserGroupDialogOpen = false;
  let isEditUserGroupDialogOpen = false;
  let searchText = "";

  $: organization = $page.params.organization;
  $: listOrganizationMemberUsergroups = getOrgUsergroupsInfinite(organization);
  const currentUser = createAdminServiceGetCurrentUser();

  $: allGroups = $listOrganizationMemberUsergroups.data?.pages.flatMap(
    (page) => page.members ?? [],
  );
  $: filteredGroups =
    allGroups?.filter(
      (group) =>
        !group.groupManaged &&
        group.groupName.toLowerCase().includes(searchText.toLowerCase()),
    ) ?? [];

  $: hasNextPage = Boolean($listOrganizationMemberUsergroups.hasNextPage);

  $: isFetchingNextPage = $listOrganizationMemberUsergroups.isFetchingNextPage;

  function handleLoadMore() {
    if ($listOrganizationMemberUsergroups.hasNextPage) {
      $listOrganizationMemberUsergroups.fetchNextPage();
    }
  }

  // Handle URL parameter to open edit dialog
  $: if (
    $page.url.searchParams.get("action") === "open-edit-user-group-dialog"
  ) {
    const groupName = $page.url.searchParams.get("groupName");
    if (groupName) {
      userGroupName = groupName;
      isEditUserGroupDialogOpen = true;
    }
    // Clear the URL parameters
    $page.url.searchParams.delete("action");
    $page.url.searchParams.delete("groupName");
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
      {m.groups_error_loading()}
      {$listOrganizationMemberUsergroups.error}
    </div>
  {:else if $listOrganizationMemberUsergroups.isSuccess}
    <div class="flex flex-col">
      <div class="flex flex-row gap-x-4">
        <Search
          bind:value={searchText}
          large
          autofocus={false}
          showBorderOnFocus={false}
        />
        <Button
          type="primary"
          large
          onClick={() => (isCreateUserGroupDialogOpen = true)}
        >
          <Plus size="16px" />
          <span>{m.groups_create_group()}</span>
        </Button>
      </div>
      <div class="mt-6">
        <OrgGroupsTable
          data={filteredGroups}
          currentUserEmail={$currentUser.data?.user.email}
          {hasNextPage}
          {isFetchingNextPage}
          onLoadMore={handleLoadMore}
        />
      </div>
      {#if filteredGroups.length > 0}
        <div class="px-2 py-3">
          <span class="font-medium text-sm text-fg-secondary">
            {m.groups_total_count({ count: filteredGroups.length })}
          </span>
        </div>
      {/if}
    </div>
  {/if}
</div>

<CreateUserGroupDialog
  bind:open={isCreateUserGroupDialogOpen}
  groupName={userGroupName}
  currentUserEmail={$currentUser.data?.user.email}
/>

<EditUserGroupDialog
  bind:open={isEditUserGroupDialogOpen}
  groupName={userGroupName}
  currentUserEmail={$currentUser.data?.user.email}
/>
