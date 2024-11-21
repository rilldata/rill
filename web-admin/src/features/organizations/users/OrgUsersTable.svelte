<script lang="ts">
  import type { V1MemberUser, V1UserInvite } from "@rilldata/web-admin/client";
  import type { ColumnDef } from "@tanstack/svelte-table";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import { flexRender } from "@tanstack/svelte-table";
  import OrgUsersTableUserCompositeCell from "./OrgUsersTableUserCompositeCell.svelte";
  import OrgUsersTableActionsCell from "./OrgUsersTableActionsCell.svelte";
  import OrgUsersTableRoleCell from "./OrgUsersTableRoleCell.svelte";

  interface OrgUser extends V1MemberUser, V1UserInvite {
    invitedBy?: string;
  }

  export let data: OrgUser[];
  export let currentUserEmail: string;

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
          photoUrl: row.original.userPhotoUrl,
        }),
      meta: {
        widthPercent: 5,
      },
    },
    {
      accessorKey: "roleName",
      header: "Role",
      cell: ({ row }) =>
        flexRender(OrgUsersTableRoleCell, {
          email: row.original.userEmail,
          role: row.original.roleName,
          isCurrentUser: row.original.userEmail === currentUserEmail,
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
        flexRender(OrgUsersTableActionsCell, {
          email: row.original.userEmail,
          isCurrentUser: row.original.userEmail === currentUserEmail,
        }),
      meta: {
        widthPercent: 0,
      },
    },
  ];
</script>

<BasicTable {data} {columns} emptyText="No users found" scrollable />
