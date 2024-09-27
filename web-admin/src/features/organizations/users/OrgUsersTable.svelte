<script lang="ts">
  import type { V1MemberUser } from "@rilldata/web-admin/client";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import { flexRender } from "@tanstack/svelte-table";
  import {
    formatDate,
    capitalize,
  } from "@rilldata/web-common/components/table/utils";
  import OrgUsersTableUserCompositeCell from "./OrgUsersTableUserCompositeCell.svelte";
  import OrgUsersTableActionsCell from "./OrgUsersTableActionsCell.svelte";

  export let data: V1MemberUser[];
  export let onRemove: (email: string) => void;

  const columns: ColumnDef<V1MemberUser, any>[] = [
    {
      accessorKey: "user",
      header: "User",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(OrgUsersTableUserCompositeCell, {
          name: row.original.userName,
          email: row.original.userEmail,
        }),
    },
    {
      accessorKey: "roleName",
      header: "Role",
      cell: ({ row }) => {
        if (!row.original.roleName) return "-";
        return capitalize(row.original.roleName);
      },
    },
    {
      accessorKey: "createdOn",
      header: "Created On",
      cell: ({ row }) => {
        if (!row.original.createdOn) return "-";
        return formatDate(row.original.createdOn);
      },
    },
    {
      accessorKey: "updatedOn",
      header: "Updated On",
      cell: ({ row }) => {
        if (!row.original.updatedOn) return "-";
        return formatDate(row.original.updatedOn);
      },
    },
    {
      accessorKey: "actions",
      header: "",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(OrgUsersTableActionsCell, {
          email: row.original.userEmail,
          onRemove: onRemove,
        }),
    },
  ];
</script>

<BasicTable {data} {columns} />
