<script lang="ts">
  import type { V1MemberUsergroup } from "@rilldata/web-admin/client";
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import OrgGroupsTableActionsCell from "./OrgGroupsTableActionsCell.svelte";
  import OrgGroupsTableRoleCell from "./OrgGroupsTableRoleCell.svelte";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import { formatDate } from "@rilldata/web-common/components/table/utils";

  export let data: V1MemberUsergroup[];
  export let currentUserEmail: string;
  export let onRename: (groupName: string, newName: string) => void;
  export let onDelete: (deletedGroupName: string) => void;
  export let onAddRole: (groupName: string, role: string) => void;
  export let onSetRole: (groupName: string, role: string) => void;
  export let onRevokeRole: (groupName: string) => void;
  export let onRemoveUser: (groupName: string, email: string) => void;

  const columns: ColumnDef<V1MemberUsergroup, any>[] = [
    {
      accessorKey: "groupName",
      header: "Name",
      meta: {
        widthPercent: 3,
      },
    },
    {
      accessorKey: "roleName",
      header: "Role",
      enableSorting: false,
      meta: {
        widthPercent: 3,
      },
      cell: ({ row }) =>
        flexRender(OrgGroupsTableRoleCell, {
          name: row.original.groupName,
          role: row.original.roleName,
          onAddRole: onAddRole,
          onSetRole: onSetRole,
          onRevokeRole: onRevokeRole,
        }),
    },
    // TODO: use relative datetime
    {
      accessorKey: "createdOn",
      header: "Created On",
      cell: ({ row }) => {
        if (!row.original.createdOn) return "-";
        return formatDate(row.original.createdOn);
      },
      meta: {
        widthPercent: 10,
      },
    },
    // TODO: use relative datetime
    {
      accessorKey: "updatedOn",
      header: "Updated On",
      cell: ({ row }) => {
        if (!row.original.updatedOn) return "-";
        return formatDate(row.original.updatedOn);
      },
      meta: {
        widthPercent: 10,
      },
    },
    {
      accessorKey: "actions",
      header: "",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(OrgGroupsTableActionsCell, {
          name: row.original.groupName,
          currentUserEmail: currentUserEmail,
          onRename: onRename,
          onDelete: onDelete,
          onRemoveUser: onRemoveUser,
        }),
      meta: {
        widthPercent: 0,
      },
    },
  ];
</script>

<BasicTable {data} {columns} />
