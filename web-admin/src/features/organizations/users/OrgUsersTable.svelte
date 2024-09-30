<script lang="ts">
  import type {
    V1MemberUser,
    V1UserInvite,
    V1MemberUsergroup,
  } from "@rilldata/web-admin/client";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import { flexRender } from "@tanstack/svelte-table";
  import { formatDate } from "@rilldata/web-common/components/table/utils";
  import OrgUsersTableUserCompositeCell from "./OrgUsersTableUserCompositeCell.svelte";
  import OrgUsersTableActionsCell from "./OrgUsersTableActionsCell.svelte";
  import OrgUsersTableRoleCell from "./OrgUsersTableRoleCell.svelte";

  interface OrgUser extends V1MemberUser, V1UserInvite {
    invitedBy?: string;
  }

  export let data: OrgUser[];
  export let currentUserEmail: string;
  export let userGroups: V1MemberUsergroup[];
  export let onRemove: (email: string) => void;
  export let onSetRole: (email: string, role: string) => void;
  export let onAddUsergroupMemberUser: (
    email: string,
    usergroup: string,
  ) => void;

  const columns: ColumnDef<OrgUser, any>[] = [
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
        }),
    },

    {
      accessorKey: "roleName",
      header: "Role",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(OrgUsersTableRoleCell, {
          email: row.original.userEmail,
          role: row.original.roleName,
          isCurrentUser: row.original.userEmail === currentUserEmail,
          onSetRole: onSetRole,
        }),
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
          pendingAcceptance: Boolean(row.original.invitedBy),
          isCurrentUser: row.original.userEmail === currentUserEmail,
          userGroups: userGroups,
          onRemove: onRemove,
          onAddUsergroupMemberUser: onAddUsergroupMemberUser,
        }),
    },
  ];
</script>

<BasicTable {data} {columns} />
