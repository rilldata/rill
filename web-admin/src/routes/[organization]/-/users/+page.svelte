<script lang="ts">
  import { page } from "$app/stores";
  import type { V1UserInvite } from "@rilldata/web-admin/client";
  import {
    createAdminServiceGetCurrentUser,
    createAdminServiceListOrganizationInvitesInfinite,
    createAdminServiceListOrganizationMemberUsersInfinite,
  } from "@rilldata/web-admin/client";
  import AddUsersDialog from "@rilldata/web-admin/features/organizations/users/AddUsersDialog.svelte";
  import OrgUsersTable from "@rilldata/web-admin/features/organizations/users/OrgUsersTable.svelte";
  import Button from "@rilldata/web-common/components/button/Button.svelte";
  import OrgUsersFilters from "@rilldata/web-common/components/menu/OrgUsersFilters.svelte";
  import { Search } from "@rilldata/web-common/components/search";
  import DelayedSpinner from "@rilldata/web-common/features/entity-management/DelayedSpinner.svelte";
  import { Plus } from "lucide-svelte";

  const PAGE_SIZE = 12;

  let userEmail = "";
  let userRole = "";
  let isSuperUser = false;
  let isAddUserDialogOpen = false;
  let searchText = "";
  let filterSelection: "all" | "members" | "guests" | "pending" = "all";

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

  // Filter by role
  // Filter by search text
  $: filteredUsers = combinedRows
    .filter((user) => {
      const searchLower = searchText.toLowerCase();
      const matchesSearch =
        (user.userEmail?.toLowerCase() || "").includes(searchLower) ||
        ("userName" in user &&
          (user.userName?.toLowerCase() || "").includes(searchLower));

      let matchesRole = false;

      if (filterSelection === "all") {
        // All org users (members + guests)
        matchesRole = !("invitedBy" in user);
      } else if (filterSelection === "members") {
        // Only members (org admin, editor, viewer)
        matchesRole =
          !("invitedBy" in user) &&
          (user.roleName === "admin" ||
            user.roleName === "editor" ||
            user.roleName === "viewer");
      } else if (filterSelection === "guests") {
        // Only guests
        matchesRole = user.roleName === "guest";
      } else if (filterSelection === "pending") {
        // Only users with pending invites
        matchesRole = "invitedBy" in user;
      }

      return matchesSearch && matchesRole;
    })
    .sort((a, b) => {
      // Sort by current user first
      if (a.userEmail === $currentUser.data?.user.email) return -1;
      if (b.userEmail === $currentUser.data?.user.email) return 1;
      return 0;
    });

  const currentUser = createAdminServiceGetCurrentUser();
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
        <OrgUsersFilters bind:filterSelection />
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
        currentUserRole={$currentUser.data?.user.role}
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
