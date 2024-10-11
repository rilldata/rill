<script lang="ts">
  import type {
    V1MemberUser,
    V1MemberUsergroup,
  } from "@rilldata/web-admin/client";
  import { flexRender } from "@tanstack/svelte-table";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import OrgGroupsTableActionsCell from "./OrgGroupsTableActionsCell.svelte";
  import OrgGroupsTableRoleCell from "./OrgGroupsTableRoleCell.svelte";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";

  export let data: V1MemberUsergroup[];
  export let currentUserEmail: string;
  export let searchUsersList: V1MemberUser[];
  export let onDelete: (deletedGroupName: string) => void;
  export let onAddRole: (groupName: string, role: string) => void;
  export let onSetRole: (groupName: string, role: string) => void;
  export let onRevokeRole: (groupName: string) => void;

  const columns: ColumnDef<V1MemberUsergroup, any>[] = [
    {
      accessorKey: "groupName",
      header: "Group",
      meta: {
        widthPercent: 5,
      },
    },
    {
      accessorKey: "roleName",
      header: "Role",
      cell: ({ row }) =>
        flexRender(OrgGroupsTableRoleCell, {
          name: row.original.groupName,
          role: row.original.roleName,
          onAddRole: onAddRole,
          onSetRole: onSetRole,
          onRevokeRole: onRevokeRole,
        }),
      meta: {
        widthPercent: 5,
        marginLeft: "8px",
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
          searchUsersList: searchUsersList,
          onDelete: onDelete,
        }),
      meta: {
        widthPercent: 0,
      },
    },
  ];
</script>

<BasicTable {data} {columns} emptyText="No groups found" scrollable />
