<script lang="ts">
  import type {
    RpcStatus,
    V1ListOrganizationInvitesResponse,
    V1ListOrganizationMemberUsersResponse,
    V1OrganizationMemberUser,
    V1OrganizationInvite,
    V1Project,
    V1Usergroup,
  } from "@rilldata/web-admin/client";
  import OrgUsersTableUserCompositeCell from "./OrgUsersTableUserCompositeCell.svelte";
  import OrgUsersTableActionsCell from "./OrgUsersTableActionsCell.svelte";
  import OrgUsersTableRoleCell from "./OrgUsersTableRoleCell.svelte";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import type {
    InfiniteData,
    InfiniteQueryObserverResult,
  } from "@tanstack/svelte-query";
  import { ExternalLinkIcon } from "lucide-svelte";
  import InfiniteScrollTable from "@rilldata/web-common/components/table/InfiniteScrollTable.svelte";
  import { onMount } from "svelte";
  import {
    adminServiceListProjectsForOrganizationAndUser,
    adminServiceListUsergroupsForOrganizationAndUser,
  } from "@rilldata/web-admin/client/gen/admin-service/admin-service";

  interface OrgUser extends V1OrganizationMemberUser, V1OrganizationInvite {
    invitedBy?: string;
    projects?: V1Project[];
    groups?: V1Usergroup[];
  }

  export let data: OrgUser[];
  export let usersQuery: InfiniteQueryObserverResult<
    InfiniteData<V1ListOrganizationMemberUsersResponse, unknown>,
    RpcStatus
  >;
  export let invitesQuery: InfiniteQueryObserverResult<
    InfiniteData<V1ListOrganizationInvitesResponse, unknown>,
    RpcStatus
  >;
  export let currentUserEmail: string;
  export let currentUserRole: string;
  export let billingContact: string | undefined;
  export let scrollToTopTrigger: any = null;
  export let organization: string;

  export let onAttemptRemoveBillingContactUser: () => void;
  export let onAttemptChangeBillingContactUserRole: () => void;

  $: safeData = Array.isArray(data) ? data : [];

  // Loading state for fetching projects and groups
  let loadingExtraData = true;
  let enrichedData: OrgUser[] = [];

  async function fetchProjectsAndGroupsForUsers(users: OrgUser[], org: string) {
    // If no users, nothing to do
    if (!users.length) return [];
    // Use userId if available, else userEmail
    const results = await Promise.all(
      users.map(async (user) => {
        let projects: V1Project[] = [];
        let groups: V1Usergroup[] = [];
        try {
          const [projectsRes, groupsRes] = await Promise.all([
            adminServiceListProjectsForOrganizationAndUser(org, {
              userId: user.userId,
            }),
            adminServiceListUsergroupsForOrganizationAndUser(org, {
              userId: user.userId,
            }),
          ]);
          console.log("projectsRes", projectsRes);
          console.log("groupsRes", groupsRes);
          projects = projectsRes.projects || [];
          groups = groupsRes.usergroups || [];
        } catch (e) {
          // Optionally handle error
        }
        return { ...user, projects, groups };
      }),
    );
    return results;
  }

  onMount(async () => {
    loadingExtraData = true;
    enrichedData = await fetchProjectsAndGroupsForUsers(safeData, organization);
    loadingExtraData = false;
  });

  $: columns = <ColumnDef<OrgUser, any>[]>[
    {
      accessorKey: "user",
      header: "User",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(OrgUsersTableUserCompositeCell, {
          name: row.original.userName ?? row.original.email,
          email: row.original.userEmail,
          pendingAcceptance: Boolean(row.original.invitedBy),
          isCurrentUser: row.original.userEmail === currentUserEmail,
          photoUrl: row.original.userPhotoUrl,
          role: row.original.roleName,
        }),
      meta: {
        widthPercent: 35,
      },
    },
    {
      accessorKey: "roleName",
      header: "Organization Role",
      cell: ({ row }) =>
        flexRender(OrgUsersTableRoleCell, {
          email: row.original.userEmail,
          role: row.original.roleName,
          isCurrentUser: row.original.userEmail === currentUserEmail,
          currentUserRole,
          isBillingContact: row.original.userEmail === billingContact,
          onAttemptChangeBillingContactUserRole,
        }),
      meta: {
        widthPercent: 20,
        marginLeft: "8px",
      },
    },
    {
      accessorKey: "groups",
      header: "Groups",
      cell: ({ row }) =>
        row.original.groups && row.original.groups.length
          ? `${row.original.groups.length} group${row.original.groups.length > 1 ? "s" : ""}`
          : "None",
      meta: {
        widthPercent: 15,
      },
    },
    {
      accessorKey: "projects",
      header: "Projects",
      cell: ({ row }) =>
        row.original.projects && row.original.projects.length
          ? `${row.original.projects.length} project${row.original.projects.length > 1 ? "s" : ""}`
          : "None",
      meta: {
        widthPercent: 15,
      },
    },
    {
      accessorKey: "actions",
      header: "",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(OrgUsersTableActionsCell, {
          email: row.original.userEmail,
          role: row.original.roleName,
          isCurrentUser: row.original.userEmail === currentUserEmail,
          currentUserRole,
          isBillingContact: row.original.userEmail === billingContact,
          onAttemptRemoveBillingContactUser,
        }),
      meta: {
        widthPercent: 5,
      },
    },
  ];

  function handleLoadMore() {
    if (usersQuery.hasNextPage) {
      usersQuery.fetchNextPage();
    }
    if (invitesQuery.hasNextPage) {
      invitesQuery.fetchNextPage();
    }
  }

  $: dynamicTableMaxHeight =
    safeData.length > 12 ? `calc(100dvh - 300px)` : "auto";

  const headerIcons = {
    roleName: {
      icon: ExternalLinkIcon,
      href: "https://docs.rilldata.com/manage/roles-permissions#organization-level-permissions",
    },
  };
</script>

{#if loadingExtraData}
  <div class="flex items-center justify-center py-8">
    Loading user projects and groups...
  </div>
{:else}
  <InfiniteScrollTable
    data={enrichedData}
    {columns}
    hasNextPage={usersQuery.hasNextPage || invitesQuery.hasNextPage}
    isFetchingNextPage={usersQuery.isFetchingNextPage ||
      invitesQuery.isFetchingNextPage}
    onLoadMore={handleLoadMore}
    maxHeight={dynamicTableMaxHeight}
    emptyStateMessage="No users found"
    {headerIcons}
    {scrollToTopTrigger}
  />
{/if}
