<script lang="ts">
  import type { V1OrganizationMemberService } from "@rilldata/web-admin/client";
  import BasicTable from "@rilldata/web-common/components/table/BasicTable.svelte";
  import KeyIcon from "@rilldata/web-common/components/icons/KeyIcon.svelte";
  import { flexRender, type ColumnDef } from "@tanstack/svelte-table";
  import ServiceActionsCell from "./ServiceActionsCell.svelte";
  import ServiceNameCell from "./ServiceNameCell.svelte";
  import ServiceProjectRolesCell from "./ServiceProjectRolesCell.svelte";

  export let data: V1OrganizationMemberService[];
  export let onSelectService: (name: string) => void;

  function formatDate(value: string | undefined) {
    if (!value) return "-";
    return new Date(value).toLocaleDateString(undefined, {
      year: "numeric",
      month: "short",
      day: "numeric",
    });
  }

  const columns: ColumnDef<V1OrganizationMemberService, any>[] = [
    {
      accessorKey: "name",
      header: "Name",
      cell: ({ row }) =>
        flexRender(ServiceNameCell, {
          name: row.original.name ?? "",
          onClick: () => onSelectService(row.original.name ?? ""),
        }),
    },
    {
      accessorKey: "roleName",
      header: "Org Role",
    },
    {
      accessorKey: "hasProjectRoles",
      header: "Projects",
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
      cell: (info) => formatDate(info.getValue() as string),
    },
    {
      accessorKey: "actions",
      header: "",
      enableSorting: false,
      cell: ({ row }) =>
        flexRender(ServiceActionsCell, {
          name: row.original.name ?? "",
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
