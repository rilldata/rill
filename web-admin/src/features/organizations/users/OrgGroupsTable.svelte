<script lang="ts">
  import type { V1MemberUsergroup } from "@rilldata/web-admin/client";
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import OrgGroupsActionsCell from "./OrgGroupsActionsCell.svelte";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import { formatDate } from "@rilldata/web-common/components/table/utils";

  export let data: V1MemberUsergroup[];
  export let onDelete: (deletedGroupName: string) => void;
  export let onSetRole: (groupName: string, role: string) => void;

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
        return row.original.roleName;
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
        flexRender(OrgGroupsActionsCell, {
          name: row.original.groupName,
          onDelete: onDelete,
          onSetRole: onSetRole,
        }),
    },
  ];
</script>

<BasicTable {data} {columns} />
