<script lang="ts">
  import type {
    RpcStatus,
    V1ListOrganizationInvitesResponse,
    V1ListOrganizationMemberUsersResponse,
    V1OrganizationMemberUser,
    V1OrganizationInvite,
    V1OrganizationPermissions,
  } from "@rilldata/web-admin/client";
  import UserCompositeCell from "@rilldata/web-admin/features/organizations/user-management/table/users/UserCompositeCell.svelte";
  import UserActionsCell from "@rilldata/web-admin/features/organizations/user-management/table/users/UserActionsCell.svelte";
  import UserRoleCell from "@rilldata/web-admin/features/organizations/user-management/table/users/UserRoleCell.svelte";
  import UserGroupsCell from "@rilldata/web-admin/features/organizations/user-management/table/users/UserGroupsCell.svelte";
  import UserProjectsCell from "@rilldata/web-admin/features/organizations/user-management/table/users/UserProjectsCell.svelte";
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
  export let organizationPermissions: V1OrganizationPermissions;
  export let billingContact: string | undefined;
  export let scrollToTopTrigger: any = null;
  export let guestOnly: boolean;

  export let onAttemptRemoveBillingContactUser: () => void;
  export let onAttemptChangeBillingContactUserRole: () => void;
  export let onEditUserGroup: (groupName: string) => void;
  export let onConvertToMember: (user: V1OrganizationMemberUser) => void;

  $: safeData = Array.isArray(data) ? data : [];

  const UserCell = <ColumnDef<OrgUser, any>>{
    accessorKey: "user",
    header: "User",
    enableSorting: false,
    cell: ({ row }) =>
      flexRender(UserCompositeCell, {
        name: row.original.userName ?? row.original.email,
        email: row.original.userEmail,
        isCurrentUser: row.original.userEmail === currentUserEmail,
        pendingAcceptance: Boolean(row.original.invitedBy),
        photoUrl: row.original.userPhotoUrl,
        role: row.original.roleName,
      }),
    meta: {
      widthPercent: 50,
    },
  };
  const RoleCell = <ColumnDef<OrgUser, any>>{
    accessorKey: "roleName",
    header: "Organization Role",
    cell: ({ row }) =>
      flexRender(UserRoleCell, {
        email: row.original.userEmail,
        role: row.original.roleName,
        isCurrentUser: row.original.userEmail === currentUserEmail,
        organizationPermissions,
        isBillingContact: row.original.userEmail === billingContact,
        onAttemptChangeBillingContactUserRole,
      }),
    meta: {
      widthPercent: 40,
      marginLeft: "8px",
    },
  };
  const UserGroupCell = <ColumnDef<OrgUser, any>>{
    accessorKey: "usergroupsCount",
    header: "Groups",
    cell: ({ row }) =>
      flexRender(UserGroupsCell, {
        userId: row.original.userId,
        organization,
        onEditUserGroup,
      }),
    meta: {
      widthPercent: 40,
      marginLeft: "8px",
    },
  };
  const ProjectsCell = <ColumnDef<OrgUser, any>>{
    accessorKey: "projectsCount",
    header: "Projects",
    cell: ({ row }) =>
      flexRender(UserProjectsCell, {
        organization,
        userId: row.original.userId,
      }),
    meta: {
      widthPercent: 40,
      marginLeft: "8px",
    },
  };
  const ContextActionsCell = <ColumnDef<OrgUser, any>>{
    accessorKey: "actions",
    header: "",
    enableSorting: false,
    cell: ({ row }) =>
      flexRender(UserActionsCell, {
        email: row.original.userEmail,
        role: row.original.roleName,
        isCurrentUser: row.original.userEmail === currentUserEmail,
        organizationPermissions,
        isBillingContact: row.original.userEmail === billingContact,
        onAttemptRemoveBillingContactUser,
        onConvertToMember: () => onConvertToMember(row.original),
      }),
    meta: {
      widthPercent: 5,
    },
  };
  $: columns = guestOnly
    ? [UserCell, UserGroupCell, ProjectsCell, ContextActionsCell]
    : [UserCell, RoleCell, UserGroupCell];

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
