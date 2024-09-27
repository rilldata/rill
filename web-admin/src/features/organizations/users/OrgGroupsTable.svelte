<script lang="ts">
  import type { V1MemberUsergroup } from "@rilldata/web-admin/client";
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import OrgGroupsTableActionsCell from "./OrgGroupsTableActionsCell.svelte";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import {
    formatDate,
    capitalize,
  } from "@rilldata/web-common/components/table/utils";

  export let data: V1MemberUsergroup[];
  export let onRename: (groupName: string, newName: string) => void;
  export let onDelete: (deletedGroupName: string) => void;
  export let onAddRole: (groupName: string, role: string) => void;
  export let onSetRole: (groupName: string, role: string) => void;
  export let onRevokeRole: (groupName: string) => void;

  const columns: ColumnDef<V1MemberUsergroup, any>[] = [
    {
      accessorKey: "groupName",
      header: "Name",
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
        flexRender(OrgGroupsTableActionsCell, {
          name: row.original.groupName,
          role: row.original.roleName,
          onRename: onRename,
          onDelete: onDelete,
          onAddRole: onAddRole,
          onSetRole: onSetRole,
          onRevokeRole: onRevokeRole,
        }),
    },
  ];
</script>

<BasicTable {data} {columns} />
