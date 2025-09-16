<script lang="ts">
  import type {
    RpcStatus,
    V1ListOrganizationInvitesResponse,
    V1ListOrganizationMemberUsersResponse,
    V1OrganizationMemberUser,
    V1OrganizationInvite,
  } from "@rilldata/web-admin/client";
  import OrgUsersTableUserCompositeCell from "./OrgUsersTableUserCompositeCell.svelte";
  import OrgUsersTableActionsCell from "./OrgUsersTableActionsCell.svelte";
  import OrgUsersTableRoleCell from "./OrgUsersTableRoleCell.svelte";
  import OrgUsersTableGroupsCell from "./OrgUsersTableGroupsCell.svelte";
  import OrgUsersTableProjectsCell from "./OrgUsersTableProjectsCell.svelte";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import type {
    InfiniteData,
    InfiniteQueryObserverResult,
  } from "@tanstack/svelte-query";
  import { ExternalLinkIcon } from "lucide-svelte";
  import InfiniteScrollTable from "@rilldata/web-common/components/table/InfiniteScrollTable.svelte";

  interface OrgUser extends V1OrganizationMemberUser, V1OrganizationInvite {
    invitedBy?: string;
  }

  export let organization: string;
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

  export let onAttemptRemoveBillingContactUser: () => void;
  export let onAttemptChangeBillingContactUserRole: () => void;
  export let onEditUserGroup: (groupName: string) => void;

  $: safeData = Array.isArray(data) ? data : [];

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
        widthPercent: 50,
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
        widthPercent: 40,
        marginLeft: "8px",
      },
    },
    {
      accessorKey: "usergroupsCount",
      header: "Groups",
      cell: ({ row }) =>
        flexRender(OrgUsersTableGroupsCell, {
          userId: row.original.userId,
          organization,
          onEditUserGroup,
        }),
      meta: {
        widthPercent: 40,
        marginLeft: "8px",
      },
    },
    {
      accessorKey: "projectsCount",
      header: "Projects",
      cell: ({ row }) =>
        flexRender(OrgUsersTableProjectsCell, {
          userId: row.original.userId,
          organization,
        }),
      meta: {
        widthPercent: 40,
        marginLeft: "8px",
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

<InfiniteScrollTable
  data={safeData}
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
