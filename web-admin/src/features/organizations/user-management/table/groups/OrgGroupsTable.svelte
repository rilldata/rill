<script lang="ts">
  import type { V1MemberUsergroup } from "@rilldata/web-admin/client";
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import GroupActionsCell from "@rilldata/web-admin/features/organizations/user-management/table/groups/GroupActionsCell.svelte";
  import GroupCompositeCell from "@rilldata/web-admin/features/organizations/user-management/table/groups/GroupCompositeCell.svelte";
  import InfiniteScrollTable from "@rilldata/web-common/components/table/InfiniteScrollTable.svelte";

  export let data: V1MemberUsergroup[];
  export let currentUserEmail: string;
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
        flexRender(GroupCompositeCell, {
          groupName: row.original.groupName,
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
        flexRender(GroupActionsCell, {
          groupName: row.original.groupName,
          currentUserEmail: currentUserEmail,
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
