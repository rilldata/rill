<script lang="ts">
  import type { V1MemberUser } from "@rilldata/web-admin/client";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import {
    formatDate,
    capitalize,
  } from "@rilldata/web-common/components/table/utils";

  export let data: V1MemberUser[];

  const columns: ColumnDef<V1MemberUser, any>[] = [
    {
      accessorKey: "userName",
      header: "User",
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
      accessorKey: "userEmail",
      header: "Email",
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
  ];
</script>

<BasicTable {data} {columns} />
