<script lang="ts">
  import type {
    V1OrganizationMemberUser,
    V1MemberUsergroup,
  } from "@rilldata/web-admin/client";
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import OrgGroupsTableActionsCell from "./OrgGroupsTableActionsCell.svelte";
  import OrgGroupsTableGroupCompositeCell from "./OrgGroupsTableGroupCompositeCell.svelte";
  import InfiniteScrollTable from "@rilldata/web-common/components/table/InfiniteScrollTable.svelte";

  export let data: V1MemberUsergroup[];
  export let currentUserEmail: string;
  export let searchUsersList: V1OrganizationMemberUser[];
  export let hasNextPage: boolean;
  export let isFetchingNextPage: boolean;
  export let onLoadMore: () => void;

  function transformGroupName(groupName: string) {
    return groupName
      .replace("autogroup:", "")
      .replace("_", " ")
      .replace(/\b\w/g, (char) => char.toUpperCase());
  }

  const columns: ColumnDef<V1MemberUsergroup, any>[] = [
    {
      accessorKey: "groupName",
      header: "Group",
      enableSorting: true,
      sortDescFirst: true,
      cell: ({ row }) =>
        flexRender(OrgGroupsTableGroupCompositeCell, {
          name: row.original.groupName?.startsWith("autogroup:")
            ? transformGroupName(row.original.groupName)
            : row.original.groupName,
          usersCount: row.original.usersCount,
        }),
      meta: {
        widthPercent: 95,
      },
    },
    // {
    //   accessorKey: "roleName",
    //   header: "Role",
    //   cell: ({ row }) =>
    //     flexRender(OrgGroupsTableRoleCell, {
    //       name: row.original.groupName,
    //       role: row.original.roleName,
    //       manageOrgAdmins: manageOrgAdmins,
    //     }),
    //   meta: {
    //     widthPercent: 20,
    //     marginLeft: "8px",
    //   },
    // },
    {
      accessorKey: "actions",
      header: "",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(OrgGroupsTableActionsCell, {
          groupName: row.original.groupName,
          currentUserEmail: currentUserEmail,
          searchUsersList: searchUsersList,
        }),
      meta: {
        widthPercent: 5,
      },
    },
  ];

  $: dynamicTableMaxHeight = data.length > 12 ? `calc(100dvh - 300px)` : "auto";
</script>

<InfiniteScrollTable
  {data}
  {columns}
  {hasNextPage}
  {isFetchingNextPage}
  {onLoadMore}
  maxHeight={dynamicTableMaxHeight}
  emptyStateMessage="No groups found"
/>
