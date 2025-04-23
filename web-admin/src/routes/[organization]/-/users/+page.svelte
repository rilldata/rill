<script lang="ts">
  import { page } from "$app/stores";
  import type { V1UserInvite } from "@rilldata/web-admin/client";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizationInvitesInfinite,
    createAdminServiceListOrganizationMemberUsersInfinite,
  } from "@rilldata/web-admin/client";
  import ChangeBillingContactDialog from "@rilldata/web-admin/features/billing/contact/ChangeBillingContactDialog.svelte";
  import { getOrganizationBillingContactUser } from "@rilldata/web-admin/features/billing/contact/selectors";
  import AddUsersDialog from "@rilldata/web-admin/features/organizations/users/AddUsersDialog.svelte";
  import OrgUsersTable from "@rilldata/web-admin/features/organizations/users/OrgUsersTable.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { Plus } from "lucide-svelte";

  const PAGE_SIZE = 12;

  let userEmail = "";
  let userRole = "";
  let isSuperUser = false;
  let isAddUserDialogOpen = false;
  let isUpdateBillingContactDialogOpen = false;
  let searchText = "";

  $: organization = $page.params.organization;

  $: orgMemberUsersInfiniteQuery =
    createAdminServiceListOrganizationMemberUsersInfinite(
      organization,
      {
        pageSize: PAGE_SIZE,
      },
      {
        query: {
          getNextPageParam: (lastPage) => {
            if (lastPage.nextPageToken !== "") {
              return lastPage.nextPageToken;
            }
            return undefined;
          },
        },
      },
    );
  $: orgInvitesInfiniteQuery =
    createAdminServiceListOrganizationInvitesInfinite(
      organization,
      {
        pageSize: PAGE_SIZE,
      },
      {
        query: {
          getNextPageParam: (lastPage) => {
            if (lastPage.nextPageToken !== "") {
              return lastPage.nextPageToken;
            }
            return undefined;
          },
        },
      },
    );

  $: allOrgMemberUsersRows =
    $orgMemberUsersInfiniteQuery.data?.pages.flatMap(
      (page) => page.members ?? [],
    ) ?? [];
  $: allOrgInvitesRows =
    $orgInvitesInfiniteQuery.data?.pages.flatMap(
      (page) => page.invites ?? [],
    ) ?? [];

  function coerceInvitesToUsers(invites: V1UserInvite[]) {
    return invites.map((invite) => ({
      ...invite,
      userEmail: invite.email,
      roleName: invite.role,
    }));
  }

  $: combinedRows = [
    ...allOrgMemberUsersRows,
    ...coerceInvitesToUsers(allOrgInvitesRows),
  ];

  // Search by email or name
  // Member users have a userName field, invites do not
  $: filteredUsers = combinedRows.filter((user) => {
    const searchLower = searchText.toLowerCase();
    return (
      (user.userEmail?.toLowerCase() || "").includes(searchLower) ||
      ("userName" in user &&
        (user.userName?.toLowerCase() || "").includes(searchLower))
    );
  });

  const currentUser = createAdminServiceGetCurrentUser();
  $: billingContactUser = getOrganizationBillingContactUser(organization);
</script>

<div class="flex flex-col w-full">
  {#if $orgMemberUsersInfiniteQuery.isLoading || $orgInvitesInfiniteQuery.isLoading}
    <DelayedSpinner
      isLoading={$orgMemberUsersInfiniteQuery.isLoading ||
        $orgInvitesInfiniteQuery.isLoading}
      size="1rem"
    />
  {:else if $orgMemberUsersInfiniteQuery.isError || $orgInvitesInfiniteQuery.isError}
    <div class="text-red-500">
      Error loading organization members: {$orgMemberUsersInfiniteQuery.error ??
        $orgInvitesInfiniteQuery.error}
    </div>
  {:else if $orgMemberUsersInfiniteQuery.isSuccess && $orgInvitesInfiniteQuery.isSuccess}
    <div class="flex flex-col gap-4">
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
          on:click={() => (isAddUserDialogOpen = true)}
        >
          <Plus size="16px" />
          <span>Add users</span>
        </Button>
      </div>
      <OrgUsersTable
        data={filteredUsers}
        usersQuery={$orgMemberUsersInfiniteQuery}
        invitesQuery={$orgInvitesInfiniteQuery}
        currentUserEmail={$currentUser.data?.user.email}
        billingContact={$billingContactUser?.email}
        onChangeBillingContact={() => (isUpdateBillingContactDialogOpen = true)}
      />
    </div>
  {/if}
</div>

<AddUsersDialog
  bind:open={isAddUserDialogOpen}
  email={userEmail}
  role={userRole}
  {isSuperUser}
/>

<ChangeBillingContactDialog
  bind:open={isUpdateBillingContactDialogOpen}
  {organization}
  currentBillingContact={$billingContactUser?.email}
/>
