<script lang="ts">
  import type { V1OrganizationMemberService } from "@rilldata/web-admin/client";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import KeyIcon from "@rilldata/web-common/components/icons/KeyIcon.svelte";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import ServiceActionsCell from "./ServiceActionsCell.svelte";
  import ServiceProjectRolesCell from "./ServiceProjectRolesCell.svelte";
  import { formatServiceDate, formatOrgRole } from "./utils";

  export let data: V1OrganizationMemberService[];
  export let onSelectService: (name: string) => void;

  const columns: ColumnDef<V1OrganizationMemberService, any>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: (info) => info.getValue() as string,
    },
    {
      accessorKey: "roleName",
      header: "Organization access",
      cell: (info) => formatOrgRole(info.getValue() as string),
    },
    {
      accessorKey: "hasProjectRoles",
      header: "Project access",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(ServiceProjectRolesCell, {
          serviceName: row.original.name ?? "",
          hasProjectRoles: row.original.hasProjectRoles ?? false,
        }),
    },
    {
      accessorKey: "createdOn",
      header: "Created",
      sortDescFirst: true,
      cell: (info) => formatServiceDate(info.getValue() as string),
    },
    {
      accessorKey: "actions",
      header: "",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(ServiceActionsCell, {
          name: row.original.name ?? "",
          onManageTokens: onSelectService,
        }),
    },
  ];
</script>

<BasicTable
  {data}
  {columns}
  emptyIcon={KeyIcon}
  emptyText="No services"
  columnLayout="minmax(200px, 2fr) 1fr 1fr minmax(120px, 1fr) 56px"
/>
